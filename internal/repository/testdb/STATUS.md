# ESTADO DE ENDPOINTS - CUSTOMERS

| Endpoint | Método | Estado | Error | Solución |
|----------|--------|--------|-------|----------|
| /customers | GET | ✅ Funciona | - | - |
| /customers | POST | ✅ Funciona | - | - |
| /customers/{id} | GET | ✅ Funciona | Usar ID real, no "{id}" | Documentación |
| /customers/{id} | PATCH | ✅ Funciona | Usar ID real, no "{id}" | Documentación |
| /customers/stats | GET | ❌ Falla | missing destination name total_customers | Arreglar DTO en server |

## ACCIONES REQUERIDAS

1. **INMEDIATO**: Arreglar CustomerStats en server
   - Archivo: `osmi-server/internal/infrastructure/repositories/postgres/customer_repository.go`
   - Agregar campo `TotalCustomers int64` con tag `db:"total_customers"`
   - Actualizar query SELECT para incluir el campo

2. **DOCUMENTACIÓN**: Actualizar ejemplos para usar IDs reales
   - Siempre guardar ID después de POST
   - Usar variables en scripts