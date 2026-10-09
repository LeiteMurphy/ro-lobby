#!/usr/bin/env bash
# Confere a configuração do Compose da publicação por túnel (spec publicacao-tunel), sem
# subir nada: CA-01.1 (origem configurável), CA-01.2 (serviço tunnel), CA-01.3 (token
# obrigatório) e a parte de configuração do CA-02.1 (modo local igual). O resto do CA-02.1
# fica com o teste de fumaça da pilha.
#
# Usa uma cópia do .env.example, para não depender do .env de quem roda.
#
# Uso, na raiz do repositório:
#   bash scripts/check-compose-tunnel.sh
set -euo pipefail

cd "$(dirname "$0")/.."

env_file=$(mktemp)
trap 'rm -f "$env_file"' EXIT
cp .env.example "$env_file"
unset APP_ORIGIN CLOUDFLARE_TUNNEL_TOKEN APP_WEB_PORT

failures=0
ok() { echo "  ok    $*"; }
fail() {
	echo "  FALHA $*" >&2
	failures=$((failures + 1))
}

base() { docker compose --env-file "$env_file" --profile app "$@"; }
tunnel() { docker compose --env-file "$env_file" -f docker-compose.yml -f docker-compose.tunnel.yml --profile app "$@"; }
# Lê um caminho (ex.: services.web.environment.ORIGIN) do JSON do `config` na entrada.
get() { node -e 'let s="";process.stdin.on("data",d=>s+=d).on("end",()=>{let v=JSON.parse(s);for(const k of process.argv[1].split("."))v=v?.[k];console.log(typeof v==="object"?JSON.stringify(v):v??"")})' "$1"; }

echo "== CA-02.1 — modo local igual"
local_json=$(base config --format json)
origin=$(get services.web.environment.ORIGIN <<<"$local_json")
[[ "$origin" == "http://localhost:3000" ]] && ok "ORIGIN padrão é $origin" || fail "ORIGIN padrão é '$origin'"
[[ -z "$(get services.tunnel <<<"$local_json")" ]] && ok "sem serviço tunnel" || fail "o serviço tunnel aparece sem o arquivo do túnel"

echo "== CA-01.1 — origem configurável"
origin=$(APP_ORIGIN=https://rolobby.com.br base config --format json | get services.web.environment.ORIGIN)
[[ "$origin" == "https://rolobby.com.br" ]] && ok "ORIGIN vem de APP_ORIGIN" || fail "ORIGIN é '$origin'"

echo "== CA-01.2 — serviço tunnel"
tunnel_json=$(CLOUDFLARE_TUNNEL_TOKEN=token-de-teste tunnel config --format json)
image=$(get services.tunnel.image <<<"$tunnel_json")
[[ "$image" =~ ^cloudflare/cloudflared:[0-9]+\.[0-9]+\.[0-9]+$ ]] && ok "imagem fixa $image" || fail "imagem '$image' sem versão fixa"
[[ "$(get services.tunnel.environment.TUNNEL_TOKEN <<<"$tunnel_json")" == "token-de-teste" ]] && ok "token no ambiente do tunnel" || fail "token não chega ao tunnel"
[[ "$(get services.tunnel.depends_on.web.condition <<<"$tunnel_json")" == "service_healthy" ]] && ok "depende do web saudável" || fail "não depende do web saudável"
[[ "$(get services.tunnel.restart <<<"$tunnel_json")" == "unless-stopped" ]] && ok "restart unless-stopped" || fail "restart diferente de unless-stopped"
[[ -z "$(get services.tunnel.ports <<<"$tunnel_json")" ]] && ok "tunnel sem porta publicada" || fail "tunnel publica porta"
published=$(node -e 'let s="";process.stdin.on("data",d=>s+=d).on("end",()=>{const sv=JSON.parse(s).services;console.log(Object.keys(sv).filter(k=>(sv[k].ports??[]).length).join(","))})' <<<"$tunnel_json")
[[ "$published" == "web" ]] && ok "só o web publica porta" || fail "publicam porta: '$published'"
for svc in api web migrate app-postgres; do
	[[ "$(get "services.$svc.environment.TUNNEL_TOKEN" <<<"$tunnel_json")" == "" ]] || fail "o token vaza para $svc"
done
ok "token só no tunnel"

echo "== CA-01.3 — token obrigatório"
if out=$(tunnel config 2>&1); then
	fail "o Compose aceitou o arquivo do túnel sem token"
elif grep -q "defina CLOUDFLARE_TUNNEL_TOKEN no .env" <<<"$out"; then
	ok "falha com a mensagem do token"
else
	fail "falhou sem a mensagem do token: $out"
fi

if ((failures > 0)); then
	echo "$failures falha(s)" >&2
	exit 1
fi
echo "Tudo certo."
