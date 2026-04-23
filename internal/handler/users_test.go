package handler

import (
	"fmt"
	"net/http"
	"testing"
)

func TestCreateUser(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	body := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
	}

	w := doRequest(http.MethodPost, "/api/users", body)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "username", "testuser")
	assertJSONField(t, resp, "email", "test@example.com")

	if _, ok := resp["id"]; !ok {
		t.Error("expected 'id' field in response")
	}
}

func TestGetUserByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	// Create a user first
	createBody := map[string]string{
		"username": "getuser",
		"email":    "getuser@example.com",
	}
	createResp := doRequest(http.MethodPost, "/api/users", createBody)
	assertStatus(t, createResp, http.StatusOK)

	var created map[string]any
	decodeJSON(t, createResp, &created)
	userID := created["id"].(string)

	// Get the user by ID
	w := doRequest(http.MethodGet, fmt.Sprintf("/api/users/%s", userID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "id", userID)
	assertJSONField(t, resp, "username", "getuser")
	assertJSONField(t, resp, "email", "getuser@example.com")
}

func TestGetUserByID_NotFound(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	w := doRequest(http.MethodGet, "/api/users/00000000-0000-0000-0000-000000000000", nil)
	assertStatus(t, w, http.StatusNotFound)
}

func TestGetUserByID_InvalidUUID(t *testing.T) {
	w := doRequest(http.MethodGet, "/api/users/not-a-uuid", nil)
	// huma validates the UUID format at the path parameter level and returns 422
	assertStatus(t, w, http.StatusUnprocessableEntity)
}

func TestDeleteUser(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	// Create a user first
	createBody := map[string]string{
		"username": "deleteuser",
		"email":    "delete@example.com",
	}
	createResp := doRequest(http.MethodPost, "/api/users", createBody)
	assertStatus(t, createResp, http.StatusOK)

	var created map[string]any
	decodeJSON(t, createResp, &created)
	userID := created["id"].(string)

	// Delete the user
	w := doRequest(http.MethodDelete, fmt.Sprintf("/api/users/%s", userID), nil)
	assertStatus(t, w, http.StatusNoContent)

	// Verify user no longer exists
	getResp := doRequest(http.MethodGet, fmt.Sprintf("/api/users/%s", userID), nil)
	assertStatus(t, getResp, http.StatusNotFound)
}

func TestDeleteUser_NotFound(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	w := doRequest(http.MethodDelete, "/api/users/00000000-0000-0000-0000-000000000000", nil)
	assertStatus(t, w, http.StatusNotFound)
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	body := map[string]string{
		"username": "dupeuser",
		"email":    "dupe1@example.com",
	}
	w := doRequest(http.MethodPost, "/api/users", body)
	assertStatus(t, w, http.StatusOK)

	// Try creating another user with the same username
	body2 := map[string]string{
		"username": "dupeuser",
		"email":    "dupe2@example.com",
	}
	w2 := doRequest(http.MethodPost, "/api/users", body2)
	assertStatus(t, w2, http.StatusInternalServerError)
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	body := map[string]string{
		"username": "user1",
		"email":    "shared@example.com",
	}
	w := doRequest(http.MethodPost, "/api/users", body)
	assertStatus(t, w, http.StatusOK)

	// Try creating another user with the same email
	body2 := map[string]string{
		"username": "user2",
		"email":    "shared@example.com",
	}
	w2 := doRequest(http.MethodPost, "/api/users", body2)
	assertStatus(t, w2, http.StatusInternalServerError)
}
