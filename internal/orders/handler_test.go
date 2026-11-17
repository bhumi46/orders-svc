package orders

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeUsersSvc is a real HTTP server standing in for users-svc, so the client is tested too.
func fakeUsersSvc(t *testing.T, status int) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/users/1" && status == http.StatusOK {
			w.WriteHeader(http.StatusOK)
			return
		}
		if status == http.StatusOK {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func newMux(usersURL string) *http.ServeMux {
	mux := http.NewServeMux()
	NewHandler(NewStore(), NewUsersClient(usersURL)).Register(mux)
	return mux
}

func do(mux http.Handler, method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

func TestCreateOrder(t *testing.T) {
	healthy := fakeUsersSvc(t, http.StatusOK)
	broken := fakeUsersSvc(t, http.StatusInternalServerError)

	tests := []struct {
		name     string
		usersURL string
		body     string
		want     int
	}{
		{"known user", healthy, `{"user_id":"1","item":"book","quantity":2}`, http.StatusCreated},
		{"unknown user", healthy, `{"user_id":"42","item":"book","quantity":1}`, http.StatusUnprocessableEntity},
		{"users-svc error", broken, `{"user_id":"1","item":"book","quantity":1}`, http.StatusBadGateway},
		{"users-svc down", "http://127.0.0.1:1", `{"user_id":"1","item":"book","quantity":1}`, http.StatusBadGateway},
		{"bad quantity", healthy, `{"user_id":"1","item":"book","quantity":0}`, http.StatusBadRequest},
		{"bad json", healthy, `nope`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(newMux(tt.usersURL), http.MethodPost, "/api/orders", tt.body)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tt.want, rec.Body)
			}
		})
	}
}

func TestListAndGet(t *testing.T) {
	mux := newMux(fakeUsersSvc(t, http.StatusOK))
	if rec := do(mux, http.MethodPost, "/api/orders", `{"user_id":"1","item":"pen","quantity":1}`); rec.Code != http.StatusCreated {
		t.Fatalf("create = %d", rec.Code)
	}
	if rec := do(mux, http.MethodGet, "/api/orders", ""); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"item":"pen"`) {
		t.Fatalf("list = %d %s", rec.Code, rec.Body)
	}
	if rec := do(mux, http.MethodGet, "/api/orders/1", ""); rec.Code != http.StatusOK {
		t.Fatalf("get 1 = %d", rec.Code)
	}
	if rec := do(mux, http.MethodGet, "/api/orders/2", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("get 2 = %d", rec.Code)
	}
}

func TestUsersClientEscapesID(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	err := NewUsersClient(srv.URL+"/").ValidateUser(context.Background(), "a/b")
	if err != ErrUnknownUser || gotPath != "/api/users/a%2Fb" {
		t.Fatalf("err=%v path=%q", err, gotPath)
	}
}
