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
}
if (Test-Path -Path $archivedir) {
    Remove-Item -Recurse -Force $archivedir
}

# 编译nbeauty
cd "${rootdir}/NetBeauty"
if ([System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Windows)) {
    pwsh "make.ps1"
} else {
    make
}

# 编译NetBeautyNuget
cd "${rootdir}/NetBeautyNuget"
dotnet pack -c Release /p:PackageOutputPath=${nupkgdir}

# 编译NetBeautyGlobalTool
cd "${rootdir}/NetBeautyGlobalTool"
dotnet pack -c Release /p:PackageOutputPath=${nupkgdir}

# 打包nbeauty
mkdir ${archivedir}
"win-x86", "win-x64", "linux-x64", "osx-x64" | ForEach-Object -Process {
    $rid = $_
    cd "${tooldir}/${rid}"
    Compress-Archive -Force -Path * -DestinationPath "${archivedir}/${rid}.zip"
}

cd $workingdir
