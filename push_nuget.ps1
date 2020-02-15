$vars = "NUGET_API_KEY", "GITHUB_SOURCE", "GITHUB_USERNAME", "GITHUB_ACCESS_TOKEN"

$vars | ForEach-Object {
    if (!(Test-Path Env:$_)) {
        Write-Host "$_ environment variable is missing."
        exit 1
    }
}


dir ./*/.nupkg/*.nupkg | ForEach-Object {
    $package = $_.FullName
    nuget push $package $env:NUGET_API_KEY -Source https://api.nuget.org/v3/index.json
    nuget source Add -Name "GitHub" -Source $env:GITHUB_SOURCE -UserName $env:GITHUB_USERNAME -Password $env:GITHUB_ACCESS_TOKEN
    nuget push -Source "GitHub" $package
}
