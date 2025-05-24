$workingdir = $pwd
$builddir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$rootdir = "${builddir}/.."
$tooldir = "${builddir}/tools"
$nupkgdir = "${builddir}/nupkg"
$archivedir = "${builddir}/archive"

# 清理
if (Test-Path -Path $tooldir) {
    Remove-Item -Recurse -Force $tooldir
}
if (Test-Path -Path $nupkgdir) {
    Remove-Item -Recurse -Force $nupkgdir
    mkdir $nupkgdir
}
if (Test-Path -Path $archivedir) {
    Remove-Item -Recurse -Force $archivedir
}

# 清理NetBeauty Nuget缓存
$cachedir = $(dotnet nuget locals global-packages -l).Replace("global-packages: ", "")
$cachedir = "${cachedir}/nulastudio.netbeauty"

if (Test-Path -Path $cachedir) {
    Remove-Item -Recurse -Force $cachedir
}

# 编译Loader
Set-Location "${rootdir}/libloader"

if (Test-Path -Path "bin/Release") {
    Remove-Item -Recurse -Force "bin/Release"
}

dotnet build -c Release /p:OutputPath="bin/Release"

$loaderdll = "${rootdir}/libloader/bin/Release/libloader.dll"

if (Test-Path -Path $loaderdll) {
    # 签名Loader
    # pwsh "${builddir}/sign.ps1" -Certificate Auto -Algorithm SHA384 -TimeStampServer "http://timestamp.sectigo.com" $loaderdll

    # 复制Loader
    Copy-Item -Force $loaderdll "${rootdir}/NetBeauty/src/libloader/libloader.dll"
}

# 更新Loader
Set-Location "${rootdir}/NetBeauty/src"
go-bindata -o ./main/bindata.go ./libloader/

# 编译nbeauty
Set-Location "${rootdir}/NetBeauty"
if ([System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Windows)) {
    pwsh "make.ps1"
} else {
    make
}

# 签名nbeauty
# pwsh "${builddir}/sign.ps1" -Certificate Auto -Algorithm SHA384 -TimeStampServer "http://timestamp.sectigo.com" "${tooldir}/win-x86/nbeauty2.exe" "${tooldir}/win-x64/nbeauty2.exe" "${tooldir}/win-arm64/nbeauty2.exe"

# 编译NetBeautyNuget
Set-Location "${rootdir}/NetBeautyNuget"
dotnet pack -c Release /p:PackageOutputPath=${nupkgdir}

# 编译NetBeautyGlobalTool
Set-Location "${rootdir}/NetBeautyGlobalTool"
dotnet pack -c Release /p:PackageOutputPath=${nupkgdir}

# 打包nbeauty
mkdir ${archivedir}
"win-x86", "win-x64", "win-arm64", "linux-x64", "osx-x64", "osx-arm64" | ForEach-Object -Process {
    $rid = $_
    Set-Location "${tooldir}/${rid}"
    Compress-Archive -Force -Path * -DestinationPath "${archivedir}/${rid}.zip"
}

Set-Location $workingdir
