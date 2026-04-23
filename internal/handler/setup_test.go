package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	testPool   *pgxpool.Pool
	testRouter *chi.Mux
)

// TestMain sets up the test database connection and router once for all tests.
func TestMain(m *testing.M) {
	// Try loading .env from project root (relative to this package: internal/handler/)
	_ = godotenv.Load("../../.env")

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		// Fallback: derive test URL from DATABASE_URL by replacing the DB name
		mainURL := os.Getenv("DATABASE_URL")
		if mainURL != "" {
			dbURL = strings.Replace(mainURL, "/dozingo_db?", "/dozingo_test?", 1)
			dbURL = strings.Replace(dbURL, "/dozingo?", "/dozingo_test?", 1)
		}
	}
	if dbURL == "" {
		log.Fatal("TEST_DATABASE_URL (or DATABASE_URL) is not set. Ensure .env is configured and Docker postgres is running.")
	}

	var err error
	testPool, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("failed to create test pool: %v", err)
	}

	if err := testPool.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping test database: %v", err)
	}

	// Set up the router with all handlers registered
	testRouter = chi.NewMux()
	api := humachi.New(testRouter, huma.DefaultConfig("Dozingo Test API", "0.1.0"))
	apiGroup := huma.NewGroup(api, "/api")

	RegisterHealth(apiGroup)
	RegisterUsers(apiGroup, testPool)
	RegisterLecturers(apiGroup, testPool)
	RegisterBoards(apiGroup, testPool)
	RegisterCells(apiGroup, testPool)
	RegisterVotes(apiGroup, testPool)

	code := m.Run()

	testPool.Close()
	os.Exit(code)
}

// cleanupTables truncates all tables in the correct order (respecting foreign keys).
func cleanupTables(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(),
		"TRUNCATE TABLE votes, cells, boards, lecturers, users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("failed to clean up tables: %v", err)
	}
}

// doRequest performs an HTTP request against the test router and returns the response.
func doRequest(method, path string, body any) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal request body: %v", err))
		}
		req = httptest.NewRequest(method, path, strings.NewReader(string(b)))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

// decodeJSON decodes a JSON response body into the given target.
func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(target); err != nil {
		t.Fatalf("failed to decode response body: %v\nbody: %s", err, w.Body.String())
	}
}

// assertStatus checks that the response status code matches the expected one.
func assertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Errorf("expected status %d, got %d\nbody: %s", expected, w.Code, w.Body.String())
	}
}

// assertJSONField checks that a specific field in the JSON response has the expected string value.
func assertJSONField(t *testing.T, data map[string]any, key string, expected string) {
	t.Helper()
	val, ok := data[key]
	if !ok {
		t.Errorf("expected field %q in response, but it was missing", key)
		return
	}
	str, ok := val.(string)
	if !ok {
		t.Errorf("expected field %q to be a string, got %T", key, val)
		return
	}
	if str != expected {
		t.Errorf("expected %q = %q, got %q", key, expected, str)
	}
}

// createTestUser creates a user via the API and returns its ID.
func createTestUser(t *testing.T, username, email string) string {
	t.Helper()
	w := doRequest(http.MethodPost, "/api/users", map[string]string{
		"username": username,
		"email":    email,
	})
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	return resp["id"].(string)
}

// createTestLecturer creates a lecturer via the API and returns its ID.
func createTestLecturer(t *testing.T, name, slug string) string {
	t.Helper()
	w := doRequest(http.MethodPost, "/api/lecturers", map[string]string{
		"name": name,
		"slug": slug,
	})
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	return resp["id"].(string)
}

// createTestBoard creates a board via the API and returns its ID.
func createTestBoard(t *testing.T, title string, size int, authorID, lecturerID string) string {
	t.Helper()
	w := doRequest(http.MethodPost, "/api/boards", map[string]any{
		"title":       title,
		"size":        size,
		"author_id":   authorID,
		"lecturer_id": lecturerID,
	})
	assertStatus(t, w, http.StatusOK)
	var resp map[string]any
	decodeJSON(t, w, &resp)
	return resp["id"].(string)
}
