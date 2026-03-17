# ESTADO DE ENDPOINTS DE CUSTOMERS - 17/03/2026

## ✅ ENDPOINTS FUNCIONALES

| Endpoint | Método | Estado | Notas |
|----------|--------|--------|-------|
| `/health` | GET | ✅ Funciona | Público, sin auth |
| `/customers` | POST | ✅ Funciona | Crear cliente con email único |
| `/customers` | GET | ✅ Funciona | Listar clientes (con paginación) |
| `/customers/{id}` | PATCH | ✅ Funciona | Actualizar cliente por ID |
| `/customers/stats` | GET | ✅ Funciona | Estadísticas de clientes |

## ⚠️ ENDPOINTS CON PROBLEMAS

| Endpoint | Método | Estado | Problema | Solución |
|----------|--------|--------|----------|----------|
| `/customers/{id}` | GET | ❌ Falla | Type mismatch: espera int64, recibe UUID | Cambiar proto de `int64` a `string` |

## 🔧 ACCIONES REQUERIDAS

### ALTA PRIORIDAD
1. **Arreglar GET /customers/{id}**
   - Archivo: `osmi-protobuf/proto/customer.proto`
   - Cambiar `int64 id = 1` → `string id = 1`
   - Regenerar protos: `make generate-proto`
   - Recompilar server y gateway

### BAJA PRIORIDAD
2. **Mejorar manejo de errores**
   - Mensajes más amigables para el cliente
   - Documentar códigos de error

## 📊 MÉTRICAS ACTUALES
- Total customers: 10
- Clientes activos: 10
- Clientes nuevos (30 días): 10