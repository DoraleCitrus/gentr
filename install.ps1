# GitHub User/Repo
$Owner = "DoraleCitrus"
$Repo = "gentr"
$Binary = "gentr.exe"

# 获取最新 Release 版本号
$LatestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/$Owner/$Repo/releases/latest"
$Version = $LatestRelease.tag_name
$VersionNum = $Version.TrimStart('v')

Write-Host "Detected latest version: $Version" -ForegroundColor Cyan

# 构建下载链接 (Windows amd64)
$ZipName = "${Repo}_${VersionNum}_windows_amd64.zip"
$DownloadUrl = "https://github.com/$Owner/$Repo/releases/download/$Version/$ZipName"

# 创建临时目录
$TempDir = Join-Path $env:TEMP "gentr_install"
New-Item -ItemType Directory -Force -Path $TempDir | Out-Null
$ZipPath = Join-Path $TempDir $ZipName

# 下载
Write-Host "Downloading from $DownloadUrl..."
Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath

# 解压
Write-Host "Extracting..."
Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force

# 安装位置 (用户目录下的 gentr 文件夹)
$InstallDir = Join-Path $env:USERPROFILE "gentr"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

# 移动二进制文件
Move-Item -Path (Join-Path $TempDir $Binary) -Destination (Join-Path $InstallDir $Binary) -Force

# 清理
Remove-Item -Path $TempDir -Recurse -Force

Write-Host "✅ Installed to $InstallDir" -ForegroundColor Green

# 添加到 PATH (仅当前会话有效，永久生效需要提示用户)
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "⚠️  Action Required: Add '$InstallDir' to your PATH environment variable to run 'gentr' from anywhere." -ForegroundColor Yellow
    Write-Host "   Run this command to add it permanently:"
    Write-Host "   [Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + ';$InstallDir', 'User')" -ForegroundColor Gray
}
else {
    Write-Host "Run 'gentr' to start!" -ForegroundColor Green
}