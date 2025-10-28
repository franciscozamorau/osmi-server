# Osmi Server
Backend gRPC para la plataforma Osmi. Este módulo implementa el núcleo del sistema de boletaje digital, utilizando una arquitectura escalable, segura y profesional. Incluye integración REST vía grpc-gateway, validación de salud, y simulación de endpoints para pruebas.
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
│   └── main.go                        # Entrypoint principal del servidor gRPC con health checks y graceful shutdown
├── config/                            # Carga de variables de entorno y configuración
├── internal/
│   ├── auth/                          # Validación de tokens y roles (pendiente)
│   ├── contex/
│   │   └── contex.go                  # Propagación de contexto entre capas
│   ├── db/
│   │   └── db.go                      # Inicialización y conexión a PostgreSQL con pgxpool
│   ├── middleware/                    # Interceptores gRPC (pendiente)
│   ├── service/
│   │   └── service.go                 # Implementación real de métodos gRPC: clientes, tickets, eventos
│   ├── utils/                         # Validaciones y helpers internos
│   ├── repository/
│   │   ├── customer_repository.go     # CRUD de clientes
│   │   ├── ticket_repository.go       # CRUD de tickets
│   │   ├── event_repository.go        # CRUD de eventos
│   └── models/
│       └── models.go                  # Estructuras de datos: Customer, Ticket, Event
├── proto/
│   └── osmi.proto                     # Definición de servicios gRPC con anotaciones REST
├── gen/
│   ├── osmi.pb.go                     # Código generado por protoc
│   ├── osmi_grpc.pb.go                # Interfaces gRPC
│   └── osmi.pb.gw.go                  # Gateway REST ↔ gRPC (generado por grpc-gateway)
├── docker/
│   └── Dockerfile                     # Imagen para despliegue gRPC
├── k8s/
│   ├── deployment.yaml                # Despliegue en Kubernetes
│   ├── service.yaml                   # Exposición del servicio
│   ├── ingress.yaml                   # Entrada HTTP
│   ├── configmap.yaml                 # Configuración externa
│   └── secret.yaml                    # Variables sensibles
├── third_party/
│   └── google/api/
│       ├── annotations.proto          # Anotaciones REST para grpc-gateway
│       └── http.proto
├── .dockerignore                      # Exclusión de binarios y archivos temporales
├── go.mod
├── go.sum
├── .gitignore
├── LICENSE.md
└── README.md

```

# Ejecución local
```
go mod tidy
go run cmd/main.go
```

## Ejecución con Docker
```
docker build -t osmi-server -f docker/Dockerfile .
docker run -p 50051:50051 osmi-server
```

## Endpoints gRPC disponibles
```bash
protobuf
service OsmiService {
  rpc CreateUser (UserRequest) returns (UserResponse);
  rpc CreateTicket (TicketRequest) returns (TicketResponse);
  rpc CreateCustomer (CustomerRequest) returns (CustomerResponse);
  rpc GetCustomer (CustomerLookup) returns (CustomerResponse);
  rpc CreateEvent (EventRequest) returns (EventResponse);
  rpc GetEvent (EventLookup) returns (EventResponse);
  rpc ListEvents (Empty) returns (EventListResponse);
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
```

# Estado actual
Todos los métodos gRPC están implementados y funcionales
El gateway traduce correctamente las rutas HTTP ↔ gRPC
La base de datos está conectada y operativa
El servidor responde con datos reales y simulados
Health checks activos en /health y /ready

# Autor
### Francisco David Zamora Urrutia Fullstack Developer · Systems Architect · Lyricist