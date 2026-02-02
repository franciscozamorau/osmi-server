FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copiar archivos de módulos
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Construir
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go

# Runtime
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/.env.production ./.env

EXPOSE 50051

CMD ["./server"]
