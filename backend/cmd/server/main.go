package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/funinthecloud/protosource"
	"github.com/funinthecloud/protosource/adapters/httpstandard"
)

func main() {
	repo := InitializeRepository()

	router := protosource.NewRouter()
	handler := InitializeHandler(repo)
	handler.RegisterRoutes(router)

	wrapped := corsMiddleware(httpstandard.WrapRouter(router, httpstandard.HeaderExtractor("X-Actor")))

	addr := ":8080"
	fmt.Printf("Showcase server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, wrapped))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Actor")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
