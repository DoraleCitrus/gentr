# GitHub User/Repo
$Owner = "DoraleCitrus"
$Repo = "gentr"
$Binary = "gentr.exe"

Write-Host "Checking for latest version..." -ForegroundColor Cyan

# ---------------------------------------------------------
# [修改] 获取最新版本号 (不使用 API，改用网页重定向，避开速率限制)
# ---------------------------------------------------------
$LatestUrl = "https://github.com/$Owner/$Repo/releases/latest"
try {
    # 发送请求，但不自动跟随重定向 (-MaximumRedirection 0)
    # 这样我们可以捕获 302 跳转的目标地址
    $Response = Invoke-WebRequest -Uri $LatestUrl -MaximumRedirection 0 -ErrorAction SilentlyContinue
}
catch {
    # PowerShell 会把重定向视为错误，我们需要从 Exception 中取出 Response
    $Response = $_.Exception.Response
}

# 从 Location 头中提取 URL，例如: https://github.com/DoraleCitrus/gentr/releases/tag/v1.0.0
$RedirectUrl = $Response.Headers.Location
if (-not $RedirectUrl) {
    # 如果重定向失败，尝试让它自动跟随一次 (有时候 PowerShell 版本差异)
    try {
        $Req = Invoke-WebRequest -Uri $LatestUrl -Method Head -ErrorAction SilentlyContinue
        $RedirectUrl = $Req.BaseResponse.ResponseUri.AbsoluteUri
    }
    catch {
        Write-Error "Failed to determine latest version. Please try again later."
        exit 1
    }
}

# 提取 tag 部分 (获取 v1.0.0)
$VersionTag = $RedirectUrl.Split('/')[-1]
# 去掉 v 前缀 (获取 1.0.0)
$VersionNum = $VersionTag.TrimStart('v')

if (-not $VersionNum) {
    Write-Error "Failed to parse version number from URL: $RedirectUrl"
    exit 1
}

Write-Host "Detected latest version: $VersionTag" -ForegroundColor Cyan
# ---------------------------------------------------------

# 构建下载链接 (Windows amd64)
$ZipName = "${Repo}_${VersionNum}_windows_amd64.zip"
$DownloadUrl = "https://github.com/$Owner/$Repo/releases/download/$VersionTag/$ZipName"

# 创建临时目录
$TempDir = Join-Path $env:TEMP "gentr_install"
if (Test-Path $TempDir) { Remove-Item $TempDir -Recurse -Force }
New-Item -ItemType Directory -Force -Path $TempDir | Out-Null
$ZipPath = Join-Path $TempDir $ZipName

# 下载
Write-Host "Downloading from $DownloadUrl..."
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -ErrorAction Stop
}
catch {
    Write-Error "Download failed. Please check your internet connection or if the release asset exists."
    exit 1
}

# 解压
Write-Host "Extracting..."
try {
    Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force -ErrorAction Stop
}
catch {
    Write-Error "Extraction failed. The zip file might be corrupted."
    exit 1
}

# 安装位置
$InstallDir = Join-Path $env:USERPROFILE "gentr"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

# 移动二进制文件
# 注意：解压后的结构可能是直接在根目录，也可能在子目录，这里做一个查找
$ExePath = Get-ChildItem -Path $TempDir -Filter $Binary -Recurse | Select-Object -First 1 -ExpandProperty FullName

if (-not $ExePath) {
    Write-Error "Could not find $Binary in the downloaded zip."
    exit 1
}

Move-Item -Path $ExePath -Destination (Join-Path $InstallDir $Binary) -Force

# 清理
Remove-Item -Path $TempDir -Recurse -Force

Write-Host "✅ Installed to $InstallDir" -ForegroundColor Green

# 添加到 PATH
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "⚠️  Action Required: Add '$InstallDir' to your PATH environment variable." -ForegroundColor Yellow
    Write-Host "   Run this command to add it permanently:"
    Write-Host "   [Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + ';$InstallDir', 'User')" -ForegroundColor Gray
}
else {
    Write-Host "Run 'gentr' to start!" -ForegroundColor Green
}