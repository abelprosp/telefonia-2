<#
.SYNOPSIS
    Equivalente em PowerShell do Makefile original.

.DESCRIPTION
    Reproduz os alvos (targets) do Makefile como comandos.
    Uso: .\make.ps1 <comando> [chave=valor ...]

    Exemplos:
        .\make.ps1 help
        .\make.ps1 build env=dev
        .\make.ps1 up env=dev c=api
        .\make.ps1 logs env=prod c=worker
        .\make.ps1 run env=dev c=api cmd="npm test"
        .\make.ps1 exec env=dev c=api s="bash"
        .\make.ps1 imganalysisui img=myimage:latest

.NOTES
    Versão 100% PowerShell (sem dependência de Bash/WSL/Git Bash).

    Os scripts auxiliares originais (scripts/docker-compose.sh,
    scripts/docker-analysis.sh, scripts/generate-topology.sh) eram scripts Bash
    cujo conteúdo não estava disponível para portar fielmente. As funções abaixo
    reimplementam o comportamento esperado com base em convenções razoáveis —
    ver comentários "ASSUNÇÃO" em cada função. Ajuste conforme a lógica real dos
    seus scripts originais, se ela for diferente.
#>

param(
    [Parameter(Position = 0)]
    [string]$Command = "help",

    [Parameter(Position = 1, ValueFromRemainingArguments = $true)]
    [string[]]$RemainingArgs
)

# ---------------------------------------------------------------------------
# Diretório-raiz (equivalente a ROOT_DIR)
# ---------------------------------------------------------------------------
$RootDir = $PSScriptRoot

# ---------------------------------------------------------------------------
# Força saída em UTF-8 no console (necessário no Windows PowerShell 5.1 para
# exibir corretamente acentos e caracteres especiais em português).
# ---------------------------------------------------------------------------
try {
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    $OutputEncoding = [System.Text.Encoding]::UTF8
    if ($PSVersionTable.PSEdition -eq 'Desktop') {
        # Windows PowerShell (5.1): também ajusta o codepage do console para 65001 (UTF-8)
        & chcp.com 65001 > $null
    }
}
catch {
    Write-Warning "Não foi possível forçar a codificação UTF-8 do console: $($_.Exception.Message)"
}

# ---------------------------------------------------------------------------
# Parse de argumentos no estilo Makefile: chave=valor
# ---------------------------------------------------------------------------
$Vars = @{}
foreach ($arg in $RemainingArgs) {
    if ($arg -match '^([a-zA-Z_]+)=(.*)$') {
        $Vars[$matches[1]] = $matches[2]
    }
}

$EnvName = $Vars['env']
$C = $Vars['c']
$Cmd = $Vars['cmd']
$S = $Vars['s']
$Img = $Vars['img']

# ---------------------------------------------------------------------------
# Header ASCII (equivalente a $(HEADER))
# ---------------------------------------------------------------------------
$Header = @"
+---------------------------------------------------------------------------------------------+
  _                       ___                      _     _    __  __      _        __ _ _
 | |  _  ___ ___  _ ___  / __|___ _ _  _ _  ___ __| |_  | |  |  \/  |__ _| |_____ / _(_) |___
 | |_| || \ \ / || (_-< | (__/ _ \ ' \| ' \/ -_) _|  _| | |  | |\/| / _` | / / -_)  _| | / -_)
 |____\_,_/_\_\\_,_/__/  \___\___/_||_|_||_\___\__|\__| | |  |_|  |_\__,_|_\_\___|_| |_|_\___|
                                                        |_|
+---------------------------------------------------------------------------------------------+
"@

# ---------------------------------------------------------------------------
# Metadados de ajuda: categoria + descrição de cada comando
# (equivalente ao parsing de '##@categoria descricao' do Makefile)
# ---------------------------------------------------------------------------
$CommandHelp = [ordered]@{
    header        = @{ Category = 'Outros'; Desc = 'Mostra o header deste help, formado com caracteres ASCII.' }
    help          = @{ Category = 'Outros'; Desc = 'Mostra esta documentação.' }
    listsrvs      = @{ Category = 'Comandos'; Desc = 'Lista todos os nomes de serviços declarados no YAML do Docker Compose, dado um env=<dev | prod> ambiente de infra' }
    build         = @{ Category = 'Comandos'; Desc = 'Realiza a build de todas as imagens Docker, ou para um c=<nome de serviço> específico, dado um env=<dev | prod> ambiente de infra' }
    clean         = @{ Category = 'Comandos'; Desc = 'Realiza a limpeza de todos os dados associados aos conteineres, dado um env=<dev | prod> ambiente de infra' }
    destroy       = @{ Category = 'Comandos'; Desc = 'Remove todas as imagens, volumes, networks e conteineres não utilizados. Use com cautela!' }
    logs          = @{ Category = 'Comandos'; Desc = 'Adiciona captura de logs para todos os conteineres ou para um c=<nome de serviço>, dado um env=<dev | prod> ambiente de infra' }
    restart       = @{ Category = 'Comandos'; Desc = 'Reinicia todos os conteineres ou apenas um c=<nome de serviço>, dado um env=<dev | prod> ambiente de infra' }
    start         = @{ Category = 'Comandos'; Desc = 'Inicia todos os conteineres em background (detached mode) ou apenas um c=<nome de serviço>, dado um env=<dev | prod> ambiente de infra' }
    init          = @{ Category = 'Comandos'; Desc = 'Inicia um conteiner em detached mode, com captura de logs, dado um env=<dev | prod> ambiente de infra' }
    status        = @{ Category = 'Comandos'; Desc = 'Lista os status dos conteineres em execução, dado um env=<dev | prod> ambiente de infra' }
    stop          = @{ Category = 'Comandos'; Desc = 'Encerra a execução de todos os conteineres ou de apenas um c=<nome de serviço>, dado um env=<dev | prod> ambiente de infra' }
    up            = @{ Category = 'Comandos'; Desc = 'Inicia todos os conteineres em modo "attached" ou apenas um c=<nome de serviço>, dado um env=<dev | prod> ambiente de infra' }
    run           = @{ Category = 'Comandos'; Desc = "Roda um comando (o que seria especificado em 'CMD' na imagem), dado um c=<nome de serviço> e um env=<dev | prod> ambiente de infra" }
    exec          = @{ Category = 'Comandos'; Desc = 'Executa um comando em um container já iniciado, dado um c=<nome de serviço> e um s=<script> e um env=<dev | prod> ambiente de infra' }
    ps            = @{ Category = 'Comandos'; Desc = "Alias do comando 'status'" }
    imganalysisui = @{ Category = 'Comandos'; Desc = 'Executa a análise de uma imagem Docker, em modo UI, dado uma img=<imagem Docker>' }
    imganalysisci = @{ Category = 'Comandos'; Desc = 'Executa a análise de uma imagem Docker, em modo CI, dado uma img=<imagem Docker>' }
    topology      = @{ Category = 'Comandos'; Desc = 'Gera um diagrama dos serviços listados no arquivo YML do Docker Compose' }
}

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

function Get-ComposeYamlPath {
    <#
        Equivalente a 'docker-compose.sh yamlpath <env>'.

        O arquivo de compose segue a convenção 'docker-compose.<env>.yml'
        na raiz do projeto (ex: docker-compose-dev.yml, docker-compose.prod.yml),
        com fallback para 'docker-compose.yml' quando nenhum env= for informado.
        Ajuste esta função se seu script original resolvia o caminho de outra forma
        (ex: pasta 'deploy/<env>/docker-compose.yml', arquivo único com profiles, etc).
    #>
    param([string]$EnvArg)

    if ([string]::IsNullOrWhiteSpace($EnvArg)) {
        return "docker-compose.yml"
    }

    $candidate = "docker-compose-$($EnvArg.ToLower()).yml"
    $fullPath = Join-Path $RootDir $candidate

    if (-not (Test-Path $fullPath)) {
        Write-Host "Aviso: '$candidate' não encontrado em $RootDir. Verifique a convenção de nomes ou ajuste Get-ComposeYamlPath." -ForegroundColor Yellow
    }

    return $candidate
}

function Invoke-ComposeCmd {
    <# Equivalente à macro 'compose_cmd' do Makefile #>
    param(
        [string]$EnvArg,
        [string[]]$ExtraArgs
    )

    $extraArgsStr = ($ExtraArgs -join ' ').Trim()

    Write-Host "call_compose_cmd @ [ENV($EnvArg)] & [ARGS($extraArgsStr)]"
    Write-Host ('-' * 99)

    $yamlPath = Get-ComposeYamlPath -EnvArg $EnvArg
    $composeFile = Join-Path $RootDir $yamlPath

    $dockerArgs = @('compose', '-f', $composeFile) + $ExtraArgs
    & docker @dockerArgs
}

function Confirm-Action {
    <# Equivalente ao alvo 'confirm' #>
    $sure = Read-Host "Tem certeza? [y/N]"
    if ($sure -match '^[sSyY]$') {
        return $true
    }
    return $false
}

function Show-Header {
    Clear-Host
    Write-Host $Header
}

function Show-Help {
    Clear-Host
    Write-Host $Header

    Write-Host "Utilização: .\make.ps1 [comando]`n"

    $categories = $CommandHelp.Values | ForEach-Object { $_.Category } | Select-Object -Unique | Sort-Object

    foreach ($category in $categories) {
        Write-Host "${category}:" -ForegroundColor White

        $entries = $CommandHelp.GetEnumerator() | Where-Object { $_.Value.Category -eq $category }

        foreach ($entry in $entries) {
            $name = $entry.Key
            $desc = $entry.Value.Desc
            $sep = ' ' * [Math]::Max(1, 32 - $name.Length)

            Write-Host -NoNewline "  "
            Write-Host -NoNewline $name -ForegroundColor Yellow
            Write-Host -NoNewline $sep
            Write-Host $desc -ForegroundColor Green
        }
        Write-Host ""
    }
}

# ---------------------------------------------------------------------------
# Alvos (targets)
# ---------------------------------------------------------------------------

function Invoke-ListSrvs { Invoke-ComposeCmd -EnvArg $EnvName -ExtraArgs @('config', '--services') }

function Invoke-Build {
    $extra = @('build')
    if ($C) { $extra += $C }
    Invoke-ComposeCmd -EnvArg $EnvName -ExtraArgs $extra
}

function Invoke-Clean {
    if (-not (Confirm-Action)) { return }
    Invoke-ComposeCmd -EnvArg $EnvName -ExtraArgs @('down')
}

function Invoke-Destroy {
    if (-not (Confirm-Action)) { return }
    & docker system prune --all --volumes --force
    & docker volume prune --all --force
    & docker network prune --force
    & docker image prune --all --force
}

function Invoke-Logs {
    $extra = @('logs', '--follow')
    if ($C) { $extra += $C }
    Invoke-ComposeCmd -EnvArg $EnvName -ExtraArgs $extra
}

function Invoke-Restart {
    $extra = @('stop')
    if ($C) { $extra += $C }
    Invoke-ComposeCmd -EnvArg $EnvName -ExtraArgs $extra
    Invoke-Init
}

function Invoke-Start {
    $extra = @('up', '-d')
    if ($C) { $extra += $C }
    Invoke-ComposeCmd -EnvArg $EnvName -ExtraArgs $extra
}

function Invoke-Init {
    Invoke-Start
    Invoke-Logs
}

function Invoke-Status { Invoke-ComposeCmd -EnvArg $EnvName -ExtraArgs @('ps') }

function Invoke-Stop {
    $extra = @('stop')
    if ($C) { $extra += $C }
    Invoke-ComposeCmd -EnvArg $EnvName -ExtraArgs $extra
}

function Invoke-Up {
    $extra = @('up')
    if ($C) { $extra += $C }
    Invoke-ComposeCmd -EnvArg $EnvName -ExtraArgs $extra
}

function Invoke-Run {
    $extra = @('run', '--rm')
    if ($C) { $extra += $C }
    if ($Cmd) { $extra += $Cmd }
    Invoke-ComposeCmd -EnvArg $EnvName -ExtraArgs $extra
}

function Invoke-Exec {
    $extra = @('exec', '-it')
    if ($C) { $extra += $C }
    if ($S) { $extra += $S }
    Invoke-ComposeCmd -EnvArg $EnvName -ExtraArgs $extra
}

function Invoke-ImgAnalysisUi {
    <#
        Equivalente a 'docker-analysis.sh ui <img>'.
        ASSUNÇÃO: modo UI usa a ferramenta 'dive' (https://github.com/wagoodman/dive)
        para explorar interativamente as camadas da imagem. Ajuste se o script
        original usava outra ferramenta.
    #>
    if (-not $Img) {
        Write-Host "Erro: informe img=<imagem Docker>" -ForegroundColor Red
        return
    }
    if (-not (Get-Command dive -ErrorAction SilentlyContinue)) {
        Write-Host "Erro: 'dive' não encontrado no PATH. Instale-o ou ajuste Invoke-ImgAnalysisUi para a ferramenta real usada." -ForegroundColor Red
        return
    }
    & dive $Img
}

function Invoke-ImgAnalysisCi {
    <#
        Equivalente a 'docker-analysis.sh ci <img>'.
        ASSUNÇÃO: modo CI usa 'trivy' (https://github.com/aquasecurity/trivy) para
        escanear vulnerabilidades e retornar código de saída não-zero em caso de
        falha, adequado para pipelines. Ajuste se o script original usava outra
        ferramenta (ex: dive --ci, grype, etc).
    #>
    if (-not $Img) {
        Write-Host "Erro: informe img=<imagem Docker>" -ForegroundColor Red
        return
    }
    if (-not (Get-Command trivy -ErrorAction SilentlyContinue)) {
        Write-Host "Erro: 'trivy' não encontrado no PATH. Instale-o ou ajuste Invoke-ImgAnalysisCi para a ferramenta real usada." -ForegroundColor Red
        return
    }
    & trivy image $Img
}

function Invoke-Topology {
    <#
        Equivalente a 'generate-topology.sh topology <env>'.
        Não há uma convenção universal para geração de diagramas de topologia a
        partir de um Docker Compose, então esta função é um placeholder explícito.
        Substitua o bloco abaixo pela lógica real (ex: parsear o YAML resolvido por
        Get-ComposeYamlPath e gerar um .dot/.svg com Graphviz, ou chamar uma
        ferramenta específica).
    #>
    $yamlPath = Get-ComposeYamlPath -EnvArg $EnvName
    $composeFile = Join-Path $RootDir $yamlPath

    Write-Host "TODO: Invoke-Topology não implementado." -ForegroundColor Yellow
    Write-Host "Arquivo de compose resolvido: $composeFile"
    Write-Host "Implemente aqui a geração real do diagrama de topologia."
}

# ---------------------------------------------------------------------------
# Dispatcher (equivalente ao .DEFAULT_GOAL := help + alvos do Makefile)
# ---------------------------------------------------------------------------
switch ($Command.ToLower()) {
    'header' { Show-Header }
    'help' { Show-Help }
    'info' { Show-Header }
    'listsrvs' { Invoke-ListSrvs }
    'build' { Invoke-Build }
    'clean' { Invoke-Clean }
    'destroy' { Invoke-Destroy }
    'logs' { Invoke-Logs }
    'restart' { Invoke-Restart }
    'start' { Invoke-Start }
    'init' { Invoke-Init }
    'status' { Invoke-Status }
    'stop' { Invoke-Stop }
    'up' { Invoke-Up }
    'run' { Invoke-Run }
    'exec' { Invoke-Exec }
    'ps' { Invoke-Status }
    'imganalysisui' { Invoke-ImgAnalysisUi }
    'imganalysisci' { Invoke-ImgAnalysisCi }
    'topology' { Invoke-Topology }
    default {
        Write-Host "Comando desconhecido: $Command" -ForegroundColor Red
        Show-Help
    }
}
