$rootdir = $pwd
$tooldir = "tools"

function Archive($rid)
{
    cd "${rootdir}/${tooldir}/${rid}"
    Compress-Archive -Force -Path * -DestinationPath "${rootdir}/${tooldir}/${rid}.zip"
    cd "${rootdir}"
}

$rids = "win-x86", "win-x64", "linux-x64", "osx-x64"

foreach ($rid in $rids)
{
    Archive -rid $rid
}
