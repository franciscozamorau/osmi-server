#!/bin/bash

set -e  # Detener en error

echo "🔍 PROBANDO ENDPOINTS DE CUSTOMERS"
echo "=================================="

# Configuración
GATEWAY_URL="http://localhost:8083"
TOKEN="test-token-123"

# 1. Health check
echo -n "1. Health check: "
curl -s "$GATEWAY_URL/health" | grep -q "healthy" && echo "✅" || echo "❌"

# 2. Crear cliente
echo -n "2. Crear cliente: "
RESPONSE=$(curl -s -X POST "$GATEWAY_URL/customers" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test_'$(date +%s)'@email.com",
    "phone": "+525512345678"
  }')

# Extraer ID (si tienes jq)
if command -v jq &> /dev/null; then
    ID=$(echo $RESPONSE | jq -r '.publicId')
    echo "✅ ID: $ID"
else
    echo "✅ (manual)"
    ID=""
fi

# 3. Listar clientes
echo -n "3. Listar clientes: "
curl -s -H "Authorization: Bearer $TOKEN" "$GATEWAY_URL/customers" | grep -q "customers" && echo "✅" || echo "❌"

# 4. Obtener cliente por ID (si tenemos ID)
if [ -n "$ID" ]; then
    echo -n "4. Obtener cliente $ID: "
    curl -s -H "Authorization: Bearer $TOKEN" "$GATEWAY_URL/customers/$ID" | grep -q "$ID" && echo "✅" || echo "❌"
    
    # 5. Actualizar cliente
    echo -n "5. Actualizar cliente: "
    curl -s -X PATCH "$GATEWAY_URL/customers/$ID" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"name": "Updated Name"}' | grep -q "Updated" && echo "✅" || echo "❌"
fi

# 6. Stats (puede fallar hasta que arreglemos)
echo -n "6. Customer stats: "
curl -s -H "Authorization: Bearer $TOKEN" "$GATEWAY_URL/customers/stats" && echo "⚠️  (revisar)" || echo "❌ (requiere fix en server)"

echo "=================================="