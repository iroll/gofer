@echo off
setlocal enabledelayedexpansion

REM ============================================================
REM gofer — Windows WPF wrapper build script
REM Target: .NET 8, win-x64
REM ============================================================

echo.
echo === gofer-wrapper: restore ===
cd gofer-wrapper || exit /b 1
dotnet restore || exit /b 1

echo.
echo === gofer-wrapper: build (Release) ===
dotnet build -c Release || exit /b 1

echo.
echo === gofer-wrapper: publish (win-x64) ===
dotnet publish ^
  -c Release ^
  -r win-x64 ^
  --self-contained false ^
  -p:UseWPF=true || exit /b 1

echo.
echo === done ===
echo Output:
echo   gofer-wrapper\bin\Release\net8.0-windows\win-x64\publish\
echo.

endlocal
