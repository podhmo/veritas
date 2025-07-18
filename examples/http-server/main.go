package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/podhmo/veritas"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.ErrorContext(ctx, "toplevel error", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// setup validator
	slog.InfoContext(ctx, "setup validator")
	v, err := veritas.NewValidatorFromJSONFile("./rules.json")
	if err != nil {
		return err
	}

	// setup routes
	slog.InfoContext(ctx, "setup routes")
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// decode request
		slog.DebugContext(ctx, "decode request")
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// validate
		slog.DebugContext(ctx, "validate request", "user", user)
		if err := v.Validate(user); err != nil {
			slog.InfoContext(ctx, "validation failed", "err", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			response := map[string]any{
				"message": "validation failed",
				"details": veritas.ToErrorMap(err),
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				slog.ErrorContext(ctx, "failed to encode error response", "err", err)
			}
			return
		}

		// write response
		slog.DebugContext(ctx, "write response")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(user); err != nil {
			slog.ErrorContext(ctx, "failed to encode success response", "err", err)
		}
	})

	// start server
	slog.InfoContext(ctx, "start server", "port", ":8080")
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "ListenAndServe failed", "err", err)
		}
	}()
	slog.InfoContext(ctx, "server started")

	<-ctx.Done()

	slog.InfoContext(ctx, "shutting down server")
	return server.Shutdown(context.Background())
}
