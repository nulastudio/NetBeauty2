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

# 编译nbloader
cd "${rootdir}/nbloader"

if (Test-Path -Path "bin/Release") {
    Remove-Item -Recurse -Force "bin/Release"
}

dotnet build -c Release /p:OutputPath="bin/Release"

$nbloader_dll = "${rootdir}/nbloader/bin/Release/nbloader.dll"

if (Test-Path -Path $nbloader_dll) {
    # 签名nbloader
    pwsh "${builddir}/sign.ps1" -Certificate Auto -Algorithm SHA384 -TimeStampServer "http://timestamp.sectigo.com" $nbloader_dll

    # 复制nbloader
    Copy-Item -Force $nbloader_dll "${rootdir}/NetBeauty/src/nbloader/nbloader.dll"
}

# 更新nbloader
cd "${rootdir}/NetBeauty/src"
go-bindata -o ./main/bindata.go ./nbloader/

# 编译nbeauty
cd "${rootdir}/NetBeauty"
if ([System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Windows)) {
    pwsh "make.ps1"
} else {
    make
}

# 签名nbeauty
pwsh "${builddir}/sign.ps1" -Certificate Auto -Algorithm SHA384 -TimeStampServer "http://timestamp.sectigo.com" "${tooldir}/win-x86/nbeauty2.exe" "${tooldir}/win-x64/nbeauty2.exe" "${tooldir}/win-arm64/nbeauty2.exe"

# 编译NetBeautyNuget
cd "${rootdir}/NetBeautyNuget"
dotnet pack -c Release /p:PackageOutputPath=${nupkgdir}

# 编译NetBeautyGlobalTool
cd "${rootdir}/NetBeautyGlobalTool"
dotnet pack -c Release /p:PackageOutputPath=${nupkgdir}

# 打包nbeauty
mkdir ${archivedir}
"win-x86", "win-x64", "win-arm64", "linux-x64", "osx-x64", "osx-arm64" | ForEach-Object -Process {
    $rid = $_
    cd "${tooldir}/${rid}"
    Compress-Archive -Force -Path * -DestinationPath "${archivedir}/${rid}.zip"
}

cd $workingdir
