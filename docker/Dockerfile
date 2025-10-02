FROM golang:1.23

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o osmi_server ./cmd/main.go

EXPOSE 50051
CMD ["./osmi_server"]