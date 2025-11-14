#!/bin/bash
set -e

echo "🔧 REGENERACIÓN PROFESIONAL DE PROTOBUF"

cd ~/Desktop/Servidor/osmi/osmi-server

# Limpiar completamente
echo "🗑️  Limpiando generaciones anteriores..."
rm -rf gen/
mkdir -p gen

# Verificar estructura
echo "📁 Verificando estructura..."
if [ ! -f "internal/proto/osmi.proto" ]; then
    echo "❌ ERROR: internal/proto/osmi.proto no encontrado"
    echo "📋 Por favor ejecuta los pasos anteriores primero"
    exit 1
fi

# Regenerar con configuración profesional
echo "🔄 Generando archivos protobuf..."
protoc \
  --proto_path=internal/proto \
  --proto_path=proto/googleapis \
  --go_out=gen \
  --go_opt=paths=source_relative \
  --go-grpc_out=gen \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=gen \
  --grpc-gateway_opt=paths=source_relative \
  --grpc-gateway_opt=logtostderr=true \
  internal/proto/osmi.proto

# Verificar resultados
echo "✅ Verificando generación..."
if [ -f "gen/osmi.pb.go" ]; then
    echo "🎉 ÉXITO: Archivos generados correctamente en gen/"
    ls -la gen/
else
    echo "❌ FALLA: No se generaron archivos en gen/"
    echo "🔍 Buscando archivos generados..."
    find . -name "*.pb.go" -type f
    exit 1
fi

echo "📦 Sincronizando módulos..."
go mod tidy

echo "🚀 REGENERACIÓN COMPLETADA - Ahora compila: go run cmd/main.go"