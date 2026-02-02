module github.com/franciscozamorau/osmi-server

go 1.24

require (
    github.com/franciscozamorau/osmi-protobuf v0.0.0
    github.com/go-chi/chi/v5 v5.0.12
    github.com/go-chi/cors v1.2.1
    github.com/go-playground/validator/v10 v10.19.0
    github.com/golang-jwt/jwt/v5 v5.2.1
    github.com/google/uuid v1.6.0
    github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1
    github.com/jackc/pgx/v5 v5.5.5
    github.com/jmoiron/sqlx v1.4.0
    github.com/joho/godotenv v1.5.1
    github.com/lib/pq v1.10.9
    github.com/redis/go-redis/v9 v9.5.1
    github.com/spf13/viper v1.19.0
    github.com/stripe/stripe-go/v78 v78.3.0
    go.uber.org/zap v1.27.0
    golang.org/x/crypto v0.23.0
    google.golang.org/grpc v1.64.0
    google.golang.org/protobuf v1.34.1
)

replace github.com/franciscozamorau/osmi-protobuf => ../osmi-protobuf
