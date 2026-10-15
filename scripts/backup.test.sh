#!/usr/bin/env bash
# Testes do backup (spec endurecimento, CA-07.1). Usa uma pasta temporária e um pg_dump de
# mentira, então não precisa do Docker nem da pilha.
#
# Uso, na raiz do repositório:
#   bash scripts/backup.test.sh
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

secret_row="ana-discord-123|Lirien|segredo-do-usuario"
export BACKUP_DIR="$dir" BACKUP_DUMP_CMD="echo 'COPY users FROM stdin;'; echo '$secret_row'"

echo "== CA-07.1 — com 10 backups, cria um e apaga o mais antigo"
for i in $(seq -w 1 10); do
	echo "antigo $i" >"$dir/ro_lobby-202610${i}-000000.sql"
	touch -d "2026-10-$i 00:00:00" "$dir/ro_lobby-202610${i}-000000.sql"
done
out=$(bash scripts/backup.sh 2>&1)
count=$(ls -1 "$dir"/ro_lobby-*.sql | wc -l | tr -d ' ')
[[ $count == 10 ]] && ok "sobram 10 backups" || fail "sobraram $count"
[[ ! -e "$dir/ro_lobby-20261001-000000.sql" ]] && ok "o mais antigo foi apagado" || fail "o mais antigo continua"
[[ -e "$dir/ro_lobby-20261002-000000.sql" ]] && ok "o segundo mais antigo continua" || fail "apagou além do mais antigo"
new=$(ls -1t "$dir"/ro_lobby-*.sql | head -n 1)
grep -qF "$secret_row" "$new" && ok "o arquivo novo tem o dump" || fail "o arquivo novo não tem o dump"
grep -qF "$secret_row" <<<"$out" && fail "a saída mostrou dados de usuário" || ok "a saída não mostra dados de usuário"
grep -q "Backup criado: ro_lobby-" <<<"$out" && ok "a saída mostra o nome do arquivo" || fail "saída: $out"
grep -q "Backups guardados: 10 (1 antigo(s) apagado(s))" <<<"$out" && ok "a saída mostra a contagem" || fail "saída: $out"

echo "== RN-18 — rótulo no nome"
out=$(bash scripts/backup.sh antes-00012 2>&1)
grep -q "Backup criado: ro_lobby-antes-00012-" <<<"$out" && ok "o rótulo entra no nome" || fail "saída: $out"
if bash scripts/backup.sh 'x;rm -rf /' >/dev/null 2>&1; then fail "aceitou rótulo com caracteres especiais"; else ok "recusa rótulo com caracteres especiais"; fi

echo "== RN-18 — dump que falha não apaga nada nem deixa arquivo pela metade"
before=$(ls -1 "$dir" | sort)
if BACKUP_DUMP_CMD="exit 1" bash scripts/backup.sh >/dev/null 2>&1; then
	fail "o backup com dump falho terminou com sucesso"
else
	ok "o backup com dump falho termina com erro"
fi
[[ "$(ls -1 "$dir" | sort)" == "$before" ]] && ok "a pasta ficou igual" || fail "a pasta mudou: $(ls -1 "$dir")"

if ((failures > 0)); then
	echo "$failures falha(s)" >&2
	exit 1
fi
echo "Tudo certo."
