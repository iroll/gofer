@echo off
setlocal

echo Installing gofer...

REM === 1. Create target folder ===
set TARGET=%LOCALAPPDATA%\gofer
if not exist "%TARGET%" mkdir "%TARGET%"

echo Copying files to %TARGET%...
copy /Y "%~dp0gofer*" "%TARGET%\" >nul
copy /Y "%~dp0register-gopher.ps1" "%TARGET%\" >nul

REM === 2. Run the protocol registration script ===
echo Registering gopher:// protocol...
powershell -NoProfile -ExecutionPolicy Bypass -File "%TARGET%\register-gopher.ps1"

REM === 3. Create desktop shortcut ===
echo Creating desktop shortcut...

set DESKTOP=%USERPROFILE%\Desktop
set SHORTCUT=%DESKTOP%\gofer.lnk
set EXE=%TARGET%\gofer-wrapper.exe

powershell -NoProfile -ExecutionPolicy Bypass ^
  -Command "$s=(New-Object -ComObject WScript.Shell).CreateShortcut('%SHORTCUT%');" ^
           "$s.TargetPath='%EXE%';" ^
           "$s.WorkingDirectory='%TARGET%';" ^
           "$s.Save()"

echo.
echo Installation complete.
echo Shortcut created on your desktop.
echo Files installed to: %TARGET%
echo.

endlocal
