#!/usr/bin/env bash
# Teste de fumaça da pilha app (spec home-local, D-04). Sobe banco, migrações, API e web
# com o build de produção e confere os critérios que dá para provar de fora:
#   CA-01.1 (tudo saudável, Home em /, "API online" em /status), CA-01.3 (só o web
#   exposto), CA-01.4 (modo dev continua só com o Postgres) e CA-01.5 (imagens enxutas e
#   sem root).
#
# Uso, na raiz do repositório e com o .env copiado do .env.example:
#   bash scripts/smoke-app.sh          sobe a pilha e deixa no ar
#   bash scripts/smoke-app.sh --down   derruba e apaga a pilha no fim (usado na CI)
set -euo pipefail

cd "$(dirname "$0")/.."

if [[ ! -f .env ]]; then
	echo "Falta o .env. Copie o .env.example: cp .env.example .env" >&2
	exit 2
fi

WEB_PORT=$(grep -E '^APP_WEB_PORT=' .env | cut -d= -f2)
WEB_PORT=${WEB_PORT:-3000}
BASE="http://localhost:${WEB_PORT}"
app() { docker compose --profile app "$@"; }
failures=0
ok() { echo "  ok    $*"; }
fail() {
	echo "  FALHA $*" >&2
	failures=$((failures + 1))
}

if [[ "${1:-}" == "--down" ]]; then
	trap 'app down -v --remove-orphans >/dev/null 2>&1 || true' EXIT
fi

echo "== Subindo a pilha app"
app up -d --wait --build

echo "== CA-01.1 — tudo sobe e responde"
migrate_id=$(app ps -aq migrate)
migrate_exit=$(docker inspect "$migrate_id" --format '{{.State.ExitCode}}')
[[ "$migrate_exit" == "0" ]] && ok "migrate terminou com 0" || fail "migrate terminou com $migrate_exit"
for svc in app-postgres api web; do
	health=$(docker inspect "$(app ps -q "$svc")" --format '{{.State.Health.Status}}')
	[[ "$health" == "healthy" ]] && ok "$svc saudável" || fail "$svc está $health"
done
home=$(curl -fsS "$BASE/") || fail "GET / não respondeu 200"
grep -q 'Grupos para hoje' <<<"$home" && grep -q 'data-testid="lobby-card"' <<<"$home" &&
	ok "GET / mostra a Home com os cards" || fail "GET / não mostrou a Home"
status=$(curl -fsS "$BASE/status") || fail "GET /status não respondeu 200"
grep -q 'API online' <<<"$status" && ok "/status mostra \"API online\"" || fail "/status não mostrou \"API online\""

echo "== CA-01.3 — só o web exposto"
[[ -n "$(app port web 3000 2>/dev/null)" ]] && ok "web publicado em $(app port web 3000)" || fail "web sem porta publicada"
for pair in api:8080 app-postgres:5432; do
	svc=${pair%%:*}
	port=${pair##*:}
	published=$(docker inspect "$(app ps -q "$svc")" --format "{{json (index .NetworkSettings.Ports \"${port}/tcp\")}}")
	[[ "$published" == "null" || "$published" == "[]" ]] && ok "$svc sem porta publicada" || fail "$svc publica $published"
done

echo "== CA-01.4 — modo dev continua só com o Postgres"
dev_services=$(docker compose config --services | tr '\n' ' ' | xargs)
[[ "$dev_services" == "postgres" ]] && ok "docker compose up sem perfil sobe: $dev_services" ||
	fail "sem perfil sobe: $dev_services"

echo "== CA-01.5 — imagens enxutas e sem root"
for image in ro-lobby-api:local ro-lobby-web:local; do
	user=$(docker image inspect "$image" --format '{{.Config.User}}')
	[[ -n "$user" && "$user" != "root" && "$user" != "0" ]] && ok "$image roda como $user" || fail "$image roda como root"
done
cid=$(docker create ro-lobby-api:local)
leftovers=$(docker export "$cid" | tar -t | grep -E '\.go$|usr/local/go/' || true)
docker rm "$cid" >/dev/null
[[ -z "$leftovers" ]] && ok "imagem da API sem Go e sem código-fonte" || fail "imagem da API tem: $(head -3 <<<"$leftovers")"
docker run --rm --entrypoint sh ro-lobby-web:local -c '[ ! -d src ] && [ -z "$(ls -A node_modules)" ]' &&
	ok "imagem do web sem código-fonte e sem dependências de desenvolvimento" ||
	fail "imagem do web tem src/ ou dependências de desenvolvimento"

echo
if ((failures > 0)); then
	echo "$failures verificação(ões) falharam." >&2
	exit 1
fi
echo "Pilha app ok em $BASE"
