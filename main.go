package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mayanksekhar/GoLang-EKS-Monitor-app/internal/dashboard"
	"github.com/mayanksekhar/GoLang-EKS-Monitor-app/internal/metrics"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	collector, err := metrics.NewCollector()
	if err != nil {
		log.Fatalf("failed to create metrics collector: %v", err)
	}

	handler := dashboard.NewHandler(collector)

	http.HandleFunc("/", handler.Index)
	http.HandleFunc("/api/metrics", handler.APIMetrics)
	http.HandleFunc("/health", handler.Health)

	fmt.Printf("EKS Monitor running on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
