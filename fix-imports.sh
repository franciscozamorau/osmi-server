#!/bin/bash
# Script profesional para corregir imports inconsistentes

set -e  # Salir ante cualquier error

echo "🔧 CORRECCIÓN PROFESIONAL DE IMPORTS"
echo "===================================="
echo "OSMI Server - Arquitectura Limpia"
echo ""

# 1. Encontrar TODOS los archivos con import incorrecto
echo "📋 Buscando imports a corregir..."
INCORRECT_IMPORTS=$(grep -r '"github.com/franciscozamorau/osmi-server/internal/api/dto"' internal/ --include="*.go" | wc -l)
echo "   Se encontraron $INCORRECT_IMPORTS imports incorrectos"

# 2. Corregir imports en handlers
echo ""
echo "🔄 Corrigiendo handlers..."

# Ticket Handler
if [ -f "internal/application/handlers/grpc/ticket_handler.go" ]; then
    echo "  📝 Corrigiendo ticket_handler.go"
    sed -i.bak '
        s|"github.com/franciscozamorau/osmi-server/internal/api/dto"|"github.com/franciscozamorau/osmi-server/internal/api/dto/request"\n\t"github.com/franciscozamorau/osmi-server/internal/api/dto/response"|g
    ' internal/application/handlers/grpc/ticket_handler.go
    
    # Cambiar dto. -> request. o response. según corresponda
    sed -i.bak '
        s/dto\.CreateTicketRequest/request.CreateTicketRequest/g
        s/dto\.UpdateTicketStatusRequest/request.UpdateTicketStatusRequest/g
        s/dto\.TicketResponse/response.TicketResponse/g
        s/dto\.TicketListResponse/response.TicketListResponse/g
    ' internal/application/handlers/grpc/ticket_handler.go
    
    rm -f internal/application/handlers/grpc/ticket_handler.go.bak
fi

# Customer Handler (ya lo corregimos, pero verificar)
if [ -f "internal/application/handlers/grpc/customer_handler.go" ]; then
    echo "  📝 Verificando customer_handler.go"
    if grep -q '"github.com/franciscozamorau/osmi-server/internal/api/dto"' internal/application/handlers/grpc/customer_handler.go; then
        sed -i.bak '
            s|"github.com/franciscozamorau/osmi-server/internal/api/dto"|"github.com/franciscozamorau/osmi-server/internal/api/dto/request"\n\t"github.com/franciscozamorau/osmi-server/internal/api/dto/response"|g
        ' internal/application/handlers/grpc/customer_handler.go
        rm -f internal/application/handlers/grpc/customer_handler.go.bak
    fi
fi

# 3. Corregir imports en services
echo ""
echo "🔄 Corrigiendo services..."

# Customer Service
if [ -f "internal/application/services/customer_service.go" ]; then
    echo "  📝 Corrigiendo customer_service.go"
    sed -i.bak '
        s|"github.com/franciscozamorau/osmi-server/internal/api/dto"|"github.com/franciscozamorau/osmi-server/internal/api/dto/request"\n\t"github.com/franciscozamorau/osmi-server/internal/api/dto/response"|g
        s/dto\.CreateCustomerRequest/request.CreateCustomerRequest/g
        s/dto\.UpdateCustomerRequest/request.UpdateCustomerRequest/g
        s/dto\.CustomerFilter/request.CustomerFilter/g
        s/dto\.Pagination/request.Pagination/g
        s/dto\.CustomerStatsResponse/response.CustomerStatsResponse/g
    ' internal/application/services/customer_service.go
    rm -f internal/application/services/customer_service.go.bak
fi

# 4. Corregir imports en repositories (si existen)
echo ""
echo "🔄 Corrigiendo repositories..."
find internal/infrastructure/repositories -name "*.go" -type f | while read file; do
    if grep -q '"github.com/franciscozamorau/osmi-server/internal/api/dto"' "$file"; then
        echo "  📝 Corrigiendo $(basename "$file")"
        sed -i.bak '
            s|"github.com/franciscozamorau/osmi-server/internal/api/dto"|"github.com/franciscozamorau/osmi-server/internal/api/dto/request"\n\t"github.com/franciscozamorau/osmi-server/internal/api/dto/response"|g
            s/dto\.CustomerFilter/request.CustomerFilter/g
            s/dto\.Pagination/request.Pagination/g
        ' "$file"
        rm -f "$file.bak"
    fi
done

# 5. Crear un archivo dto.go raíz para compatibilidad temporal
echo ""
echo "🔗 Creando puente de compatibilidad..."
cat > internal/api/dto/dto.go << 'EOF'
// dto.go - Archivo puente para compatibilidad
// Este archivo permite que imports antiguos sigan funcionando
// mientras migramos a la nueva estructura
package dto

// Re-exportar tipos de request
type CreateCustomerRequest = request.CreateCustomerRequest
type UpdateCustomerRequest = request.UpdateCustomerRequest
type CustomerFilter = request.CustomerFilter
type CreateTicketRequest = request.CreateTicketRequest
type UpdateTicketStatusRequest = request.UpdateTicketStatusRequest

// Re-exportar tipos de response
type CustomerResponse = response.CustomerResponse
type CustomerStatsResponse = response.CustomerStatsResponse
type TicketResponse = response.TicketResponse
type TicketListResponse = response.TicketListResponse

// Re-exportar tipos de filter
type Pagination = filter.Pagination
EOF

echo "  ✅ Archivo puente creado: internal/api/dto/dto.go"

# 6. Verificar correcciones
echo ""
echo "🔍 Verificando correcciones..."
REMAINING=$(grep -r '"github.com/franciscozamorau/osmi-server/internal/api/dto"' internal/ --include="*.go" | grep -v "dto.go" | wc -l)

if [ "$REMAINING" -eq 0 ]; then
    echo "✅ ¡TODOS los imports han sido corregidos!"
else
    echo "⚠️  Aún quedan $REMAINING imports por corregir manualmente:"
    grep -r '"github.com/franciscozamorau/osmi-server/internal/api/dto"' internal/ --include="*.go" | grep -v "dto.go"
fi

echo ""
echo "🎯 PASO FINAL: Prueba de compilación"
echo "====================================="
echo "Ejecuta estos comandos:"
echo ""
echo "1. Verificar estructura:"
echo "   find internal/api/dto -name \"*.go\" | head -20"
echo ""
echo "2. Probar compilación:"
echo "   go build ./..."
echo ""
echo "3. Si hay errores, corregir archivo por archivo:"
echo "   go build -v ./... 2>&1 | grep \"cannot find package\""
echo ""
echo "✅ Corrección profesional completada"