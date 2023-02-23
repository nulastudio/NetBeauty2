param(
    $Certificate,
    $Algorithm,
    $TimeStampServer,
    [string[]]
    [Parameter(Position=0, ValueFromRemainingArguments)]
    $Files
    )

$signingParameters = @{
    Force = $true
}

if ($Algorithm) {
    $signingParameters["HashAlgorithm"] = $Algorithm
}

if ($TimeStampServer) {
    $signingParameters["TimeStampServer"] = $TimeStampServer
}

Write-Host ""

if ($Files -eq $null -or $Files.Count -eq 0) {
    Write-Host "No Files To Sign."
    exit
}

$signingParameters["FilePath"] = $Files | Select-Object -Unique

if ($Certificate -ne "Auto" -and $Certificate -ne "Prompt") {
    $Certificate = "Auto"
}

if ($Certificate -eq "Auto" -or $Certificate -eq "Prompt") {
    $title = "Available Signing Certificates"
    $text = "Please Choose One Certificate For Signing:"

    $certificates = Get-ChildItem -Path Cert: -Recurse -CodeSigningCert

    [void][Reflection.Assembly]::LoadWithPartialName("System.Security")

    $collection = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2Collection

    $certificates | ForEach-Object { [void]$collection.Add($_) }

    $cert = $null

    if ($collection.Count -ne 0) {
        if ($Certificate -eq "Auto" -and $collection.Count -eq 1) {
            $cert = $collection[0]
        } else {
            $cert = [System.Security.Cryptography.x509Certificates.X509Certificate2UI]::SelectFromCollection($collection, $title, $text, 0)
            if ($cert) {
                $cert = $cert[0]
            }
        }
    }

    if (!$cert) {
        Write-Host "No Available Certificates For Signing."
        exit
    }

    $signingParameters["Certificate"] = $cert
} else {
    $signingParameters["Certificate"] = Get-PfxCertificate -FilePath $Certificate
}

Write-Host "Signing Files..."

Set-AuthenticodeSignature @signingParameters | Select-Object -Property Status, StatusMessage, Path | Format-List
