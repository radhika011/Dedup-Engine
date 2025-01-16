@echo off
cd "%ProgramFiles%\backup"

.\servicemgr -service stop
.\servicemgr -service uninstall

cd ..

rmdir /s /q "%ProgramFiles%\backup"