package main

import (
	"log"
	"net/http"

	"ulpf-backend/internal/routes"
)

func main() {

	routes.Setup()

	log.Println("ULPF backend running on http://localhost:8000")

	log.Fatal(
		http.ListenAndServe(":8000", nil),
	)
}