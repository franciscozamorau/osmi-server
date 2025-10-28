@echo off
echo ========================================
echo    GENERANDO CODIGO PROTOBUF
echo ========================================

echo 1. Limpiando generaciones anteriores...
rmdir /s /q gen 2>nul
mkdir gen

echo 2. Generando para osmi-server...
protoc -Iproto -Ithird_party ^
  --go_out=gen --go_opt=paths=source_relative ^
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative ^
  proto\osmi.proto

if %errorlevel% neq 0 (
    echo Error generando para osmi-server
    pause
    exit /b %errorlevel%
)

echo 3. Generando para osmi-gateway...
cd ..\osmi-gateway
rmdir /s /q gen 2>nul
mkdir gen

protoc -I..\osmi-server\proto -I..\osmi-server\third_party ^
  --go_out=gen --go_opt=paths=source_relative ^
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative ^
  --grpc-gateway_out=gen --grpc-gateway_opt=paths=source_relative ^
  ..\osmi-server\proto\osmi.proto

if %errorlevel% neq 0 (
    echo Error generando para osmi-gateway
    pause
    exit /b %errorlevel%
)

echo 4. Volviendo a osmi-server...
cd ..\osmi-server

echo ========================================
echo    CODIGO GENERADO EXITOSAMENTE
echo ========================================
echo Archivos generados:
dir gen /B
echo.
pause