$workingdir = $pwd
$builddir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$rootdir = "${builddir}/.."
$tooldir = "${builddir}/tools"
$nupkgdir = "${builddir}/nupkg"
$archivedir = "${builddir}/archive"

# 清理
rm -rf $tooldir
rm -rf $nupkgdir
rm -rf $archivedir

# 编译ncbeauty
cd $rootdir
make

# 编译NetCoreBeautyNuget
cd "${rootdir}/NetCoreBeautyNuget"
dotnet pack -c Release /p:PackageOutputPath=${nupkgdir}

# 编译NetCoreBeautyGlobalTool
cd "${rootdir}/NetCoreBeautyGlobalTool"
dotnet pack -c Release /p:PackageOutputPath=${nupkgdir}

# 打包ncbeauty
mkdir ${archivedir}
"win-x86", "win-x64", "linux-x64", "osx-x64" | ForEach-Object -Process {
    $rid = $_
    cd "${tooldir}/${rid}"
    Compress-Archive -Force -Path * -DestinationPath "${archivedir}/${rid}.zip"
}

cd $workingdir
