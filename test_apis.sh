#!/bin/bash
set -e

echo "í·ª PROBANDO APIS OSMI"
echo "======================"

# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# FunciÃ³n para probar endpoint
test_endpoint() {
    local name=$1
    local url=$2
    local method=${3:-GET}
    local data=${4:-}
    
    echo -n "Testing $name ($method $url)... "
    
    if [ "$method" = "POST" ] && [ -n "$data" ]; then
        response=$(curl -s -X "$method" -H "Content-Type: application/json" -d "$data" "$url" || echo "ERROR")
    else
        response=$(curl -s -X "$method" "$url" || echo "ERROR")
    fi
    
    if echo "$response" | grep -q "ERROR"; then
        echo -e "${RED}FAILED${NC}"
        echo "  Error: No se pudo conectar"
        return 1
    elif echo "$response" | grep -q '"status":' || echo "$response" | grep -q '"message":'; then
        echo -e "${GREEN}OK${NC}"
        echo "  Response: $(echo $response | tr -d '\n' | cut -c1-80)..."
        return 0
    else
        echo -e "${RED}FAILED${NC}"
        echo "  Response: $response"
        return 1
    fi
}

# Iniciar servidor si no estÃ¡ corriendo
echo "1. Verificando servidor..."
if ! curl -s http://localhost:8081/health > /dev/null 2>&1; then
    echo "Servidor no estÃ¡ corriendo. Iniciando..."
    go run cmd/main_test.go &
    SERVER_PID=$!
    sleep 3
fi

# Esperar a que el servidor estÃ© listo
echo "2. Esperando que el servidor estÃ© listo..."
for i in {1..10}; do
    if curl -s http://localhost:8081/health > /dev/null 2>&1; then
        echo "âœ… Servidor listo"
        break
    fi
    echo "   Intento $i/10..."
    sleep 2
done

# Probar endpoints
echo ""
echo "3. Probando endpoints..."
echo "-----------------------"

# Health endpoints
test_endpoint "Health Check" "http://localhost:8081/health"
test_endpoint "Ready Check" "http://localhost:8081/ready"

# API endpoints
test_endpoint "Test API" "http://localhost:8080/api/v1/test"
test_endpoint "Get Tickets" "http://localhost:8080/api/v1/tickets"

# Probar POST
test_endpoint "Create Ticket" "http://localhost:8080/api/v1/tickets" "POST" '{"event_id":"evt_123","customer_id":"cust_456","quantity":2}'

# Probar eventos
test_endpoint "Get Events" "http://localhost:8080/api/v1/events"

echo ""
echo "í³Š RESUMEN DE PRUEBAS"
echo "====================="
echo "Para probar con Postman:"
echo ""
echo "1. Health Check:"
echo "   GET http://localhost:8081/health"
echo ""
echo "2. Test API:"
echo "   GET http://localhost:8080/api/v1/test"
echo ""
echo "3. Tickets API:"
echo "   GET  http://localhost:8080/api/v1/tickets"
echo "   POST http://localhost:8080/api/v1/tickets"
echo "   Body: {\"event_id\": \"evt_123\", \"customer_id\": \"cust_456\", \"quantity\": 2}"
echo ""
echo "4. Events API:"
echo "   GET http://localhost:8080/api/v1/events"
echo ""
echo "í³ EJEMPLOS PARA POSTMAN:"
echo "========================="
cat > postman_examples.json << 'EOJ'
{
  "info": {
    "name": "OSMI API Examples",
    "description": "Ejemplos para probar en Postman"
  },
  "item": [
    {
      "name": "Health Check",
      "request": {
        "method": "GET",
        "url": "http://localhost:8081/health"
      }
    },
    {
      "name": "Create Ticket",
      "request": {
        "method": "POST",
        "url": "http://localhost:8080/api/v1/tickets",
        "body": {
          "mode": "raw",
          "raw": "{\n  \"event_id\": \"evt_123\",\n  \"customer_id\": \"cust_456\",\n  \"quantity\": 2\n}"
        },
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ]
      }
    }
  ]
}
EOJ

echo "Ejemplos guardados en: postman_examples.json"
echo ""
echo "í¾¯ Para detener el servidor:"
echo "   kill $SERVER_PID 2>/dev/null || echo 'Servidor detenido'"
