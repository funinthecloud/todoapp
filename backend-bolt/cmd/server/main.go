package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/funinthecloud/protosource"
	"github.com/funinthecloud/protosource/adapters/httpstandard"
	"github.com/funinthecloud/protosource/authz"
	"github.com/funinthecloud/protosource/authz/allowall"
	"github.com/funinthecloud/protosource-auth/authz/httpauthz"
)

func main() {
	repo, err := InitializeRepository()
	if err != nil {
		log.Fatal(err)
	}

	authorizer := provideAuthorizer()

	router := protosource.NewRouter()
	handler := InitializeHandler(repo, authorizer)
	handler.RegisterRoutes(router)

	router.SetCORS(protosource.CORSConfig{
		AllowOrigins:     buildAllowedOrigins(),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type,X-Actor,Authorization",
		AllowCredentials: true,
	})

	addr := ":8080"
	fmt.Printf("Showcase server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, httpstandard.WrapRouter(router, actorExtractor)))
}

// provideAuthorizer returns httpauthz when PROTOSOURCE_AUTH_URL is set,
// otherwise falls back to allowall for local development.
func provideAuthorizer() authz.Authorizer {
	authURL := os.Getenv("PROTOSOURCE_AUTH_URL")
	if authURL == "" {
		return allowall.Authorizer{}
	}
	return httpauthz.New(authURL, httpauthz.WithTokenSource(
		httpauthz.Chain(httpauthz.Cookie("shadow"), httpauthz.AuthorizationHeader()),
	))
}

func buildAllowedOrigins() []string {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		raw = "http://localhost:5173"
	}
	var origins []string
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}

// actorExtractor prefers a "shadow" cookie (HttpOnly, set by the auth
// service), then falls back to Authorization: Bearer <token> and X-Actor
// for developer convenience.
func actorExtractor(r *http.Request) string {
	if c, err := r.Cookie("shadow"); err == nil && c.Value != "" {
		return c.Value
	}
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return r.Header.Get("X-Actor")
}
