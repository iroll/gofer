$exePath = (Resolve-Path ".\gofer-wrapper.exe").Path

$base = "HKCU:\Software\Classes\gopher"

New-Item -Path $base -Force | Out-Null
Set-ItemProperty -Path $base -Name "(Default)" -Value "URL:Gopher Protocol"
Set-ItemProperty -Path $base -Name "URL Protocol" -Value ""

New-Item -Path "$base\shell\open\command" -Force | Out-Null
Set-ItemProperty `
  -Path "$base\shell\open\command" `
  -Name "(Default)" `
  -Value "`"$exePath`" `"%1`""

Write-Host "gopher:// protocol registered for:"
Write-Host $exePath
