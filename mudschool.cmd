@echo off
rem Builds and runs the course tool on Windows (cmd.exe and PowerShell):
rem   .\mudschool.cmd <command>
setlocal
set "MUD_ROOT=%~dp0"
pushd "%MUD_ROOT%tools\mudschool" || exit /b 1
go build -o "%MUD_ROOT%.bin\mudschool.exe" .
if errorlevel 1 (
  popd
  exit /b 1
)
popd
"%MUD_ROOT%.bin\mudschool.exe" %*
exit /b %errorlevel%
