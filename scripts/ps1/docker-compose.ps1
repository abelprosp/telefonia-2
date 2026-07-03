<#
.SYNOPSIS
    Script equivalente em PowerShell para docker-compose.sh
    Retorna o caminho para o arquivo docker-compose.yml baseado no ambiente.
#>

function Get-YamlPath {
    param (
        [string]$Environment = "develop"
    )

    $INFRA_ROOT_PATH = "infra/docker"
    $COMPOSE_YAML_BASENAME = "docker-compose"

    switch -Regex ($Environment.ToLower()) {
        "^(dev|develop|development|desenvolvimento)$" {
            return "$INFRA_ROOT_PATH/develop/$COMPOSE_YAML_BASENAME.yml"
        }
        "^(prod|production)$" {
            return "$INFRA_ROOT_PATH/production/$COMPOSE_YAML_BASENAME.yml"
        }
        default {
            Write-Warning "Ambiente não reconhecido: $Environment"
            return $null
        }
    }
}

# Main execution - mimics original behavior
if ($args.Count -gt 0) {
    $command = $args[0].ToLower()
    if ($command -eq "yamlpath") {
        $envArg = if ($args.Count -gt 1) { $args[1] } else { "develop" }
        $path = Get-YamlPath -Environment $envArg
        if ($path) {
            Write-Output $path
        }
    }
    else {
        Write-Host "Comando não suportado. Use: yamlpath [environment]" -ForegroundColor Red
    }
}
else {
    Write-Host "Uso: .\docker-compose.ps1 yamlpath [environment]" -ForegroundColor Yellow
}
