#!/bin/bash

export ASPNETCORE_ENVIRONMENT=Migrations

migrationsProject="src/Luxus.Connect.Infra.Data/Luxus.Connect.Infra.Data.csproj"
startupProject="src/Luxus.Connect.Api/Luxus.Connect.Api.csproj"
dbContext="AppDbContext"

from=""
to=""

usage() {
    cat <<EOF
Usage:
  $0 [--from <migration>] [--to <migration>]
  $0 [<to-migration>]
  $0 <from-migration> <to-migration>

  --from    Optional. Source migration name (documented; EF applies up to --to).
  --to      Target migration name. Omit for all pending migrations.
  Positional: one argument = --to; two arguments = --from and --to (in order).

Examples:
  $0
  $0 20260320013352_03
  $0 --to 20260320013352_03
  $0 --from 20260318014205_01 --to 20260321202553_RemoveCustomerSellerId
EOF
}

while [ $# -gt 0 ]; do
    case "$1" in
        --from)
            if [ -z "${2:-}" ]; then echo "error: --from requires a migration name" >&2; exit 1; fi
            from="$2"
            shift 2
            ;;
        --to)
            if [ -z "${2:-}" ]; then echo "error: --to requires a migration name" >&2; exit 1; fi
            to="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        --)
            shift
            break
            ;;
        -*)
            echo "error: unknown option: $1" >&2
            usage >&2
            exit 1
            ;;
        *)
            break
            ;;
    esac
done

# Remaining positional: 0, 1 (= to), or 2 (= from, to)
case $# in
    0) ;;
    1)
        if [ -n "$to" ]; then
            echo "error: duplicate target migration (--to and positional)" >&2
            exit 1
        fi
        to="$1"
        ;;
    2)
        if [ -n "$from" ] || [ -n "$to" ]; then
            echo "error: do not mix --from/--to with two positional arguments" >&2
            exit 1
        fi
        from="$1"
        to="$2"
        ;;
    *)
        echo "error: too many positional arguments (expected 0–2)" >&2
        usage >&2
        exit 1
        ;;
esac

if [ -n "$from" ] && [ -z "$to" ]; then
    echo "error: --from requires --to (or a second positional migration name)" >&2
    exit 1
fi

echo "Updating App database..."
if [ -n "$from" ] && [ -n "$to" ]; then
    echo "  from: $from"
    echo "  to:   $to"
elif [ -n "$to" ]; then
    echo "  to:   $to"
fi

if [ -n "$to" ]; then
    dotnet ef database update "$to" \
        --project "$migrationsProject" \
        --startup-project "$startupProject" \
        --context "$dbContext"
else
    dotnet ef database update \
        --project "$migrationsProject" \
        --startup-project "$startupProject" \
        --context "$dbContext"
fi
