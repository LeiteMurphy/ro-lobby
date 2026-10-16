#!/usr/bin/env bash
# Testes do verificador do .env (spec endurecimento): CA-01.1 a CA-01.4, CA-08.1 e CA-08.2.
# Monta .env de mentira a partir do .env.example e confere a saída do check-env.sh, nos dois
# modos (arquivo e ambiente). Não precisa do Docker.
#
# Uso, na raiz do repositório:
#   bash scripts/check-env.test.sh
set -euo pipefail

cd "$(dirname "$0")/.."

dir=$(mktemp -d)
trap 'rm -rf "$dir"' EXIT

failures=0
ok() { echo "  ok    $*"; }
fail() {
	echo "  FALHA $*" >&2
	failures=$((failures + 1))
}

STRONG=Abcdefghijklmnopqrstuvwx0123456789
SECRET=segredo-de-verdade-do-discord-xyz
TUNNEL=token-de-verdade-do-tunel-xyz

# .env a partir do exemplo, trocando ou acrescentando chaves: env_with arq CHAVE=valor ...
env_with() {
	local file=$1
	shift
	cp .env.example "$file"
	local pair key
	for pair in "$@"; do
		key=${pair%%=*}
		sed -i "/^#\{0,1\} *$key=/d" "$file"
		echo "$pair" >>"$file"
	done
}

# Roda o verificador e guarda a saída e o código: run_check arq [opções]
run_check() {
	local file=$1
	shift
	set +e
	out=$(sh scripts/check-env.sh --env-file "$file" "$@" 2>&1)
	code=$?
	set -e
}

no_values() {
	local value
	for value in "$@"; do
		if grep -qF -- "$value" <<<"$out"; then
			fail "a saída mostrou um valor do .env"
			return
		fi
	done
	ok "nenhum valor aparece na saída"
}

echo "== CA-01.4 — o .env.example vale para uso local"
run_check .env.example
[[ $code == 0 ]] && ok "código 0" || fail "código $code: $out"
grep -q "pronto para uso local" <<<"$out" && ok "diz que está pronto" || fail "saída: $out"

echo "== CA-08.1 — .env bom com origem pública passa"
env_with "$dir/bom" APP_ORIGIN=https://rolobby.com.br "APP_POSTGRES_PASSWORD=$STRONG" \
	"DISCORD_CLIENT_SECRET=$SECRET" DISCORD_CLIENT_ID=123456789012345678 \
	"CLOUDFLARE_TUNNEL_TOKEN=$TUNNEL" DISCORD_BOT_TOKEN=bot-de-verdade DISCORD_GUILD_ID=1560117933705797643
run_check "$dir/bom"
[[ $code == 0 ]] && ok "código 0" || fail "código $code: $out"
grep -q "pronto para publicar" <<<"$out" && ok "diz que está pronto para publicar" || fail "saída: $out"
no_values "$STRONG" "$SECRET" "$TUNNEL" rolobby.com.br

echo "== CA-01.1 — senha de exemplo com origem pública"
example_pw=$(sed -n 's/^APP_POSTGRES_PASSWORD=//p' .env.example)
env_with "$dir/exemplo" APP_ORIGIN=https://rolobby.com.br "DISCORD_CLIENT_SECRET=$SECRET" \
	DISCORD_CLIENT_ID=123456789012345678 "CLOUDFLARE_TUNNEL_TOKEN=$TUNNEL"
run_check "$dir/exemplo"
[[ $code != 0 ]] && ok "código diferente de zero" || fail "aceitou a senha de exemplo"
grep -q "APP_POSTGRES_PASSWORD: está com o valor do .env.example" <<<"$out" && ok "cita APP_POSTGRES_PASSWORD" || fail "saída: $out"
no_values "$example_pw" "$SECRET"

echo "== CA-01.2 — senha com 23 caracteres com origem pública"
env_with "$dir/curta" APP_ORIGIN=https://rolobby.com.br APP_POSTGRES_PASSWORD=Abcdefghijklmnopqrstu12 \
	"DISCORD_CLIENT_SECRET=$SECRET" DISCORD_CLIENT_ID=123456789012345678 "CLOUDFLARE_TUNNEL_TOKEN=$TUNNEL"
run_check "$dir/curta"
[[ $code != 0 ]] && ok "código diferente de zero" || fail "aceitou a senha curta"
grep -q "APP_POSTGRES_PASSWORD: tem menos de 24 caracteres" <<<"$out" && ok "cita o mínimo de 24" || fail "saída: $out"
env_with "$dir/24" APP_ORIGIN=https://rolobby.com.br APP_POSTGRES_PASSWORD=Abcdefghijklmnopqrstu123 \
	"DISCORD_CLIENT_SECRET=$SECRET" DISCORD_CLIENT_ID=123456789012345678 "CLOUDFLARE_TUNNEL_TOKEN=$TUNNEL"
run_check "$dir/24"
[[ $code == 0 ]] && ok "24 caracteres passam" || fail "recusou 24 caracteres: $out"

echo "== CA-01.3 — Client Secret de exemplo com origem pública"
env_with "$dir/secret" APP_ORIGIN=https://rolobby.com.br "APP_POSTGRES_PASSWORD=$STRONG" \
	DISCORD_CLIENT_ID=123456789012345678 "CLOUDFLARE_TUNNEL_TOKEN=$TUNNEL"
run_check "$dir/secret"
[[ $code != 0 ]] && ok "código diferente de zero" || fail "aceitou o Client Secret de exemplo"
grep -q "DISCORD_CLIENT_SECRET: está com o valor do .env.example" <<<"$out" && ok "cita DISCORD_CLIENT_SECRET" || fail "saída: $out"

echo "== CA-08.2 — três problemas, sem vazar valor"
env_with "$dir/tres" APP_ORIGIN=https://rolobby.com.br "DISCORD_CLIENT_SECRET=$SECRET" \
	DISCORD_CLIENT_ID=123456789012345678 "CLOUDFLARE_TUNNEL_TOKEN=$TUNNEL" DISCORD_BOT_TOKEN=bot-de-verdade-xyz
run_check "$dir/tres"
[[ $code != 0 ]] && ok "código diferente de zero" || fail "aceitou o .env com problemas"
grep -q "APP_POSTGRES_PASSWORD: está com o valor do .env.example" <<<"$out" && ok "aponta a senha de exemplo" || fail "saída: $out"
grep -q "APP_POSTGRES_PASSWORD: tem menos de 24 caracteres" <<<"$out" && ok "aponta a senha curta" || fail "saída: $out"
grep -q "DISCORD_GUILD_ID: falta" <<<"$out" && ok "aponta o bot pela metade" || fail "saída: $out"
grep -q "3 problema(s)" <<<"$out" && ok "conta 3 problemas" || fail "saída: $out"
no_values "$example_pw" "$SECRET" "$TUNNEL" bot-de-verdade-xyz

echo "== RN-01 — APP_POSTGRES_PASSWORD obrigatória"
env_with "$dir/sem" APP_POSTGRES_PASSWORD=
run_check "$dir/sem"
[[ $code != 0 ]] && grep -q "APP_POSTGRES_PASSWORD: falta no .env" <<<"$out" && ok "aponta a falta" || fail "código $code: $out"

echo "== RN-02 — senha com caractere que quebra a URL do banco"
env_with "$dir/url" 'APP_POSTGRES_PASSWORD=senha@com/barra'
run_check "$dir/url"
[[ $code != 0 ]] && grep -q "use só letras, números" <<<"$out" && ok "aponta o caractere" || fail "código $code: $out"

echo "== RN-02 — token do túnel com origem pública (só no host)"
env_with "$dir/tunel" APP_ORIGIN=https://rolobby.com.br "APP_POSTGRES_PASSWORD=$STRONG" \
	"DISCORD_CLIENT_SECRET=$SECRET" DISCORD_CLIENT_ID=123456789012345678
run_check "$dir/tunel"
[[ $code != 0 ]] && grep -q "CLOUDFLARE_TUNNEL_TOKEN: falta" <<<"$out" && ok "aponta o token do túnel" || fail "código $code: $out"

echo "== Modo ambiente (serviço env-check)"
set +e
out=$(env -i PATH="$PATH" APP_ORIGIN=https://rolobby.com.br POSTGRES_USER=ro_lobby POSTGRES_DB=ro_lobby \
	"APP_POSTGRES_PASSWORD=$example_pw" "DISCORD_CLIENT_SECRET=$SECRET" DISCORD_CLIENT_ID=123456789012345678 \
	sh scripts/check-env.sh --from-environment --example .env.example 2>&1)
code=$?
set -e
[[ $code != 0 ]] && grep -q "APP_POSTGRES_PASSWORD: está com o valor do .env.example" <<<"$out" && ok "recusa a senha de exemplo" || fail "código $code: $out"
grep -q CLOUDFLARE_TUNNEL_TOKEN <<<"$out" && fail "cobrou o token do túnel no modo ambiente" || ok "não cobra o token do túnel"
no_values "$example_pw" "$SECRET"

if ((failures > 0)); then
	echo "$failures falha(s)" >&2
	exit 1
fi
echo "Tudo certo."
