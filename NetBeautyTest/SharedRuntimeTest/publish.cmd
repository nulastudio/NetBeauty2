@echo off

cd WsServer
dotnet publish -r win-x64 -c Release

cd ..

cd WsClient
dotnet publish -r win-x64 -c Release

cd ..

pause
