<#
.SYNOPSIS
    generate-topology.ps1
.DESCRIPTION
    Just an ad-hoc PS script made for updating the docker compose topology PNG in the main docs.
    Still a pile of trash, hard-coded params for develop env only.
    Must port the corresponiding shell script to a PS syntax. But later...
#>

$dockerComposeDiagramLocation = "./docs/images"

$dockerComposeFilePath = "docker-compose.yml"

$outputFile = "docker-topology-develop.png"

$cwd = $(Get-Location)

Clear-Host;

docker run `
    --rm `
    -it `
    --name dcv `
    -v "${cwd}:/input:rw" `
    -v "${cwd}/${dockerComposeDiagramLocation}:/output:rw" `
    pmsipilot/docker-compose-viz `
    render `
        -m `
        image `
        --force `
        --horizontal `
        --output-file /output/${outputFile} `
        ${dockerComposeFilePath};

Write-Output "Topology diagram created/updated in '${dockerComposeDiagramLocation}/${outputFile}!'";
