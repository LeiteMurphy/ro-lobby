#!/usr/bin/env bash
# Backup do banco da pilha app (spec endurecimento, RN-18). Faz o pg_dump do container
# app-postgres em ~/ro-lobby-backups/ e mantém só os 10 backups mais recentes. A saída mostra
# só o nome do arquivo e a contagem, nunca dados dos usuários.
#
# Uso, na raiz do repositório e com a pilha app no ar:
#   bash scripts/backup.sh               ro_lobby-<data>.sql
#   bash scripts/backup.sh antes-00012   ro_lobby-antes-00012-<data>.sql (antes de migrar)
#
# Variáveis para os testes: BACKUP_DIR (pasta), BACKUP_KEEP (quantos manter) e
# BACKUP_DUMP_CMD (comando que escreve o dump na saída padrão).
set -euo pipefail

cd "$(dirname "$0")/.."

dir=${BACKUP_DIR:-"$HOME/ro-lobby-backups"}
keep=${BACKUP_KEEP:-10}
label=${1:-}
if [[ -n "$label" && ! "$label" =~ ^[A-Za-z0-9_-]+$ ]]; then
	echo "Use só letras, números, - e _ no rótulo." >&2
	exit 2
fi

env_value() { sed -n "s/^$1=//p" .env 2>/dev/null | tail -n 1 | tr -d '\r'; }
db_user=$(env_value POSTGRES_USER)
db_name=$(env_value POSTGRES_DB)
# docker exec direto no container (projeto ro-lobby): não depende das variáveis do Compose,
# então funciona mesmo no meio de uma troca de senha (RN-19).
dump_cmd=${BACKUP_DUMP_CMD:-"docker exec ro-lobby-app-postgres-1 pg_dump -U ${db_user:-ro_lobby} -d ${db_name:-ro_lobby}"}

mkdir -p "$dir"
chmod 700 "$dir" 2>/dev/null || true
name="ro_lobby-${label:+$label-}$(date +%Y%m%d-%H%M%S).sql"
file="$dir/$name"
tmp="$file.parcial"
trap 'rm -f "$tmp"' EXIT

# O dump vai direto para o arquivo, sem passar pela tela.
if ! bash -c "$dump_cmd" >"$tmp" 2>/dev/null || [[ ! -s "$tmp" ]]; then
	echo "O backup falhou: o pg_dump não gerou o arquivo. Confira se a pilha app está no ar." >&2
	exit 1
fi
mv "$tmp" "$file"
chmod 600 "$file" 2>/dev/null || true
echo "Backup criado: $name"

# RN-18: só os $keep mais recentes ficam.
mapfile -t old < <(ls -1t "$dir"/ro_lobby-*.sql 2>/dev/null | tail -n +"$((keep + 1))")
for f in "${old[@]}"; do
	rm -f -- "$f"
done
total=$(ls -1 "$dir"/ro_lobby-*.sql 2>/dev/null | wc -l | tr -d ' ')
echo "Backups guardados: $total (${#old[@]} antigo(s) apagado(s))"
