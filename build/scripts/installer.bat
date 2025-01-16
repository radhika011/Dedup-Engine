@echo off

mkdir "%ProgramFiles%\backup"

copy "%CD%\bkupClient.exe" "%ProgramFiles%\backup"
copy "%CD%\client-background.exe" "%ProgramFiles%\backup"
copy "%CD%\.env.common" "%ProgramFiles%\backup"
copy "%CD%\.env.background" "%ProgramFiles%\backup"
copy "%CD%\servicemgr.exe" "%ProgramFiles%\backup"

cd "%ProgramFiles%\backup"

.\servicemgr -service install
.\servicemgr -service start

icacls "%ProgramFiles%\backup" /grant "%USERNAME%":(OI)(CI)F /T