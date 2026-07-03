<#
.SYNOPSIS
    Script equivalente em PowerShell para docker-analysis.sh
    Realiza análise de imagens Docker usando Dive.
#>

$DEFAULT_DIVE_VERSION = "latest"
$DEFAULT_DIVE_IMAGE = "docker.io/wagoodman/dive"
$DEFAULT_DIVE_UI_CONFIG_PATH = "$(Get-Location)/.dive-ui.yaml"
$DEFAULT_DIVE_CI_CONFIG_PATH = "$(Get-Location)/.dive-ci.yaml"
$DEFAULT_DOCKER_SOCK_PATH = "npipe:////./pipe/docker_engine"

function Show-Usage {
    param (
        [string]$Mode
    )

    switch ($Mode) {
        "ui" {
            $ConfigPath = $DEFAULT_DIVE_UI_CONFIG_PATH
            $WhatItDoes = "Realiza análise da imagem Docker, mostrando ao final a UI."
        }
        "ci" {
            $ConfigPath = $DEFAULT_DIVE_CI_CONFIG_PATH
            $WhatItDoes = "Realiza análise da imagem Docker, sem mostrar a UI, i.e. modo pass/fail."
        }
        default {
            Write-Host "Modo de execução não suportado. Valores suportados: ui | ci" -ForegroundColor Red
            exit 1
        }
    }

    Write-Host "----------------------------------------------------------------------------------------------------------------------------------" -ForegroundColor Cyan
    Write-Host "${Mode.ToUpper()}: $WhatItDoes" -ForegroundColor White
    Write-Host "----------------------------------------------------------------------------------------------------------------------------------" -ForegroundColor Cyan
    Write-Host "Comando: ./docker-analysis.ps1 $Mode [target-image] [config-file-path] [dive-version]"
    Write-Host "----------------------------------------------------------------------------------------------------------------------------------"
    Write-Host "Onde:"
    Write-Host "- target-image (string):       O repositório da imagem a ser analisada, sem a tag (e.g. companyrepo/whateverisinside)."
    Write-Host "- config-file-path (string):   O caminho (absoluto) do arquivo de configurações."
    Write-Host "                               Valor padrão: `"$ConfigPath`""
    Write-Host "- dive-version (string):       A versão, em formato de versionamento semântico da aplicação Dive."
    Write-Host "                               Valor padrão: `"latest`""
    Write-Host ""
    Write-Host "Exemplo: ./docker-analysis.ps1 $Mode docker.io/wagoodman/dive C:\foo\.dive.yaml 0.9.2"
    Write-Host "----------------------------------------------------------------------------------------------------------------------------------"
}

function Trim-String {
    param (
        [string]$InputString
    )
    if ($InputString) {
        return $InputString.Trim()
    }
    return ""
}

function Validate-Input {
    param (
        [string]$Mode,
        [string]$TargetImage
    )

    if (-not $TargetImage) {
        Write-Host "ERRO: O repositório para a imagem Docker deve ser especificado!" -ForegroundColor Red
        Show-Usage $Mode
        exit 1
    }
}

function Invoke-DiveUI {
    param (
        [string]$TargetImage,
        [string]$ConfigPath,
        [string]$DiveVersion
    )

    $TargetImage = Trim-String $TargetImage
    $ConfigPath = Trim-String (if ($ConfigPath) { $ConfigPath } else { $DEFAULT_DIVE_UI_CONFIG_PATH })
    $DiveVersion = Trim-String (if ($DiveVersion) { $DiveVersion } else { $DEFAULT_DIVE_VERSION })

    Validate-Input "ui" $TargetImage

    $DiveImage = "$DEFAULT_DIVE_IMAGE`:$DiveVersion"

    docker run --rm -it `
        -v $DEFAULT_DOCKER_SOCK_PATH`:$DEFAULT_DOCKER_SOCK_PATH `
        -v "$(Get-Location):$(Get-Location)" `
        -v $ConfigPath`:$env:USERPROFILE/.dive.yaml `
        $DiveImage `
        $TargetImage

    exit 0
}

function Invoke-DiveCI {
    param (
        [string]$TargetImage,
        [string]$ConfigPath,
        [string]$DiveVersion
    )

    $TargetImage = Trim-String $TargetImage
    $ConfigPath = Trim-String (if ($ConfigPath) { $ConfigPath } else { $DEFAULT_DIVE_CI_CONFIG_PATH })
    $DiveVersion = Trim-String (if ($DiveVersion) { $DiveVersion } else { $DEFAULT_DIVE_VERSION })

    Validate-Input "ci" $TargetImage

    $DiveImage = "$DEFAULT_DIVE_IMAGE`:$DiveVersion"

    docker run `
        --rm `
        -it `
        -e CI=true `
        -v $DEFAULT_DOCKER_SOCK_PATH`:$DEFAULT_DOCKER_SOCK_PATH `
        $DiveImage `
        --ci-config $ConfigPath `
        $TargetImage

    exit 0
}

function Get-ContainerVolumes {
    $containers = docker ps -a --format "{{.ID}}"
    foreach ($d in $containers) {
        $d_name = docker inspect -f "{{.Name}}" $d
        Write-Host "========================================================="
        Write-Host "$d_name ($d) volumes:"

        $VOLUME_IDS = docker inspect -f "{{.Config.Volumes}}" $d
        $VOLUME_IDS = $VOLUME_IDS -replace 'map\[', '' -replace '\]', ''

        $array = $VOLUME_IDS -split '\s+'

        foreach ($item in $array) {
            if ($item -and $item -ne '') {
                $VOLUME_ID = $item -replace ':{}', ''
                try {
                    $VOLUME_SIZE = docker exec -ti $d_name du -d 0 -h $VOLUME_ID 2>$null
                    Write-Host "$VOLUME_SIZE"
                }
                catch {
                    Write-Host "  (Unable to check size for $VOLUME_ID)"
                }
            }
        }
    }

    exit 0
}

# Main execution logic
if ($args.Count -gt 0) {
    $mode = $args[0].ToLower()
    switch ($mode) {
        "ui" {
            Invoke-DiveUI $args[1] $args[2] $args[3]
        }
        "ci" {
            Invoke-DiveCI $args[1] $args[2] $args[3]
        }
        "volumes" {
            Get-ContainerVolumes
        }
        default {
            Show-Usage $mode
            exit 1
        }
    }
}
else {
    Write-Host "Uso: ./docker-analysis.ps1 <ui|ci|volumes> [parâmetros...]" -ForegroundColor Yellow
    exit 1
}
