# Osmi Server

Backend gRPC para la plataforma Osmi. Este módulo implementa el núcleo del sistema de boletaje digital, utilizando una arquitectura escalable y profesional.

---

## 🧱 Osmi Core Stack

- **Go** → lenguaje principal
- **gRPC** → protocolo de comunicación
- **Protobuf** → definición de servicios
- **grpc-gateway** → puente REST ↔ gRPC (próxima etapa)
- **PostgreSQL** → base de datos relacional (próxima etapa)
- **Kubernetes** → orquestación y despliegue (próxima etapa)

---

## 📁 Estructura del proyecto

osmi-server/
├── cmd/
│   └── main.go
├── internal/
│   ├── grpc/
│   ├── gateway/
│   ├── service/
│   ├── middleware/
│   ├── repository/
│   ├── auth/
│   └── utils/
├── proto/
│   └── osmi.proto
├── gen/
│   ├── osmi.pb.go
│   ├── osmi_grpc.pb.go
├── config/
├── docker/
│   └── Dockerfile
├── k8s/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress.yaml
│   ├── configmap.yaml
│   └── secret.yaml
├── go.mod
├── go.sum



---

## 🚀 Cómo correr el servidor localmente

```bash
go mod tidy
go run cmd/main.go


---
## 🚀 Cómo correr con Docker

```bash

docker build -t osmi-server -f docker/Dockerfile .
docker run -p 50051:50051 osmi-server


## 📡 Endpoint gRPC disponible

rpc CreateTicket (TicketRequest) returns (TicketResponse);


📌 Próximas etapas
[ ] Cliente gRPC (osmi-client)

[ ] Gateway REST (osmi-gateway)

[ ] Conexión a PostgreSQL (osmi-db)

[ ] Seguridad con JWT (internal/auth)

[ ] Despliegue en Kubernetes (k8s/)

🧠 Autor
Francisco D. Zamora — Fullstack Developer & Systems Engineer