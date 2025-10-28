# Osmi Server — CHANGELOG

## v0.1.0 — 2025-10-21

- Implementación completa de métodos gRPC: CreateCustomer, GetCustomer, CreateTicket, CreateEvent, GetEvent, ListEvents
- Integración con PostgreSQL vía pgxpool
- Migraciones desacopladas en osmi-db
- Traducción REST ↔ gRPC vía grpc-gateway
- Health & readiness probes activos
- Dockerfile funcional y validado
- `.env` protegido y sin fugas
- Documentación técnica consolidada
