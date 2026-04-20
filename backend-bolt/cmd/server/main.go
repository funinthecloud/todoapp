package main

import (
	"context"
	"encoding/json"
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

	router.Handle("GET", "whoami", whoamiHandler(authorizer))

	addr := ":8080"
	fmt.Printf("Showcase server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, corsMiddleware(httpstandard.WrapRouter(router, actorExtractor))))
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

// corsMiddleware handles CORS with credentials support. Validates the
// request Origin against CORS_ALLOWED_ORIGINS (comma-separated) or
// defaults to http://localhost:5173 for local dev.
func corsMiddleware(next http.Handler) http.Handler {
	allowed := buildAllowedOrigins()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,X-Actor,Authorization")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func buildAllowedOrigins() map[string]bool {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		raw = "http://localhost:5173"
	}
	m := make(map[string]bool)
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			m[o] = true
		}
	}
	return m
}

func whoamiHandler(authorizer authz.Authorizer) protosource.HandlerFunc {
	return func(ctx context.Context, req protosource.Request) protosource.Response {
		enrichedCtx, err := authorizer.Authorize(ctx, req, "showcase.app.todolist.v1.WhoAmI")
		if err != nil {
			body, _ := json.Marshal(map[string]string{"error": "unauthorized"})
			return protosource.Response{
				StatusCode: http.StatusUnauthorized,
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       string(body),
			}
		}
		actor := authz.UserIDFromContext(enrichedCtx)
		if actor == "" {
			actor = req.Actor
		}
		if actor == "" {
			body, _ := json.Marshal(map[string]string{"error": "unauthorized"})
			return protosource.Response{
				StatusCode: http.StatusUnauthorized,
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       string(body),
			}
		}
		body, _ := json.Marshal(map[string]string{"actor": actor})
		return protosource.Response{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       string(body),
		}
	}
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
