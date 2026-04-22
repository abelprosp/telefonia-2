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
WORKDIR /src

COPY ["nuget.config", "."]
COPY ["src/Luxus.Connect.Api/Luxus.Connect.Api.csproj", "src/Luxus.Connect.Api/"]
COPY ["src/Luxus.Connect.Infra.Http/Luxus.Connect.Infra.Http.csproj", "src/Luxus.Connect.Infra.Http/"]
COPY ["src/Luxus.Connect.Infra.Crosscutting/Luxus.Connect.Infra.Crosscutting.csproj", "src/Luxus.Connect.Infra.Crosscutting/"]
COPY ["src/Luxus.Connect.Infra.Data.Query/Luxus.Connect.Infra.Data.Query.csproj", "src/Luxus.Connect.Infra.Data.Query/"]
COPY ["src/Luxus.Connect.Application/Luxus.Connect.Application.csproj", "src/Luxus.Connect.Application/"]
COPY ["src/Luxus.Connect.Contracts/Luxus.Connect.Contracts.csproj", "src/Luxus.Connect.Contracts/"]
COPY ["src/Luxus.Connect.Domain/Luxus.Connect.Domain.csproj", "src/Luxus.Connect.Domain/"]
COPY ["src/Luxus.Connect.Infra.Data/Luxus.Connect.Infra.Data.csproj", "src/Luxus.Connect.Infra.Data/"]
COPY ["src/Luxus.Connect.Infra.IoC/Luxus.Connect.Infra.IoC.csproj", "src/Luxus.Connect.Infra.IoC/"]

RUN dotnet restore "src/Luxus.Connect.Api/Luxus.Connect.Api.csproj"

COPY . .
WORKDIR "/src/src/Luxus.Connect.Api"
RUN dotnet build "Luxus.Connect.Api.csproj" -c $BUILD_CONFIGURATION -o /app/build

# This stage is used to publish the service project to be copied to the final stage
FROM build AS publish
ARG BUILD_CONFIGURATION=Release
RUN dotnet publish "Luxus.Connect.Api.csproj" -c $BUILD_CONFIGURATION -o /app/publish /p:UseAppHost=false

# This stage is used in production or when running from VS in regular mode (Default when not using the Debug configuration)
FROM base AS final
WORKDIR /app
COPY --from=publish /app/publish .

ENTRYPOINT ["dotnet", "Luxus.Connect.Api.dll"]
