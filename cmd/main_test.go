package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Configurar puertos
	grpcPort := getEnv("GRPC_PORT", "50051")
	healthPort := getEnv("HEALTH_PORT", "8081")
	httpPort := getEnv("HTTP_PORT", "8080")

	// Iniciar health check server
	go startHealthServer(healthPort)

	// Iniciar servidor HTTP de prueba
	go startTestHTTPServer(httpPort)

	// Mensaje de inicio
	log.Printf("Ì∫Ä OSMI Server starting...")
	log.Printf("   gRPC Port: %s", grpcPort)
	log.Printf("   Health Port: %s", healthPort)
	log.Printf("   HTTP Port: %s", httpPort)
	log.Printf("   Environment: %s", getEnv("ENVIRONMENT", "development"))

	// Mantener el programa corriendo
	select {}
}

func startHealthServer(port string) {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "healthy", "service": "osmi-server", "timestamp": "%s"}`, time.Now().UTC().Format(time.RFC3339))
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ready", "timestamp": "%s"}`, time.Now().UTC().Format(time.RFC3339))
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("‚úÖ Health server running on port %s", port)
	
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("‚ùå Health server failed: %v", err)
	}
}

func startTestHTTPServer(port string) {
	mux := http.NewServeMux()

	// Endpoints de prueba
	mux.HandleFunc("/api/v1/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "OSMI API is working!", "version": "1.0.0", "timestamp": "%s"}`, 
			time.Now().UTC().Format(time.RFC3339))
	})

	mux.HandleFunc("/api/v1/tickets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"tickets": [], "total": 0, "message": "Ticket endpoint is ready"}`)
		} else if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"id": "tkt_12345", "status": "created", "message": "Ticket created successfully"}`)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, `{"error": "Method not allowed"}`)
		}
	})

	mux.HandleFunc("/api/v1/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"events": [], "total": 0, "message": "Events endpoint is ready"}`)
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("‚úÖ Test HTTP server running on port %s", port)
	
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("‚ö† Test HTTP server failed: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
