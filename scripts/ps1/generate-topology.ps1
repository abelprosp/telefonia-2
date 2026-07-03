# <#
# .SYNOPSIS
#     generate-topology.ps1
# .DESCRIPTION
#     Just an ad-hoc PS script made for updating the docker compose topology PNG in the main docs.
#     Still a pile of trash, hard-coded params for develop env only.
#     Must port the corresponiding shell script to a PS syntax. But later...
# #>

# $dockerComposeDiagramLocation = "./docs/images"

# $dockerComposeFilePath = "docker-compose.yml"

# $outputFile = "docker-topology-develop.png"

# $cwd = $(Get-Location)

# Clear-Host;

# docker run `
#     --rm `
#     -it `
#     --name dcv `
#     -v "${cwd}:/input:rw" `
#     -v "${cwd}/${dockerComposeDiagramLocation}:/output:rw" `
#     pmsipilot/docker-compose-viz `
#     render `
#         -m `
#         image `
#         --force `
#         --horizontal `
#         --output-file /output/${outputFile} `
#         ${dockerComposeFilePath};

# Write-Output "Topology diagram created/updated in '${dockerComposeDiagramLocation}/${outputFile}!'";


<#
.SYNOPSIS
    Equivalente em PowerShell do script generate-topology.sh.

.DESCRIPTION
    Gera um diagrama de topologia (PNG) a partir de um arquivo Docker Compose,
    usando a imagem 'pmsipilot/docker-compose-viz'.

.USAGE
    .\generate-topology.ps1 topology <env>

    Exemplos:
        .\generate-topology.ps1 topology dev
        .\generate-topology.ps1 topology prod
#>

param(
    [Parameter(Position = 0)]
    [string]$FunctionName,

    [Parameter(Position = 1, ValueFromRemainingArguments = $true)]
    [string[]]$FunctionArgs
)

function Topology {
    param([string]$Environment)

    $SupportedEnvs = @('dev', 'develop', 'development', 'desenvolvimento', 'prod', 'production', 'producao')
    $ComposeYamlBasename = 'docker-compose'

    switch -Regex ($Environment) {
        '^(dev|develop|development|desenvolvimento)$' {
            $EnvFolder = 'develop'
            $DockerComposeFile = "$ComposeYamlBasename.yml"
            break
        }
        '^(prod|production|producao)$' {
            $EnvFolder = 'production'
            $DockerComposeFile = "$ComposeYamlBasename.yml"
            break
        }
        default {
            $choices = $SupportedEnvs -join ','
            Write-Host -NoNewline "Environment '$Environment' not supported. Choices are: "
            Write-Host $choices
            exit 1
        }
    }

    $DockerComposeDiagramLocation = "docs/images"

    Write-Host "Generating topology diagram for '$Environment' environment..."

    $OutputFile = "docker-topology-$EnvFolder.png"
    $DockerComposeFilePath = "infra/docker/$EnvFolder/$DockerComposeFile"

    $currentDir = (Get-Location).Path
    $diagramDir = Join-Path $currentDir $DockerComposeDiagramLocation

    # Equivalente a 'chmod 777 ./resources/docs/images'. Permissões POSIX não se
    # aplicam da mesma forma no Windows; garantimos apenas que o diretório exista.
    # Em Linux/macOS (pwsh), replicamos o chmod original.
    if (-not (Test-Path $diagramDir)) {
        New-Item -ItemType Directory -Path $diagramDir -Force | Out-Null
    }
    if (-not $IsWindows) {
        & chmod 777 $diagramDir
    }

    Write-Host "Docker Compose YAML to be used:   $DockerComposeFilePath"
    Write-Host "Topology diagram PNG saved in:    $DockerComposeDiagramLocation/$OutputFile"

    $dockerArgs = @(
        'run',
        '--rm',
        '-it',
        '--name', 'dcv'
    )

    # Equivalente a '-u $(id -u):$(id -g)'. Só faz sentido em Linux/macOS;
    # no Windows não há UID/GID de usuário para mapear no container.
    if (-not $IsWindows) {
        $uid = & id -u
        $gid = & id -g
        $dockerArgs += @('-u', "${uid}:${gid}")
    }

    $dockerArgs += @(
        '-v', "${currentDir}:/input:rw",
        '-v', "${diagramDir}:/output:rw",
        'pmsipilot/docker-compose-viz',
        'render',
        '-m', 'image',
        '--force',
        '--horizontal',
        '--output-file', "/output/$OutputFile",
        $DockerComposeFilePath
    )

    & docker @dockerArgs
}

# ---------------------------------------------------------------------------
# Dispatcher: equivalente a '$@' no final do script Bash original, que chama
# a função cujo nome é o primeiro argumento (ex: './generate-topology.sh topology dev').
# ---------------------------------------------------------------------------
if ($FunctionName) {
    if (Get-Command -Name $FunctionName -CommandType Function -ErrorAction SilentlyContinue) {
        & $FunctionName @FunctionArgs
    }
    else {
        Write-Host "Função '$FunctionName' não encontrada neste script." -ForegroundColor Red
        exit 1
    }
}