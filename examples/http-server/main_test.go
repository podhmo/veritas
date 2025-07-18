package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/podhmo/veritas"
)

func TestUserAPI(t *testing.T) {
	// setup validator
	v, err := veritas.NewValidatorFromJSONFile("./rules.json")
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	// setup mux
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := v.Validate(user); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			response := map[string]any{
				"message": "validation failed",
				"details": veritas.ToErrorMap(err),
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	})

	// setup server
	server := httptest.NewServer(mux)
	defer server.Close()

	// test cases
	tests := []struct {
		name           string
		body           string
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "success",
			body:           `{"name": "test", "email": "test@example.com"}`,
			wantStatusCode: http.StatusCreated,
			wantBody:       `{"name":"test","email":"test@example.com"}`,
		},
		{
			name:           "validation error - name required",
			body:           `{"name": "", "email": "test@example.com"}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `"details":{"Name":"name is required"}`,
		},
		{
			name:           "validation error - email invalid",
			body:           `{"name": "test", "email": "invalid-email"}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `"details":{"Email":"email must contain @"}`,
		},
		{
			name:           "validation error - both invalid",
			body:           `{"name": "", "email": "invalid-email"}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `"details":{"Email":"email must contain @","Name":"name is required"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", server.URL+"/users", bytes.NewBufferString(tt.body))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("unexpected status code: got %v, want %v", resp.StatusCode, tt.wantStatusCode)
			}

			var body bytes.Buffer
			body.ReadFrom(resp.Body)

			if !strings.Contains(body.String(), tt.wantBody) {
				t.Errorf("unexpected body: got %v, want %v", body.String(), tt.wantBody)
			}
		})
	}
}

func TestRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx)
	}()

	// Wait for the server to start (or fail)
	time.Sleep(100 * time.Millisecond)

	// Shutdown the server by canceling the context
	cancel()

	// Check if run() returned an error
	select {
	case err := <-errCh:
		if err != nil && !strings.Contains(err.Error(), "http: server closed") {
			t.Errorf("run() returned an unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("run() did not exit in time")
	}
}
