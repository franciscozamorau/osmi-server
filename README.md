# Osmi Server
Backend gRPC para la plataforma Osmi. Este módulo implementa el núcleo del sistema de boletaje digital, utilizando una arquitectura escalable, segura y profesional. Incluye integración REST vía grpc-gateway, validación de salud, y simulación de endpoints para pruebas.
---

# Osmi Core Stack
```bash
Go → lenguaje principal
gRPC → protocolo de comunicación entre servicios
Protobuf → definición de contratos y mensajes
grpc-gateway → puente REST ↔ gRPC (activo)
PostgreSQL → base de datos relacional (en proceso)
Kubernetes → orquestación y despliegue (en proceso)
Docker → contenedorización del servicio
.env + godotenv → gestión de variables de entorno
Health & Readiness Probes → verificación de estado del sistema
```

# Estructura del proyecto
```bash
osmi-server/
├── cmd/
│   └── main.go                  # Entrypoint principal del servidor gRPC
├── internal/
│   ├── service/                 # Implementación de métodos gRPC simulados
│   ├── repository/              # Acceso a datos (PostgreSQL)
│   ├── db/                      # Inicialización y conexión a la base de datos
│   └── ...                      # Otros módulos internos
├── proto/
│   └── osmi.proto               # Definición de servicios y mensajes
├── gen/
│   ├── osmi.pb.go               # Código generado por protoc
│   ├── osmi_grpc.pb.go          # Interfaces gRPC
├── docker/
│   └── Dockerfile               # Imagen para despliegue
├── k8s/
│   ├── deployment.yaml          # Despliegue en Kubernetes
│   ├── service.yaml             # Exposición del servicio
│   ├── ingress.yaml             # Entrada HTTP
│   ├── configmap.yaml           # Configuración externa
│   └── secret.yaml              # Variables sensibles
├── go.mod
├── go.sum
├── LICENSE.md
└── README.md
```

# Ejecución local
```bash
go mod tidy
go run cmd/main.go
```

# Ejecución con Docker
```bash
docker build -t osmi-server -f docker/Dockerfile .
docker run -p 50051:50051 osmi-server
Endpoints gRPC disponibles
proto
service OsmiService {
  rpc CreateUser (UserRequest) returns (UserResponse);
  rpc CreateTicket (TicketRequest) returns (TicketResponse);
  rpc CreateCustomer (CustomerRequest) returns (CustomerResponse);
  rpc GetCustomer (CustomerLookup) returns (CustomerResponse);
}

Endpoints REST vía grpc-gateway
Método	Ruta	Descripción
POST	/users	Crear usuario
POST	/tickets	Crear ticket
POST	/customers	Crear cliente
GET	/customers/{id}	Obtener cliente por ID

Health & Readiness
GET /health → Verifica conexión a base de datos
GET /ready → Verifica estado de conexión y estadísticas
```

# Estado actual
Todos los endpoints gRPC y REST están simulados y funcionales
El gateway traduce correctamente las rutas HTTP ↔ gRPC
La base de datos está inicializada y conectada
El servidor responde con datos simulados para pruebas

# Autor
### Francisco David Zamora Urrutia Fullstack Developer · Systems Architect · Lyricist