#!/bin/bash

export ASPNETCORE_ENVIRONMENT=Migrations

migrationsProject="src/Luxus.Connect.Infra.Data/Luxus.Connect.Infra.Data.csproj"
startupProject="src/Luxus.Connect.Api/Luxus.Connect.Api.csproj"
dbContext="AppDbContext"

if [ "$#" -lt 1 ]; then
    echo "Usage: $0 <migration-name>"
    echo "Example: $0 AddCustomerTable"
    exit 1
fi

name=$1

echo "Adding migration '$name' to App"

dotnet ef migrations add "$name" \
    --project "$migrationsProject" \
    --startup-project "$startupProject" \
    --context "$dbContext" \
    --output-dir Migrations
