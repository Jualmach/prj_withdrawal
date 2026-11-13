param(
    [string]$OutputPath = (Join-Path (Split-Path -Parent $PSScriptRoot) 'prj_withdrawal-source.zip'),
    [switch]$Force
)

$projectRoot = Split-Path -Parent $PSScriptRoot
$resolvedOutput = [System.IO.Path]::GetFullPath($OutputPath)
$rootPrefix = [System.IO.Path]::GetFullPath($projectRoot).TrimEnd('\') + '\'

if (Test-Path -LiteralPath $resolvedOutput) {
    if (-not $Force) {
        throw "El archivo de salida ya existe; use -Force para reemplazarlo: $resolvedOutput"
    }
    Remove-Item -LiteralPath $resolvedOutput
}

Add-Type -AssemblyName System.IO.Compression.FileSystem
Add-Type -AssemblyName System.IO.Compression
$archive = [System.IO.Compression.ZipFile]::Open(
    $resolvedOutput,
    [System.IO.Compression.ZipArchiveMode]::Create
)

try {
    Get-ChildItem -LiteralPath $projectRoot -File -Recurse |
        Where-Object {
            $_.FullName -notmatch '[\\/]\.git([\\/]|$)' -and
            $_.Name -ne 'state.json' -and
            $_.Extension -ne '.exe' -and
            $_.Extension -ne '.test' -and
            $_.FullName -ne $resolvedOutput
        } |
        ForEach-Object {
            $entryName = $_.FullName.Substring($rootPrefix.Length).Replace('\', '/')
            [System.IO.Compression.ZipFileExtensions]::CreateEntryFromFile(
                $archive,
                $_.FullName,
                $entryName,
                [System.IO.Compression.CompressionLevel]::Optimal
            ) | Out-Null
        }
}
finally {
    $archive.Dispose()
}

Write-Output "ZIP creado: $resolvedOutput"
