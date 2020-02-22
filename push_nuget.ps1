$vars = "NUGET_API_KEY", "GITHUB_SOURCE", "GITHUB_USERNAME", "GITHUB_ACCESS_TOKEN"

$vars | ForEach-Object {
    if (!(Test-Path Env:$_)) {
        Write-Host "$_ environment variable is missing."
        exit 1
    }
}

nuget source Add -Name "NuGet" -Source https://api.nuget.org/v3/index.json
nuget source Add -Name "GitHub" -Source $env:GITHUB_SOURCE -UserName $env:GITHUB_USERNAME -Password $env:GITHUB_ACCESS_TOKEN
nuget setApiKey $env:NUGET_API_KEY -Source "NuGet"

$pwd = Split-Path -Parent $MyInvocation.MyCommand.Definition

dir "$pwd/*/.nupkg/*.nupkg" | ForEach-Object {
    $package = $_ -Replace $pwd, ""
    nuget push "$package" -Source "NuGet"
    nuget push "$package" -Source "GitHub"
}
