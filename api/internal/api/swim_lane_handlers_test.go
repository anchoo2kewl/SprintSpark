package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// createDefaultSwimLanes inserts the 3 default swim lanes for a test project
// and returns their IDs in order [To Do, In Progress, Done].
func createDefaultSwimLanes(t *testing.T, ts *TestServer, projectID int64) [3]int64 {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	defaults := []struct {
		name     string
		color    string
		position int
	}{
		{"To Do", "#6B7280", 0},
		{"In Progress", "#3B82F6", 1},
		{"Done", "#10B981", 2},
	}

	var ids [3]int64
	for i, sl := range defaults {
		result, err := ts.DB.ExecContext(ctx,
			`INSERT INTO swim_lanes (project_id, name, color, position) VALUES (?, ?, ?, ?)`,
			projectID, sl.name, sl.color, sl.position,
		)
		if err != nil {
			t.Fatalf("Failed to create swim lane %q: %v", sl.name, err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("Failed to get swim lane ID: %v", err)
		}
		ids[i] = id
	}
	return ids
}

func TestHandleListSwimLanes(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")
	createDefaultSwimLanes(t, ts, projectID)

	rec, req := ts.MakeAuthRequest(t, http.MethodGet, fmt.Sprintf("/api/projects/%d/swim-lanes", projectID), nil, userID,
		map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

	ts.HandleListSwimLanes(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	var lanes []SwimLane
	DecodeJSON(t, rec, &lanes)

	if len(lanes) != 3 {
		t.Fatalf("Expected 3 swim lanes, got %d", len(lanes))
	}

	// Verify they come back ordered by position
	expectedNames := []string{"To Do", "In Progress", "Done"}
	for i, name := range expectedNames {
		if lanes[i].Name != name {
			t.Errorf("Lane %d: expected name %q, got %q", i, name, lanes[i].Name)
		}
		if lanes[i].Position != i {
			t.Errorf("Lane %d: expected position %d, got %d", i, i, lanes[i].Position)
		}
		if lanes[i].ProjectID != projectID {
			t.Errorf("Lane %d: expected project_id %d, got %d", i, projectID, lanes[i].ProjectID)
		}
	}
}

func TestHandleCreateSwimLane(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")
	createDefaultSwimLanes(t, ts, projectID)

	body := CreateSwimLaneRequest{
		Name:     "Review",
		Color:    "#F59E0B",
		Position: 3,
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPost, fmt.Sprintf("/api/projects/%d/swim-lanes", projectID), body, userID,
		map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

	ts.HandleCreateSwimLane(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusCreated)

	var lane SwimLane
	DecodeJSON(t, rec, &lane)

	if lane.Name != "Review" {
		t.Errorf("Expected name 'Review', got %q", lane.Name)
	}
	if lane.Color != "#F59E0B" {
		t.Errorf("Expected color '#F59E0B', got %q", lane.Color)
	}
	if lane.Position != 3 {
		t.Errorf("Expected position 3, got %d", lane.Position)
	}
	if lane.ProjectID != projectID {
		t.Errorf("Expected project_id %d, got %d", projectID, lane.ProjectID)
	}
	if lane.ID == 0 {
		t.Error("Expected non-zero swim lane ID")
	}
}

func TestHandleCreateSwimLaneDefaultColor(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")

	body := CreateSwimLaneRequest{
		Name:     "Backlog",
		Position: 0,
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPost, fmt.Sprintf("/api/projects/%d/swim-lanes", projectID), body, userID,
		map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

	ts.HandleCreateSwimLane(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusCreated)

	var lane SwimLane
	DecodeJSON(t, rec, &lane)

	if lane.Color != "#6B7280" {
		t.Errorf("Expected default color '#6B7280', got %q", lane.Color)
	}
}

func TestHandleCreateSwimLaneValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       CreateSwimLaneRequest
		wantStatus int
		wantError  string
	}{
		{
			name:       "missing name",
			body:       CreateSwimLaneRequest{Name: "", Color: "#FF0000", Position: 0},
			wantStatus: http.StatusBadRequest,
			wantError:  "swim lane name is required",
		},
		{
			name:       "name too long",
			body:       CreateSwimLaneRequest{Name: strings.Repeat("x", 51), Color: "#FF0000", Position: 0},
			wantStatus: http.StatusBadRequest,
			wantError:  "swim lane name is too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			userID := ts.CreateTestUser(t, "test@example.com", "password123")
			projectID := ts.CreateTestProject(t, userID, "Test Project")

			rec, req := ts.MakeAuthRequest(t, http.MethodPost, fmt.Sprintf("/api/projects/%d/swim-lanes", projectID), tt.body, userID,
				map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

			ts.HandleCreateSwimLane(rec, req)

			AssertError(t, rec, tt.wantStatus, tt.wantError, "invalid_input")
		})
	}
}

func TestHandleCreateSwimLaneMaxLimit(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create 6 swim lanes (the maximum)
	for i := 0; i < 6; i++ {
		_, err := ts.DB.ExecContext(ctx,
			`INSERT INTO swim_lanes (project_id, name, color, position) VALUES (?, ?, ?, ?)`,
			projectID, fmt.Sprintf("Lane %d", i), "#AABBCC", i,
		)
		if err != nil {
			t.Fatalf("Failed to create swim lane %d: %v", i, err)
		}
	}

	body := CreateSwimLaneRequest{
		Name:     "Seventh Lane",
		Color:    "#FF0000",
		Position: 6,
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPost, fmt.Sprintf("/api/projects/%d/swim-lanes", projectID), body, userID,
		map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

	ts.HandleCreateSwimLane(rec, req)

	AssertError(t, rec, http.StatusBadRequest, "maximum 6 swim lanes", "max_limit_reached")
}

func TestHandleUpdateSwimLane(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")
	laneIDs := createDefaultSwimLanes(t, ts, projectID)

	newName := "Backlog"
	newColor := "#EF4444"
	newPosition := 5
	body := UpdateSwimLaneRequest{
		Name:     &newName,
		Color:    &newColor,
		Position: &newPosition,
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPatch, fmt.Sprintf("/api/swim-lanes/%d", laneIDs[0]), body, userID,
		map[string]string{"id": fmt.Sprintf("%d", laneIDs[0])})

	ts.HandleUpdateSwimLane(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	var lane SwimLane
	DecodeJSON(t, rec, &lane)

	if lane.Name != "Backlog" {
		t.Errorf("Expected name 'Backlog', got %q", lane.Name)
	}
	if lane.Color != "#EF4444" {
		t.Errorf("Expected color '#EF4444', got %q", lane.Color)
	}
	if lane.Position != 5 {
		t.Errorf("Expected position 5, got %d", lane.Position)
	}
}

func TestHandleUpdateSwimLanePartial(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")
	laneIDs := createDefaultSwimLanes(t, ts, projectID)

	// Update only the name, leave color and position unchanged
	newName := "Updated Name"
	body := UpdateSwimLaneRequest{
		Name: &newName,
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPatch, fmt.Sprintf("/api/swim-lanes/%d", laneIDs[1]), body, userID,
		map[string]string{"id": fmt.Sprintf("%d", laneIDs[1])})

	ts.HandleUpdateSwimLane(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	var lane SwimLane
	DecodeJSON(t, rec, &lane)

	if lane.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %q", lane.Name)
	}
	// Color should remain unchanged
	if lane.Color != "#3B82F6" {
		t.Errorf("Expected original color '#3B82F6', got %q", lane.Color)
	}
	// Position should remain unchanged
	if lane.Position != 1 {
		t.Errorf("Expected original position 1, got %d", lane.Position)
	}
}

func TestHandleUpdateSwimLaneNotFound(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")

	newName := "Ghost Lane"
	body := UpdateSwimLaneRequest{
		Name: &newName,
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPatch, "/api/swim-lanes/99999", body, userID,
		map[string]string{"id": "99999"})

	ts.HandleUpdateSwimLane(rec, req)

	AssertError(t, rec, http.StatusNotFound, "swim lane not found", "not_found")
}

func TestHandleDeleteSwimLane(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")
	laneIDs := createDefaultSwimLanes(t, ts, projectID)

	// Delete one of the 3 lanes (should succeed since 3 > 2)
	rec, req := ts.MakeAuthRequest(t, http.MethodDelete, fmt.Sprintf("/api/swim-lanes/%d", laneIDs[2]), nil, userID,
		map[string]string{"id": fmt.Sprintf("%d", laneIDs[2])})

	ts.HandleDeleteSwimLane(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusNoContent)

	// Verify it was actually deleted
	var count int
	err := ts.DB.QueryRow("SELECT COUNT(*) FROM swim_lanes WHERE id = ?", laneIDs[2]).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query swim lane count: %v", err)
	}
	if count != 0 {
		t.Error("Swim lane was not deleted from database")
	}
}

func TestHandleDeleteSwimLaneNotFound(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")

	rec, req := ts.MakeAuthRequest(t, http.MethodDelete, "/api/swim-lanes/99999", nil, userID,
		map[string]string{"id": "99999"})

	ts.HandleDeleteSwimLane(rec, req)

	AssertError(t, rec, http.StatusNotFound, "swim lane not found", "not_found")
}

func TestHandleDeleteSwimLaneMinimumLimit(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create exactly 2 swim lanes (the minimum)
	result1, err := ts.DB.ExecContext(ctx,
		`INSERT INTO swim_lanes (project_id, name, color, position) VALUES (?, ?, ?, ?)`,
		projectID, "To Do", "#6B7280", 0,
	)
	if err != nil {
		t.Fatalf("Failed to create swim lane: %v", err)
	}
	laneID, err := result1.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get swim lane ID: %v", err)
	}

	_, err = ts.DB.ExecContext(ctx,
		`INSERT INTO swim_lanes (project_id, name, color, position) VALUES (?, ?, ?, ?)`,
		projectID, "Done", "#10B981", 1,
	)
	if err != nil {
		t.Fatalf("Failed to create swim lane: %v", err)
	}

	// Try to delete one of the 2 remaining lanes (should fail)
	rec, req := ts.MakeAuthRequest(t, http.MethodDelete, fmt.Sprintf("/api/swim-lanes/%d", laneID), nil, userID,
		map[string]string{"id": fmt.Sprintf("%d", laneID)})

	ts.HandleDeleteSwimLane(rec, req)

	AssertError(t, rec, http.StatusBadRequest, "minimum 2 swim lanes required", "min_limit_reached")
}

func TestHandleListSwimLanesUnauthorized(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	user1ID := ts.CreateTestUser(t, "user1@example.com", "password123")
	user2ID := ts.CreateTestUser(t, "user2@example.com", "password123")
	projectID := ts.CreateTestProject(t, user1ID, "User1 Project")
	createDefaultSwimLanes(t, ts, projectID)

	// user2 is NOT a member of user1's project
	rec, req := ts.MakeAuthRequest(t, http.MethodGet, fmt.Sprintf("/api/projects/%d/swim-lanes", projectID), nil, user2ID,
		map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

	ts.HandleListSwimLanes(rec, req)

	AssertError(t, rec, http.StatusForbidden, "access denied", "forbidden")
}

func TestHandleCreateSwimLaneUnauthorized(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	user1ID := ts.CreateTestUser(t, "user1@example.com", "password123")
	user2ID := ts.CreateTestUser(t, "user2@example.com", "password123")
	projectID := ts.CreateTestProject(t, user1ID, "User1 Project")

	body := CreateSwimLaneRequest{
		Name:     "Hacked Lane",
		Color:    "#FF0000",
		Position: 0,
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPost, fmt.Sprintf("/api/projects/%d/swim-lanes", projectID), body, user2ID,
		map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

	ts.HandleCreateSwimLane(rec, req)

	AssertError(t, rec, http.StatusForbidden, "access denied", "forbidden")
}

func TestHandleUpdateSwimLaneUnauthorized(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	user1ID := ts.CreateTestUser(t, "user1@example.com", "password123")
	user2ID := ts.CreateTestUser(t, "user2@example.com", "password123")
	projectID := ts.CreateTestProject(t, user1ID, "User1 Project")
	laneIDs := createDefaultSwimLanes(t, ts, projectID)

	newName := "Hacked Name"
	body := UpdateSwimLaneRequest{
		Name: &newName,
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPatch, fmt.Sprintf("/api/swim-lanes/%d", laneIDs[0]), body, user2ID,
		map[string]string{"id": fmt.Sprintf("%d", laneIDs[0])})

	ts.HandleUpdateSwimLane(rec, req)

	AssertError(t, rec, http.StatusForbidden, "access denied", "forbidden")
}

func TestHandleDeleteSwimLaneUnauthorized(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	user1ID := ts.CreateTestUser(t, "user1@example.com", "password123")
	user2ID := ts.CreateTestUser(t, "user2@example.com", "password123")
	projectID := ts.CreateTestProject(t, user1ID, "User1 Project")
	laneIDs := createDefaultSwimLanes(t, ts, projectID)

	rec, req := ts.MakeAuthRequest(t, http.MethodDelete, fmt.Sprintf("/api/swim-lanes/%d", laneIDs[0]), nil, user2ID,
		map[string]string{"id": fmt.Sprintf("%d", laneIDs[0])})

	ts.HandleDeleteSwimLane(rec, req)

	AssertError(t, rec, http.StatusForbidden, "access denied", "forbidden")
}

func TestHandleListSwimLanesInvalidProjectID(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")

	rec, req := ts.MakeAuthRequest(t, http.MethodGet, "/api/projects/abc/swim-lanes", nil, userID,
		map[string]string{"projectId": "abc"})

	ts.HandleListSwimLanes(rec, req)

	AssertError(t, rec, http.StatusBadRequest, "invalid project ID", "invalid_input")
}
