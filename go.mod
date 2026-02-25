module github.com/franciscozamorau/osmi-server

go 1.24.0

require (
	github.com/franciscozamorau/osmi-protobuf v0.0.0-00010101000000-000000000000
	github.com/go-chi/chi/v5 v5.0.12
	github.com/go-chi/cors v1.2.1
	github.com/go-playground/validator/v10 v10.19.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0
	github.com/jackc/pgx/v5 v5.5.5
	github.com/jmoiron/sqlx v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.11.2
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.46.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260209200024-4cfbd4190f57
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260209200024-4cfbd4190f57 // indirect
)

replace github.com/franciscozamorau/osmi-protobuf => ../osmi-protobuf
