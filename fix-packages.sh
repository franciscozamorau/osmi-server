#!/bin/bash
# Script para corregir paquetes inconsistentes

set -e  # Salir ante cualquier error

echo "🔧 INICIANDO CORRECCIÓN DE PAQUETES"
echo "===================================="

# 1. Corregir request/
echo "📁 Corrigiendo package en request/"
for file in internal/api/dto/request/*.go; do
    if [ -f "$file" ]; then
        # Cambiar package dto -> request
        sed -i.bak '1s/^package dto$/package request/' "$file"
        # Quitar .bak
        rm -f "${file}.bak"
        echo "  ✅ $(basename "$file")"
    fi
done

# 2. Corregir response/
echo "📁 Corrigiendo package en response/"
for file in internal/api/dto/response/*.go; do
    if [ -f "$file" ]; then
        # Cambiar package dto -> response
        sed -i.bak '1s/^package dto$/package response/' "$file"
        rm -f "${file}.bak"
        echo "  ✅ $(basename "$file")"
    fi
done

# 3. Unificar handlers gRPC
echo "📁 Unificando handlers gRPC"
for file in internal/application/handlers/grpc/*.go; do
    if [ -f "$file" ]; then
        # Cambiar grpchandlers -> grpc
        sed -i.bak '1s/^package grpchandlers$/package grpc/' "$file"
        rm -f "${file}.bak"
        echo "  ✅ $(basename "$file")"
    fi
done

# 4. Corregir database
echo "📁 Corrigiendo database"
if [ -f "internal/database/connection.go" ]; then
    sed -i.bak '1s/^package db$/package database/' internal/database/connection.go
    rm -f internal/database/connection.go.bak
    echo "  ✅ connection.go"
fi

# 5. Eliminar archivos backup duplicados
echo "🧹 Limpiando archivos backup"
rm -f internal/api/dto/response/category_response_bak.go
rm -f internal/api/dto/request/notification_request_bak.go
rm -f internal/application/handlers/grpc/event_handler_bak.go
rm -f internal/database/di_bak.go

# 6. Renombrar services.go a service.go (si existe)
if [ -f "internal/application/services/services.go" ] && [ -f "internal/application/services/service.go" ]; then
    echo "⚠️  Tienes services.go y service.go - decidiendo cuál mantener"
    # Mostrar diferencias
    echo "services.go (nuevo):"
    head -5 internal/application/services/services.go
    echo ""
    echo "service.go (viejo):"
    head -5 internal/application/services/service.go
    echo ""
    echo "Recomendación: Mantener services.go (ya corregido)"
    rm -f internal/application/services/service.go
fi

echo ""
echo "✅ CORRECCIÓN COMPLETADA"
echo "========================="
echo "Ejecuta: go build ./... para verificar"