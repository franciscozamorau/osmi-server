# Osmi Server
Backend gRPC para la plataforma Osmi. Este módulo implementa el núcleo del sistema de boletaje digital utilizando una arquitectura escalable, segura y profesional. Incluye servicios gRPC completos, integración con PostgreSQL, y health checks.
---

# Osmi Core Stack
```
Go → lenguaje principal
gRPC → protocolo de comunicación entre servicios
Protobuf → definición de contratos y mensajes
grpc-gateway → puente REST ↔ gRPC (activo)
PostgreSQL → base de datos relacional (conectado)
Kubernetes → orquestación y despliegue (en proceso)
Docker → contenedorización del servicio
.env + godotenv → gestión de variables de entorno
Health & Readiness Probes → verificación de estado del sistema
```

# Estructura del proyecto
```bash
osmi-server/
├── cmd/
│   └── main.go                 # Entrypoint principal del servidor gRPC
├── proto/
│   └── osmi.proto             # Definición de servicios gRPC con anotaciones REST
├── gen/                       # Código generado por protoc (NO EDITAR)
│   ├── osmi.pb.go             # Estructuras de datos
│   ├── osmi_grpc.pb.go        # Servidor gRPC
│   └── osmi.pb.gw.go          # Gateway HTTP (para osmi-gateway)
├── internal/
│   ├── db/
│   │   └── db.go              # Conexión PostgreSQL con pgxpool
│   ├── service/
│   │   └── service.go         # Implementación de métodos gRPC
│   ├── repository/
│   │   ├── customer_repository.go     # CRUD de clientes
│   │   ├── ticket_repository.go       # CRUD de tickets
│   │   └── event_repository.go        # CRUD de eventos
│   ├── models/
│   │   └── models.go          # Estructuras: Customer, Ticket, Event
│   ├── context/
│   │   └── context.go         # Propagación de contexto y auditoría
│   ├── auth/                  # Validación de tokens y roles (en desarrollo)
│   ├── middleware/            # Interceptores gRPC (en desarrollo)
│   └── utils/                 # Validaciones y helpers
├── third_party/
│   └── googleapis/            # Dependencias de protobuf
├── docker/
│   └── Dockerfile             # Imagen para despliegue
├── k8s/                       # Configuración Kubernetes
├── config/                    # Configuración de aplicación
├── .env                       # Variables de entorno
├── .dockerignore              # Exclusión de archivos en Docker
├── generate_proto_fixed.bat   # Script generación código proto
├── go.mod
├── go.sum
├── .gitignore
├── CHANGELOG.md
├── LICENSE
└── README.md

```

# Ejecución local
```
Requisitos:
Go 1.21+
PostgreSQL ejecutándose
Variables de entorno configuradas en .env

# Instalar dependencias
go mod tidy

# Generar código protobuf (Windows)
generate_proto_fixed.bat

# Ejecutar servidor
go run cmd/main.go

El servidor estará disponible en: localhost:50051
```

## Ejecución con Docker
```
# Construir imagen
docker build -t osmi-server -f docker/Dockerfile .

# Ejecutar contenedor
docker run -p 50051:50051 osmi-server
```

## Endpoints gRPC disponibles
```bash
protobuf
service OsmiService {
  rpc CreateTicket(TicketRequest) returns (TicketResponse);
  rpc ListTickets(UserLookup) returns (TicketListResponse);
  rpc CreateCustomer(CustomerRequest) returns (CustomerResponse);
  rpc GetCustomer(CustomerLookup) returns (CustomerResponse);
  rpc CreateUser(UserRequest) returns (UserResponse);
  rpc CreateEvent(EventRequest) returns (EventResponse);
  rpc GetEvent(EventLookup) returns (EventResponse);
  rpc ListEvents(Empty) returns (EventListResponse);
}
```

## Endpoints REST vía grpc-gateway
```
Método	Ruta	Descripción
POST	/users	Crear usuario
POST	/tickets	Crear ticket
POST	/customers	Crear cliente
GET	/customers/{id}	Obtener cliente por ID
```

### Health & Readiness
```
GET /health → Verifica conexión a base de datos
GET /ready → Verifica estado de conexión y estadísticas
Disponibles en: http://localhost:8081
```

## Generación de Código Proto
### Después de modificar proto/osmi.proto, ejecutar:

```bash
generate_proto_fixed.bat
Este script genera código para:

osmi-server: Servidor gRPC en gen/

osmi-gateway: Gateway HTTP en ../osmi-gateway/gen/
```

## Estado actual
### Completado
Servidor gRPC completamente funcional en puerto 50051
Todos los métodos del servicio implementados
Conexión a PostgreSQL operativa
Health checks activos en puerto 8081
Repositorios para Customers, Tickets y Events
Script de generación de código protobuf

## En Desarrollo
Kubernetes deployment
Autenticación y autorización
Interceptores gRPC
Métricas y monitoring

## Configuración
### Variables de entorno requeridas en .env:
```bash
DATABASE_URL=postgresql://user:pass@host:port/db
GRPC_PORT=50051
HEALTH_PORT=8081
```

## COMANDOS EXACTOS PARA REGENERAR CÓDIGO gRPC
📁 PRIMERO: En el SERVIDOR (osmi-server)
bash
# 1. Navegar al directorio del servidor
cd /c/Users/Desfragmentado/Desktop/Servidor/osmi/osmi-server

# 2. Verificar que tienes el proto actualizado
ls -la proto/osmi.proto

# 3. Regenerar TODO el código gRPC (esto creará/actualizará los archivos en gen/)
protoc \
  --go_out=. \
  --go-grpc_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=. \
  --grpc-gateway_opt=paths=source_relative \
  -I=./proto \
  -I=./proto/googleapis \
  proto/osmi.proto

# 4. Verificar que se generaron los archivos
ls -la gen/

# Deberías ver estos archivos actualizados:
# osmi.pb.go
# osmi_grpc.pb.go  
# osmi.pb.gw.go
📁 SEGUNDO: En el GATEWAY (osmi-gateway)
bash
# 1. Navegar al directorio del gateway
cd /c/Users/Desfragmentado/Desktop/Servidor/osmi/osmi-gateway

# 2. COPIAR el proto actualizado del servidor al gateway (IMPORTANTE!)
cp ../osmi-server/proto/osmi.proto proto/
cp -r ../osmi-server/proto/googleapis proto/

# 3. Regenerar el código del gateway
protoc \
  --go_out=. \
  --go-grpc_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=. \
  --grpc-gateway_opt=paths=source_relative \
  -I=./proto \
  -I=./proto/googleapis \
  proto/osmi.proto

# 4. Verificar que se generaron los archivos
ls -la gen/
🔧 TERCERO: Limpiar y recompilar AMBOS proyectos
bash
# En el SERVIDOR
cd /c/Users/Desfragmentado/Desktop/Servidor/osmi/osmi-server
go clean -cache
go mod tidy
go build -o osmi-server cmd/main.go

# En el GATEWAY  
cd /c/Users/Desfragmentado/Desktop/Servidor/osmi/osmi-gateway
go clean -cache
go mod tidy
go build -o osmi-gateway cmd/main.go

# Autor
### Francisco David Zamora Urrutia Fullstack Developer · Systems Architect · Lyricist