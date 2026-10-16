#!/usr/bin/env bash
# Confere a configuração do Compose da publicação por túnel (spec publicacao-tunel), sem
# subir nada: CA-01.1 (origem configurável), CA-01.2 (serviço tunnel), CA-01.3 (token
# obrigatório) e a parte de configuração do CA-02.1 (modo local igual). O resto do CA-02.1
# fica com o teste de fumaça da pilha.
#
# Usa uma cópia do .env.example, para não depender do .env de quem roda. Precisa do Docker
# e do node no PATH (o node lê o JSON do `docker compose config`).
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
leaks=$(node -e 'let s="";process.stdin.on("data",d=>s+=d).on("end",()=>{const sv=JSON.parse(s).services;console.log(Object.keys(sv).filter(k=>k!=="tunnel"&&JSON.stringify(sv[k]).includes("token-de-teste")).join(","))})' <<<"$tunnel_json")
[[ -z "$leaks" ]] && ok "token só no tunnel" || fail "o token vaza para: $leaks"

echo "== CA-01.3 — token obrigatório"
if out=$(tunnel config 2>&1); then
	fail "o Compose aceitou o arquivo do túnel sem token"
elif grep -q "defina CLOUDFLARE_TUNNEL_TOKEN no .env" <<<"$out"; then
	ok "falha com a mensagem do token"
else
	fail "falhou sem a mensagem do token: $out"
fi

echo "== Endurecimento CA-01.1 / CA-01.4 — env-check antes do banco"
app_json=$(base config --format json)
[[ "$(get services.app-postgres.depends_on.env-check.condition <<<"$app_json")" == "service_completed_successfully" ]] && ok "app-postgres espera o env-check" || fail "app-postgres não depende do env-check"
[[ "$(get services.migrate.depends_on.app-postgres.condition <<<"$app_json")" == "service_healthy" ]] && ok "migrate espera o banco" || fail "migrate não depende do banco"
app_pw=$(sed -n 's/^APP_POSTGRES_PASSWORD=//p' .env.example)
dev_pw=$(sed -n 's/^POSTGRES_PASSWORD=//p' .env.example)
[[ "$(get services.app-postgres.environment.POSTGRES_PASSWORD <<<"$app_json")" == "$app_pw" ]] && ok "app-postgres usa APP_POSTGRES_PASSWORD" || fail "app-postgres não usa APP_POSTGRES_PASSWORD"
grep -qF ":$app_pw@app-postgres" <<<"$(get services.api.environment.DATABASE_URL <<<"$app_json")" && ok "a API conecta com APP_POSTGRES_PASSWORD" || fail "a DATABASE_URL da API não usa APP_POSTGRES_PASSWORD"
grep -qF ":$dev_pw@" <<<"$app_json" && fail "a senha de dev aparece na pilha app" || ok "a senha de dev não entra na pilha app"
[[ "$(get services.env-check.network_mode <<<"$app_json")" == "none" ]] && ok "env-check sem rede" || fail "env-check com rede"
if base run --rm --no-deps env-check >/dev/null 2>&1; then ok "env-check aceita o .env.example local"; else fail "env-check recusou o .env.example local"; fi
if out=$(APP_ORIGIN=https://rolobby.com.br base run --rm --no-deps env-check 2>&1); then
	fail "env-check aceitou a senha de exemplo com origem pública"
elif grep -q "APP_POSTGRES_PASSWORD: está com o valor do .env.example" <<<"$out" && ! grep -qF "$app_pw" <<<"$out"; then
	ok "env-check recusa a senha de exemplo com origem pública, sem mostrar o valor"
else
	fail "env-check falhou sem a mensagem esperada"
fi

echo "== Endurecimento CA-02.1 / CA-02.2 — portas só em 127.0.0.1"
dev_json=$(docker compose --env-file "$env_file" --profile dev config --format json)
[[ "$(get services.postgres.ports.0.host_ip <<<"$dev_json")" == "127.0.0.1" ]] && ok "postgres de dev em 127.0.0.1" || fail "postgres de dev fora de 127.0.0.1"
[[ "$(get services.web.ports.0.host_ip <<<"$tunnel_json")" == "127.0.0.1" ]] && ok "web em 127.0.0.1" || fail "web fora de 127.0.0.1"
# Serviços com porta publicada fora de 127.0.0.1, no JSON do `config` na entrada.
exposed() { node -e 'let s="";process.stdin.on("data",d=>s+=d).on("end",()=>{const sv=JSON.parse(s).services;console.log(Object.entries(sv).filter(([,v])=>(v.ports??[]).some(p=>p.host_ip!=="127.0.0.1")).map(([k])=>k).join(","))})'; }
open_ports="$(exposed <<<"$tunnel_json")$(exposed <<<"$dev_json")"
[[ -z "$open_ports" ]] && ok "nenhuma porta fora de 127.0.0.1" || fail "portas abertas para a rede em: $open_ports"
for svc in app-postgres api; do
	[[ -z "$(get services.$svc.ports <<<"$tunnel_json")" ]] && ok "$svc sem porta publicada" || fail "$svc publica porta"
done

if ((failures > 0)); then
	echo "$failures falha(s)" >&2
	exit 1
fi
echo "Tudo certo."
