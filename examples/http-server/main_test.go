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
	v, err := veritas.NewValidatorFromJSONFile("./rules.json", veritas.WithTypeAdapters(
		map[string]veritas.TypeAdapter{
			"http-server.User": func(ob any) (map[string]any, error) {
				v, ok := ob.(User)
				if !ok {
					return nil, nil // a
				}
				return map[string]any{
					"Name":  v.Name,
					"Email": v.Email,
				}, nil
			},
		},
	))
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

		if err := v.Validate(r.Context(), user); err != nil {
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
			wantBody:       `"details":{"Name":"size(self) > 0"}`,
		},
		{
			name:           "validation error - email invalid",
			body:           `{"name": "test", "email": "invalid-email"}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `"details":{"Email":"self.contains('@')"}`,
		},
		{
			name:           "validation error - both invalid",
			body:           `{"name": "", "email": "invalid-email"}`,
			wantStatusCode: http.StatusBadRequest,
			wantBody:       `"details":{"Email":"self.contains('@')","Name":"size(self) > 0"}`,
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

			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			if tt.wantStatusCode == http.StatusBadRequest {
				details, ok := body["details"].(map[string]any)
				if !ok {
					t.Fatalf("details is not a map[string]any")
				}
				if tt.name == "validation error - both invalid" {
					if details["Name"] != "size(self) > 0" {
						t.Errorf("unexpected error for Name: got %v, want %v", details["Name"], "size(self) > 0")
					}
					if details["Email"] != "self.contains('@')" {
						t.Errorf("unexpected error for Email: got %v, want %v", details["Email"], "self.contains('@')")
					}
				} else if tt.name == "validation error - name required" {
					if details["Name"] != "size(self) > 0" {
						t.Errorf("unexpected error for Name: got %v, want %v", details["Name"], "size(self) > 0")
					}
				} else if tt.name == "validation error - email invalid" {
					if details["Email"] != "self.contains('@')" {
						t.Errorf("unexpected error for Email: got %v, want %v", details["Email"], "self.contains('@')")
					}
				}
			} else {
				var wantBody map[string]any
				if err := json.Unmarshal([]byte(tt.wantBody), &wantBody); err != nil {
					t.Fatalf("failed to unmarshal wantBody: %v", err)
				}
				if body["name"] != wantBody["name"] || body["email"] != wantBody["email"] {
					t.Errorf("unexpected body: got %v, want %v", body, wantBody)
				}
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
