# Build stage
FROM golang:1.25.1-alpine AS builder

WORKDIR /go/src/github.com/franciscozamorau/osmi-server

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o osmi_server ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /go/src/github.com/franciscozamorau/osmi-server/osmi_server .

EXPOSE 50051

CMD ["./osmi_server"]
