$ErrorActionPreference = "Stop"
$repository = if ($env:LAGO_INSTALL_REPOSITORY) { $env:LAGO_INSTALL_REPOSITORY } else { "getlago/lago-cli" }
$version = if ($env:LAGO_INSTALL_VERSION) { $env:LAGO_INSTALL_VERSION.TrimStart("v") } else { (Invoke-RestMethod "https://api.github.com/repos/$repository/releases/latest").tag_name.TrimStart("v") }
$architecture = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq "Arm64") { "arm64" } else { "amd64" }
$archive = "lago_${version}_windows_${architecture}.zip"
$base = "https://github.com/$repository/releases/download/v$version"
$temporary = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $temporary | Out-Null
try {
  Invoke-WebRequest "$base/$archive" -OutFile (Join-Path $temporary $archive)
  Invoke-WebRequest "$base/checksums.txt" -OutFile (Join-Path $temporary "checksums.txt")
  $expected = ((Get-Content (Join-Path $temporary "checksums.txt")) | Where-Object { $_ -match "\s+$([regex]::Escape($archive))$" }) -split "\s+" | Select-Object -First 1
  $actual = (Get-FileHash (Join-Path $temporary $archive) -Algorithm SHA256).Hash.ToLowerInvariant()
  if ($actual -ne $expected.ToLowerInvariant()) { throw "Checksum verification failed" }
  Expand-Archive (Join-Path $temporary $archive) -DestinationPath $temporary
  $installDirectory = if ($env:LAGO_INSTALL_DIR) { $env:LAGO_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "Lago\bin" }
  New-Item -ItemType Directory -Force -Path $installDirectory | Out-Null
  Copy-Item (Join-Path $temporary "lago.exe") (Join-Path $installDirectory "lago.exe") -Force
  Write-Host "Installed lago $version to $installDirectory\lago.exe"
} finally {
  Remove-Item -Recurse -Force $temporary
}
