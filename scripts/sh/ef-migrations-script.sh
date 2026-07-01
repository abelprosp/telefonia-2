#!/bin/bash

export ASPNETCORE_ENVIRONMENT=Migrations

migrationsProject="src/Luxus.Connect.Infra.Data/Luxus.Connect.Infra.Data.csproj"
startupProject="src/Luxus.Connect.Api/Luxus.Connect.Api.csproj"
dbContext="AppDbContext"

from=${1:-}
to=${2:-}

echo "Generating App migrations script..."

args=(
    --project "$migrationsProject"
    --startup-project "$startupProject"
    --context "$dbContext"
)

[ -n "$from" ] && args+=(--from "$from")
[ -n "$to" ] && args+=(--to "$to")

dotnet ef migrations script "${args[@]}"
