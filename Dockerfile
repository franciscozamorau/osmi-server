FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY osmi-server/go.mod osmi-server/go.sum ./osmi-server/
COPY osmi-protobuf/go.mod ./osmi-protobuf/
COPY osmi-protobuf/go.sum ./osmi-protobuf/
COPY osmi-protobuf/gen ./osmi-protobuf/gen
COPY osmi-protobuf/proto ./osmi-protobuf/proto

WORKDIR /app/osmi-server

RUN go mod download

COPY osmi-server ./

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go


FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/osmi-server/server .
COPY --from=builder /app/osmi-server/.env.production ./.env

EXPOSE 50051

CMD ["./server"]