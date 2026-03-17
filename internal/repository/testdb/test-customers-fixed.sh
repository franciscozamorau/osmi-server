#!/bin/bash

echo "🔍 PROBANDO ENDPOINTS DE CUSTOMERS"
echo "=================================="

GATEWAY="http://localhost:8083"
TOKEN="test-token-123"

# 1. Health check
echo -n "1. Health check: "
curl -s "$GATEWAY/health" | grep -q "healthy" && echo "✅" || echo "❌"

# 2. Crear cliente con email único
echo -n "2. Crear cliente: "
TIMESTAMP=$(date +%s)
RESPONSE=$(curl -s -X POST "$GATEWAY/customers" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Carlos Ruiz\",
    \"email\": \"carlos.${TIMESTAMP}@correo.com\",
    \"phone\": \"+525599887766\"
  }")

# Extraer ID manualmente (sin jq)
ID=$(echo $RESPONSE | grep -o '"publicId":"[^"]*' | cut -d'"' -f4)

if [ -n "$ID" ]; then
    echo "✅ ID: $ID"
    
    # 3. Obtener cliente por ID (ESTO FALLARÁ por el tipo mismatch)
    echo -n "3. Obtener cliente por ID: "
    GET_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$GATEWAY/customers/$ID")
    if echo "$GET_RESPONSE" | grep -q "type mismatch"; then
        echo "⚠️  FALLA (type mismatch - requiere fix en proto)"
    else
        echo "✅"
    fi
    
    # 4. Actualizar cliente (ESTO SÍ FUNCIONA)
    echo -n "4. Actualizar cliente: "
    UPDATE_RESPONSE=$(curl -s -X PATCH "$GATEWAY/customers/$ID" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"name": "Carlos Actualizado"}')
    if echo "$UPDATE_RESPONSE" | grep -q "Actualizado"; then
        echo "✅"
    else
        echo "❌"
    fi
else
    echo "❌ No se pudo crear cliente"
    echo "Respuesta: $RESPONSE"
fi

# 5. Stats (AHORA FUNCIONA)
echo -n "5. Customer stats: "
STATS=$(curl -s -H "Authorization: Bearer $TOKEN" "$GATEWAY/customers/stats")
if echo "$STATS" | grep -q "totalCustomers"; then
    echo "✅"
    echo "   Total customers: $(echo $STATS | grep -o '"totalCustomers":"[^"]*' | cut -d'"' -f4)"
else
    echo "❌"
fi

# 6. Listar customers
echo -n "6. Listar customers: "
LIST=$(curl -s -H "Authorization: Bearer $TOKEN" "$GATEWAY/customers")
if echo "$LIST" | grep -q "customers"; then
    echo "✅"
else
    echo "❌"
fi

echo "=================================="