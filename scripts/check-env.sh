#!/bin/sh
# Verificador do .env (spec endurecimento, RN-02, RN-04, RN-05). Aponta variável obrigatória
# faltando, valor de exemplo em uso com origem pública, senha de produção curta e bot pela
# metade. Nunca imprime o valor de nenhuma variável: só o nome e o problema.
#
# Roda em dois lugares, por isso é sh puro (sem bash):
#   no host, lendo o .env:          sh scripts/check-env.sh
#   no serviço env-check da pilha:  sh check-env.sh --from-environment --example <arquivo>
#
# Opções:
#   --env-file <arquivo>   .env a conferir (padrão: .env na raiz do repositório)
#   --example <arquivo>    .env.example com os valores de exemplo (padrão: o da raiz)
#   --from-environment     lê as variáveis do ambiente em vez de um arquivo; não confere o
#                          token do túnel, que não chega ao serviço env-check
#
# Sai com 0 quando está tudo certo e com 1 quando há problema.
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
env_file="$root/.env"
example="$root/.env.example"
from_env=0
while [ $# -gt 0 ]; do
	case "$1" in
	--env-file) env_file=$2; shift 2 ;;
	--example) example=$2; shift 2 ;;
	--from-environment) from_env=1; shift ;;
	*) echo "opção desconhecida: $1" >&2; exit 2 ;;
	esac
done

# Lê KEY=VALUE de um arquivo sem executá-lo (o .env não é um script). Aspas em volta do
# valor são removidas, como faz o Docker Compose.
file_value() {
	[ -f "$2" ] || return 0
	sed -n "s/^[[:space:]]*$1=//p" "$2" | tail -n 1 | sed -e 's/^"\(.*\)"$/\1/' -e "s/^'\(.*\)'\$/\1/" | tr -d '\r'
}

value() {
	if [ "$from_env" = 1 ]; then
		eval "printf '%s' \"\${$1:-}\""
	else
		file_value "$1" "$env_file"
	fi
}

if [ "$from_env" = 0 ] && [ ! -f "$env_file" ]; then
	echo "Falta o .env. Copie o .env.example: cp .env.example .env" >&2
	exit 1
fi

problems=0
problem() {
	echo "  - $1: $2"
	problems=$((problems + 1))
}

origin=$(value APP_ORIGIN)
public=0
case "$origin" in https://*) public=1 ;; esac

# RN-01 e RN-04: obrigatórias em qualquer modo.
for key in POSTGRES_USER POSTGRES_DB APP_POSTGRES_PASSWORD DISCORD_CLIENT_ID DISCORD_CLIENT_SECRET; do
	[ -n "$(value "$key")" ] || problem "$key" "falta no .env"
done

# A senha vai dentro da DATABASE_URL da pilha: caracteres especiais quebrariam a URL.
app_password=$(value APP_POSTGRES_PASSWORD)
if [ -n "$app_password" ] && printf '%s' "$app_password" | grep -q '[^A-Za-z0-9_-]'; then
	problem APP_POSTGRES_PASSWORD "use só letras, números, - e _ (a senha vai na URL do banco)"
fi

# RN-02: com origem pública, nada de exemplo nem senha curta.
if [ "$public" = 1 ]; then
	if [ -n "$app_password" ] && [ "${#app_password}" -lt 24 ]; then
		problem APP_POSTGRES_PASSWORD "tem menos de 24 caracteres"
	fi
	for key in APP_POSTGRES_PASSWORD DISCORD_CLIENT_SECRET DISCORD_CLIENT_ID; do
		current=$(value "$key")
		[ -n "$current" ] || continue
		if [ "$current" = "$(file_value "$key" "$example")" ]; then
			problem "$key" "está com o valor do .env.example"
		fi
	done
	if [ "$from_env" = 0 ] && [ -z "$(value CLOUDFLARE_TUNNEL_TOKEN)" ]; then
		problem CLOUDFLARE_TUNNEL_TOKEN "falta no .env (a origem é pública)"
	fi
fi

# RN-22 da discord-bot: o bot liga com as duas variáveis ou desliga com as duas vazias.
bot_token=$(value DISCORD_BOT_TOKEN)
guild_id=$(value DISCORD_GUILD_ID)
if [ -n "$bot_token" ] && [ -z "$guild_id" ]; then
	problem DISCORD_GUILD_ID "falta, porque o DISCORD_BOT_TOKEN está preenchido"
elif [ -z "$bot_token" ] && [ -n "$guild_id" ]; then
	problem DISCORD_BOT_TOKEN "falta, porque o DISCORD_GUILD_ID está preenchido"
fi

if [ "$problems" -gt 0 ]; then
	echo "O .env tem $problems problema(s). Nenhum valor foi mostrado." >&2
	exit 1
fi
if [ "$public" = 1 ]; then
	echo "O .env está pronto para publicar (origem pública)."
else
	echo "O .env está pronto para uso local."
fi
