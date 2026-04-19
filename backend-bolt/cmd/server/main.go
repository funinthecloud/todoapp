package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/funinthecloud/protosource"
	"github.com/funinthecloud/protosource/adapters/httpstandard"
	"github.com/funinthecloud/protosource/authz"
)

func main() {
	repo, err := InitializeRepository()
	if err != nil {
		log.Fatal(err)
	}

	authorizer := InitializeAuthorizer()

	router := protosource.NewRouter()
	handler := InitializeHandler(repo)
	handler.RegisterRoutes(router)

	router.Handle("GET", "whoami", whoamiHandler(authorizer))

	router.SetCORS(protosource.CORSConfig{
		AllowOrigin:  "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,X-Actor,Authorization",
	})

	addr := ":8080"
	fmt.Printf("Showcase server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, httpstandard.WrapRouter(router, actorExtractor)))
}

// actorExtractor prefers an Authorization: Bearer <token> header (shadow
// tokens issued by protosource-auth) and falls back to X-Actor for
// developer convenience when running against allowall.Authorizer.
//
// When PROTOSOURCE_AUTH_URL is set in wire.go's provideAuthorizer, the
// Authorizer dereferences the bearer token against the auth service
// and enriches the context with the real user id. The Actor field
// carries the raw token in that mode — a future protosource template
// update will read the authenticated user id from context instead,
// making the command's actor the human-readable user id.
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
		body, _ := json.Marshal(map[string]string{"actor": actor})
		return protosource.Response{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       string(body),
		}
	}
}

func actorExtractor(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return r.Header.Get("X-Actor")
}
