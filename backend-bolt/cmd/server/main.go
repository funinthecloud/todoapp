package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/funinthecloud/protosource"
	"github.com/funinthecloud/protosource/adapters/httpstandard"
)

func main() {
	repo, err := InitializeRepository()
	if err != nil {
		log.Fatal(err)
	}

	router := protosource.NewRouter()
	handler := InitializeHandler(repo)
	handler.RegisterRoutes(router)

	router.SetCORS(protosource.CORSConfig{
		AllowOrigin:  "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,X-Actor",
	})

	addr := ":8080"
	fmt.Printf("Showcase server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, httpstandard.WrapRouter(router, httpstandard.HeaderExtractor("X-Actor"))))
}
