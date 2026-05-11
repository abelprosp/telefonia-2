# syntax=docker/dockerfile:1.4
# See https://aka.ms/customizecontainer to learn how to customize your debug container and how Visual Studio uses this Dockerfile to build your images for faster debugging.

# This stage is used when running from VS in fast mode (Default for Debug configuration)
FROM mcr.microsoft.com/dotnet/aspnet:10.0 AS base
USER $APP_UID
WORKDIR /app
EXPOSE 80
EXPOSE 443

# This stage is used to build the service project
FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
ARG BUILD_CONFIGURATION=Release
ARG NUGET_FEED_URL=""
ARG NUGET_PAT=""
WORKDIR /src

COPY ["docker/nuget.Docker.config", "docker/nuget.Docker.config"]
COPY ["src/Luxus.Connect.Api/Luxus.Connect.Api.csproj", "src/Luxus.Connect.Api/"]
COPY ["src/Luxus.Connect.Infra.Http/Luxus.Connect.Infra.Http.csproj", "src/Luxus.Connect.Infra.Http/"]
COPY ["src/Luxus.Connect.Infra.Crosscutting/Luxus.Connect.Infra.Crosscutting.csproj", "src/Luxus.Connect.Infra.Crosscutting/"]
COPY ["src/Luxus.Connect.Infra.Data.Query/Luxus.Connect.Infra.Data.Query.csproj", "src/Luxus.Connect.Infra.Data.Query/"]
COPY ["src/Luxus.Connect.Application/Luxus.Connect.Application.csproj", "src/Luxus.Connect.Application/"]
COPY ["src/Luxus.Connect.Contracts/Luxus.Connect.Contracts.csproj", "src/Luxus.Connect.Contracts/"]
COPY ["src/Luxus.Connect.Domain/Luxus.Connect.Domain.csproj", "src/Luxus.Connect.Domain/"]
COPY ["src/Luxus.Connect.Infra.Data/Luxus.Connect.Infra.Data.csproj", "src/Luxus.Connect.Infra.Data/"]
COPY ["src/Luxus.Connect.Infra.IoC/Luxus.Connect.Infra.IoC.csproj", "src/Luxus.Connect.Infra.IoC/"]

RUN --mount=type=secret,id=nuget_config \
  set -e; \
  if grep -qE 'packageSourceCredentials|pkgs\.dev\.azure\.com' /run/secrets/nuget_config 2>/dev/null; then \
    cp /run/secrets/nuget_config /src/nuget.config; \
  elif [ -n "$NUGET_FEED_URL" ] && [ -n "$NUGET_PAT" ]; then \
    cp docker/nuget.Docker.config /src/nuget.config; \
    dotnet nuget add source "$NUGET_FEED_URL" --name "PrivateArtifacts" --username "az" \
      --password "$NUGET_PAT" --store-password-in-clear-text --valid-authentication-types "basic"; \
  else \
    echo "Falta acesso aos pacotes privados Goal.*. Opções:"; \
    echo "  A) Guardar nuget.credentials.config (Git ignora) e NUGET_CONFIG_FILE no .env."; \
    echo "  B) Ou substituir docker/build-secrets/nuget.config pelo XML da equipa (cuidado com git commit)."; \
    echo "  C) Ou NUGET_FEED_URL + NUGET_PAT no .env (PAT: Packaging Read)."; \
    exit 1; \
  fi

RUN dotnet restore "src/Luxus.Connect.Api/Luxus.Connect.Api.csproj"

COPY . .
WORKDIR "/src/src/Luxus.Connect.Api"
RUN dotnet build "Luxus.Connect.Api.csproj" -c $BUILD_CONFIGURATION -o /app/build

# This stage is used to publish the service project to be copied to the final stage
FROM build AS publish
ARG BUILD_CONFIGURATION=Release
RUN dotnet publish "Luxus.Connect.Api.csproj" -c $BUILD_CONFIGURATION -o /app/publish /p:UseAppHost=false

# This stage is used when running from VS in regular mode (Default when not using the Debug configuration)
FROM base AS final
WORKDIR /app
COPY --from=publish /app/publish .

ENTRYPOINT ["dotnet", "Luxus.Connect.Api.dll"]
