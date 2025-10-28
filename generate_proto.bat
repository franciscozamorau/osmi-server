@echo off
echo ========================================
echo    GENERANDO CODIGO PROTOBUF
echo ========================================

echo 1. Verificando estructura de third_party...
dir third_party\googleapis\google\api 2>nul
if %errorlevel% neq 0 (
    echo ERROR: No se encontraron las dependencias de Google APIs
    echo Ejecuta: cd third_party && git clone https://github.com/googleapis/googleapis
    pause
    exit /b 1
)

echo 2. Limpiando generaciones anteriores...
rmdir /s /q gen 2>nul
mkdir gen

echo 3. Generando para osmi-server...
protoc -Iproto -Ithird_party\googleapis ^
  --go_out=gen --go_opt=paths=source_relative ^
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative ^
  proto\osmi.proto

if %errorlevel% neq 0 (
    echo ERROR generando para osmi-server
    pause
    exit /b %errorlevel%
)

echo 4. Generando gRPC-Gateway para osmi-server...
protoc -Iproto -Ithird_party\googleapis ^
  --grpc-gateway_out=gen ^
  --grpc-gateway_opt=paths=source_relative ^
  --grpc-gateway_opt=generate_unbound_methods=true ^
  proto\osmi.proto

if %errorlevel% neq 0 (
    echo ERROR generando gRPC-Gateway para osmi-server
    pause
    exit /b %errorlevel%
)

echo 5. Generando para osmi-gateway...
cd ..\osmi-gateway
rmdir /s /q gen 2>nul
mkdir gen

protoc -I..\osmi-server\proto -I..\osmi-server\third_party\googleapis ^
  --go_out=gen --go_opt=paths=source_relative ^
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative ^
  --grpc-gateway_out=gen ^
  --grpc-gateway_opt=paths=source_relative ^
  --grpc-gateway_opt=generate_unbound_methods=true ^
  ..\osmi-server\proto\osmi.proto

if %errorlevel% neq 0 (
    echo ERROR generando para osmi-gateway
    pause
    exit /b %errorlevel%
)

echo 6. Volviendo a osmi-server...
cd ..\osmi-server

echo ========================================
echo    CODIGO GENERADO EXITOSAMENTE
echo ========================================
echo Archivos generados en osmi-server\gen:
dir gen /B
echo.
echo Archivos generados en osmi-gateway\gen:
dir ..\osmi-gateway\gen /B
echo.
pause