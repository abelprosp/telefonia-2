$goVersion = "1.26.4"
$arch = "amd64"
$goTarballName = "go-${goVersion}-${arch}.zip"

Invoke-WebRequest -Uri "https://go.dev/dl/go${goVersion}.windows-${arch}.zip" -OutFile $goTarballName

# Extract to your AppData\Local directory
Expand-Archive -Path $goTarballName -DestinationPath "$HOME\AppData\Local" -Force

# Rename the folder to simply 'go' for easier pathing
Rename-Item -Path "$HOME\AppData\Local\go" -NewName "$HOME\AppData\Local\Go"

# Set GOROOT (the location of the Go installation)
[System.Environment]::SetEnvironmentVariable("GOROOT", "$HOME\AppData\Local\Go", "User")

# Set GOPATH (where your go projects and downloaded modules will live)
[System.Environment]::SetEnvironmentVariable("GOPATH", "$HOME\go", "User")

# Add Go's bin folder to your User PATH
$oldPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
$newPath = "$oldPath;$HOME\AppData\Local\Go\bin;$HOME\go\bin"

[System.Environment]::SetEnvironmentVariable("Path", $newPath, "User")
