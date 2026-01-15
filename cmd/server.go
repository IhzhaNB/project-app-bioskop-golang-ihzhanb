package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// APIServerWithPort starts server
func APIServer(route *chi.Mux, port string) {
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Server running on http://localhost%s\n", addr)

	if err := http.ListenAndServe(addr, route); err != nil {
		log.Fatal("Server error:", err)
	}
}
