@echo off
cd /d "%~dp0"
set GOOS=wasip1
set GOARCH=wasm
go build -o ..\dist\openapi-go-gin-mock-v1.0.0.wasm ..\cmd\plugin\main.go
