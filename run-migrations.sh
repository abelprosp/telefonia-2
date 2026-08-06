#!/bin/bash

set -euo pipefail

# ---------------------------------------------------------------------------
# run-migrations.sh — Executa as migrações SQL (001–015) no PostgreSQL.
#
# Uso:
#   DATABASE_URL="postgres://..." ./run-migrations.sh
#   ./run-migrations.sh                  # usa DATABASE_URL do ambiente
#
# Requer: psql (cliente PostgreSQL) disponível no PATH.
# ---------------------------------------------------------------------------

MIGRATIONS_DIR="$(cd "$(dirname "$0")/db/migrations" && pwd)"

MIGRATIONS=(
  "001_initial_schema.sql"
  "002_partner_line_requests.sql"
  "003_financial.sql"
  "004_sales_contracts.sql"
  "005_billing_email.sql"
  "006_invoice_layout.sql"
  "007_phone_line_customer_amount.sql"
  "008_sicredi_boleto.sql"
  "009_sicredi_payment.sql"
  "010_device_stock.sql"
  "011_customer_devices.sql"
  "012_invoice_layout_utf8.sql"
  "013_sicredi_webhook.sql"
  "014_sicredi_webhook_dedup.sql"
  "015_dual_billing_processing.sql"
)

TOTAL=${#MIGRATIONS[@]}

# ---------------------------------------------------------------------------
# Validações
# ---------------------------------------------------------------------------

if [ -z "${DATABASE_URL:-}" ]; then
  echo "erro: variável DATABASE_URL não definida." >&2
  echo "      Defina-a antes de executar o script:" >&2
  echo "      export DATABASE_URL=\"postgres://user:pass@host:5432/dbname\"" >&2
  exit 1
fi

if ! command -v psql &>/dev/null; then
  echo "erro: psql não encontrado no PATH. Instale o cliente PostgreSQL." >&2
  exit 1
fi

if [ ! -d "$MIGRATIONS_DIR" ]; then
  echo "erro: diretório de migrações não encontrado: $MIGRATIONS_DIR" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Execução
# ---------------------------------------------------------------------------

echo "========================================"
echo " Migrações SQL — connect-api"
echo " Diretório: $MIGRATIONS_DIR"
echo " Total: $TOTAL migrações"
echo "========================================"
echo ""

FAILED=0

for i in "${!MIGRATIONS[@]}"; do
  FILE="${MIGRATIONS[$i]}"
  STEP=$((i + 1))
  FILEPATH="$MIGRATIONS_DIR/$FILE"

  printf "[%02d/%02d] %s ... " "$STEP" "$TOTAL" "$FILE"

  if [ ! -f "$FILEPATH" ]; then
    echo "ERRO (arquivo não encontrado)"
    echo "       Caminho esperado: $FILEPATH" >&2
    FAILED=$((FAILED + 1))
    continue
  fi

  if psql "$DATABASE_URL" \
       --single-transaction \
       --set ON_ERROR_STOP=1 \
       -f "$FILEPATH" \
       --quiet \
       2>&1; then
    echo "OK"
  else
    EXIT_CODE=$?
    echo "FALHOU (código $EXIT_CODE)"
    echo "" >&2
    echo "  Migração com falha: $FILE" >&2
    echo "  Abortando — nenhuma migração posterior será executada." >&2
    echo "" >&2
    FAILED=$((FAILED + 1))
    break
  fi
done

echo ""
echo "========================================"

SUCCEEDED=$((STEP - FAILED))

if [ "$FAILED" -eq 0 ]; then
  echo " Concluído: $TOTAL/$TOTAL migrações aplicadas com sucesso."
  echo "========================================"
  exit 0
else
  DONE=$((STEP - 1))
  echo " Falha: $DONE/$TOTAL migrações executadas antes do erro."
  echo "========================================"
  exit 1
fi
