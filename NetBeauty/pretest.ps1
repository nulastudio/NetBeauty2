function Get-Separator
{
    return (Join-Path . .).Trim('.')
}

function Format-Path($path)
{
    $separator = Get-Separator
    return "${path}".Replace('/', $separator).Replace('\', $separator)
}

$scriptdir = Format-Path (Split-Path -Parent $MyInvocation.MyCommand.Definition)
$testdatadir = Format-Path "${scriptdir}/test_data"
$testdir = Format-Path "${scriptdir}/test"

if (!(Test-Path $testdatadir)) {
    mkdir -p $testdatadir >$null 2>$null
}

if (!(Test-Path $testdir)) {
    mkdir -p $testdir >$null 2>$null
} else {
    Remove-Item -Recurse -Force "${testdir}/*" >$null 2>$null
}

Copy-Item -Recurse -Force -Path "${testdatadir}/*" -Destination $testdir >$null 2>$null
