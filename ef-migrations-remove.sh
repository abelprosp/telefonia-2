#!/bin/bash

export ASPNETCORE_ENVIRONMENT=Migrations

migrationsProject="src/Luxus.Connect.Infra.Data/Luxus.Connect.Infra.Data.csproj"
startupProject="src/Luxus.Connect.Api/Luxus.Connect.Api.csproj"
dbContext="AppDbContext"

echo "Removing last App migration..."

dotnet ef migrations remove \
    --project "$migrationsProject" \
    --startup-project "$startupProject" \
    --context "$dbContext"
