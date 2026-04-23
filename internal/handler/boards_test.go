package handler

import (
	"fmt"
	"net/http"
	"testing"
)

func TestCreateBoard(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	userID := createTestUser(t, "boardauthor", "boardauthor@example.com")
	lecturerID := createTestLecturer(t, "Board Lecturer", "board-lecturer")

	body := map[string]any{
		"title":       "Test Board",
		"size":        5,
		"author_id":   userID,
		"lecturer_id": lecturerID,
	}

	w := doRequest(http.MethodPost, "/api/boards", body)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "title", "Test Board")
	assertJSONField(t, resp, "author_id", userID)
	assertJSONField(t, resp, "lecturer_id", lecturerID)

	if _, ok := resp["id"]; !ok {
		t.Error("expected 'id' field in response")
	}

	// Check size (JSON numbers decode as float64)
	if size, ok := resp["size"].(float64); !ok || int(size) != 5 {
		t.Errorf("expected size = 5, got %v", resp["size"])
	}
}

func TestGetBoards(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	userID := createTestUser(t, "listauthor", "listauthor@example.com")
	lecturerID := createTestLecturer(t, "List Lecturer", "list-lecturer")

	createTestBoard(t, "Board A", 5, userID, lecturerID)
	createTestBoard(t, "Board B", 3, userID, lecturerID)

	w := doRequest(http.MethodGet, "/api/boards", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 2 {
		t.Errorf("expected 2 boards, got %d", len(resp))
	}
}

func TestGetBoards_Empty(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	w := doRequest(http.MethodGet, "/api/boards", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 0 {
		t.Errorf("expected 0 boards, got %d", len(resp))
	}
}

func TestGetBoards_FilterByAuthor(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	user1 := createTestUser(t, "author1", "author1@example.com")
	user2 := createTestUser(t, "author2", "author2@example.com")
	lecturerID := createTestLecturer(t, "Filter Lecturer", "filter-lecturer")

	createTestBoard(t, "User1 Board", 5, user1, lecturerID)
	createTestBoard(t, "User2 Board", 5, user2, lecturerID)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards?author_id=%s", user1), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 1 {
		t.Errorf("expected 1 board, got %d", len(resp))
		return
	}
	assertJSONField(t, resp[0], "title", "User1 Board")
}

func TestGetBoards_FilterByLecturer(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	userID := createTestUser(t, "filterauthor", "filterauthor@example.com")
	lec1 := createTestLecturer(t, "Lecturer A", "lecturer-a")
	lec2 := createTestLecturer(t, "Lecturer B", "lecturer-b")

	createTestBoard(t, "Lec1 Board", 5, userID, lec1)
	createTestBoard(t, "Lec2 Board", 5, userID, lec2)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards?lecturer_id=%s", lec1), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 1 {
		t.Errorf("expected 1 board, got %d", len(resp))
		return
	}
	assertJSONField(t, resp[0], "title", "Lec1 Board")
}

func TestGetBoards_FilterBySize(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	userID := createTestUser(t, "sizeauthor", "sizeauthor@example.com")
	lecturerID := createTestLecturer(t, "Size Lecturer", "size-lecturer")

	createTestBoard(t, "Small Board", 3, userID, lecturerID)
	createTestBoard(t, "Large Board", 7, userID, lecturerID)

	w := doRequest(http.MethodGet, "/api/boards?size=3", nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 1 {
		t.Errorf("expected 1 board, got %d", len(resp))
		return
	}
	assertJSONField(t, resp[0], "title", "Small Board")
}

func TestGetBoardByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	userID := createTestUser(t, "getboardauthor", "getboard@example.com")
	lecturerID := createTestLecturer(t, "GetBoard Lecturer", "getboard-lecturer")
	boardID := createTestBoard(t, "GetMe Board", 5, userID, lecturerID)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s", boardID), nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]any
	decodeJSON(t, w, &resp)

	assertJSONField(t, resp, "id", boardID)
	assertJSONField(t, resp, "title", "GetMe Board")
	assertJSONField(t, resp, "author_id", userID)
	assertJSONField(t, resp, "lecturer_id", lecturerID)
}

func TestGetBoardByID_NotFound(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	w := doRequest(http.MethodGet, "/api/boards/00000000-0000-0000-0000-000000000000", nil)
	assertStatus(t, w, http.StatusNotFound)
}

func TestDeleteBoard(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	userID := createTestUser(t, "delboardauthor", "delboard@example.com")
	lecturerID := createTestLecturer(t, "DelBoard Lecturer", "delboard-lecturer")
	boardID := createTestBoard(t, "DeleteMe Board", 5, userID, lecturerID)

	// Delete the board
	w := doRequest(http.MethodDelete, fmt.Sprintf("/api/boards/%s", boardID), nil)
	assertStatus(t, w, http.StatusNoContent)

	// Verify board is gone
	getResp := doRequest(http.MethodGet, fmt.Sprintf("/api/boards/%s", boardID), nil)
	assertStatus(t, getResp, http.StatusNotFound)
}

func TestDeleteBoard_NotFound(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	w := doRequest(http.MethodDelete, "/api/boards/00000000-0000-0000-0000-000000000000", nil)
	assertStatus(t, w, http.StatusNotFound)
}

func TestGetBoards_CombinedFilters(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	user1 := createTestUser(t, "comboauthor1", "combo1@example.com")
	user2 := createTestUser(t, "comboauthor2", "combo2@example.com")
	lec := createTestLecturer(t, "Combo Lecturer", "combo-lecturer")

	createTestBoard(t, "Match", 5, user1, lec)
	createTestBoard(t, "Wrong Author", 5, user2, lec)
	createTestBoard(t, "Wrong Size", 3, user1, lec)

	w := doRequest(http.MethodGet, fmt.Sprintf("/api/boards?author_id=%s&size=5", user1), nil)
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]any
	decodeJSON(t, w, &resp)

	if len(resp) != 1 {
		t.Errorf("expected 1 board, got %d", len(resp))
		return
	}
	assertJSONField(t, resp[0], "title", "Match")
}
