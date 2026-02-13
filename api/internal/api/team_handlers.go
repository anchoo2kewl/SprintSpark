package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Team struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	OwnerID   int64     `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TeamMember struct {
	ID       int64     `json:"id"`
	TeamID   int64     `json:"team_id"`
	UserID   int64     `json:"user_id"`
	UserName *string   `json:"user_name,omitempty"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	Status   string    `json:"status"`
	JoinedAt time.Time `json:"joined_at"`
}

type TeamInvitation struct {
	ID            int64      `json:"id"`
	TeamID        int64      `json:"team_id"`
	TeamName      string     `json:"team_name"`
	InviterID     int64      `json:"inviter_id"`
	InviterName   *string    `json:"inviter_name,omitempty"`
	InviteeEmail  string     `json:"invitee_email"`
	InviteeID     *int64     `json:"invitee_id,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	RespondedAt   *time.Time `json:"responded_at,omitempty"`
}

type CreateTeamRequest struct {
	Name string `json:"name"`
}

type InviteTeamMemberRequest struct {
	Email string `json:"email"`
}

type UpdateTeamMemberRequest struct {
	Role string `json:"role"`
}

// HandleGetMyTeam returns the current user's team
func (s *Server) HandleGetMyTeam(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)

	// Get user's active team membership
	query := `
		SELECT t.id, t.name, t.owner_id, t.created_at, t.updated_at
		FROM teams t
		JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = ? AND tm.status = 'active'
		LIMIT 1
	`

	var team Team
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&team.ID, &team.Name, &team.OwnerID, &team.CreatedAt, &team.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "no active team found", "not_found")
		return
	} else if err != nil {
		s.logger.Error("Failed to get user's team", zap.Error(err), zap.Int64("user_id", userID))
		respondError(w, http.StatusInternalServerError, "failed to fetch team", "internal_error")
		return
	}

	respondJSON(w, http.StatusOK, team)
}

// HandleGetTeamMembers returns all members of the user's team
func (s *Server) HandleGetTeamMembers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)

	// Get user's team ID
	teamID, err := s.getUserTeamID(ctx, userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "no active team found", "not_found")
		return
	}

	// Get all team members
	query := `
		SELECT tm.id, tm.team_id, tm.user_id, u.name, u.email, tm.role, tm.status, tm.joined_at
		FROM team_members tm
		JOIN users u ON tm.user_id = u.id
		WHERE tm.team_id = ?
		ORDER BY tm.role DESC, tm.joined_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, teamID)
	if err != nil {
		s.logger.Error("Failed to get team members", zap.Error(err), zap.Int64("team_id", teamID))
		respondError(w, http.StatusInternalServerError, "failed to fetch team members", "internal_error")
		return
	}
	defer rows.Close()

	members := []TeamMember{}
	for rows.Next() {
		var m TeamMember
		if err := rows.Scan(&m.ID, &m.TeamID, &m.UserID, &m.UserName, &m.Email, &m.Role, &m.Status, &m.JoinedAt); err != nil {
			s.logger.Error("Failed to scan team member", zap.Error(err))
			respondError(w, http.StatusInternalServerError, "failed to scan team member", "internal_error")
			return
		}
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error iterating team members", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "error iterating team members", "internal_error")
		return
	}

	respondJSON(w, http.StatusOK, members)
}

// HandleInviteTeamMember sends an invitation to join the team
func (s *Server) HandleInviteTeamMember(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)

	var req InviteTeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "invalid_input")
		return
	}

	// Validate email
	if req.Email == "" || !isValidEmail(req.Email) {
		respondError(w, http.StatusBadRequest, "valid email is required", "invalid_input")
		return
	}

	// Get user's team ID
	teamID, err := s.getUserTeamID(ctx, userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "no active team found", "not_found")
		return
	}

	// Check if user is owner or admin
	role, err := s.getUserTeamRole(ctx, userID, teamID)
	if err != nil || (role != "owner" && role != "admin") {
		respondError(w, http.StatusForbidden, "only team owners and admins can invite members", "forbidden")
		return
	}

	// Check if invitee is already a member
	var existingMemberID int64
	checkMemberQuery := `
		SELECT tm.id FROM team_members tm
		JOIN users u ON tm.user_id = u.id
		WHERE tm.team_id = ? AND u.email = ?
	`
	err = s.db.QueryRowContext(ctx, checkMemberQuery, teamID, req.Email).Scan(&existingMemberID)
	if err == nil {
		respondError(w, http.StatusConflict, "user is already a team member", "already_member")
		return
	} else if err != sql.ErrNoRows {
		s.logger.Error("Failed to check existing member", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to check membership", "internal_error")
		return
	}

	// Check if there's already a pending invitation
	var existingInvitationID int64
	checkInvitationQuery := `
		SELECT id FROM team_invitations
		WHERE team_id = ? AND invitee_email = ? AND status = 'pending'
	`
	err = s.db.QueryRowContext(ctx, checkInvitationQuery, teamID, req.Email).Scan(&existingInvitationID)
	if err == nil {
		respondError(w, http.StatusConflict, "pending invitation already exists", "invitation_exists")
		return
	} else if err != sql.ErrNoRows {
		s.logger.Error("Failed to check existing invitation", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to check invitation", "internal_error")
		return
	}

	// Get invitee user ID if they exist
	var inviteeID *int64
	var tempInviteeID int64
	getUserQuery := `SELECT id FROM users WHERE email = ?`
	err = s.db.QueryRowContext(ctx, getUserQuery, req.Email).Scan(&tempInviteeID)
	if err == nil {
		inviteeID = &tempInviteeID
	} else if err != sql.ErrNoRows {
		s.logger.Error("Failed to get invitee user", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to get user", "internal_error")
		return
	}

	// Generate acceptance token for one-click email acceptance
	acceptanceToken, tokenErr := generateInviteCode()
	if tokenErr != nil {
		s.logger.Error("Failed to generate acceptance token", zap.Error(tokenErr))
		respondError(w, http.StatusInternalServerError, "failed to create invitation", "internal_error")
		return
	}

	// Create invitation with acceptance token (expires in 7 days)
	insertQuery := `
		INSERT INTO team_invitations (team_id, inviter_id, invitee_email, invitee_id, status, created_at, acceptance_token, token_expires_at)
		VALUES (?, ?, ?, ?, 'pending', CURRENT_TIMESTAMP, ?, datetime('now', '+7 days'))
	`
	result, err := s.db.ExecContext(ctx, insertQuery, teamID, userID, req.Email, inviteeID, acceptanceToken)
	if err != nil {
		s.logger.Error("Failed to create invitation", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create invitation", "internal_error")
		return
	}

	invitationID, err := result.LastInsertId()
	if err != nil {
		s.logger.Error("Failed to get invitation ID", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to get invitation ID", "internal_error")
		return
	}

	// Fetch created invitation
	invitation, err := s.getInvitation(ctx, invitationID)
	if err != nil {
		s.logger.Error("Failed to fetch created invitation", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to fetch invitation", "internal_error")
		return
	}

	s.logger.Info("Team invitation created",
		zap.Int64("invitation_id", invitationID),
		zap.Int64("team_id", teamID),
		zap.String("invitee_email", req.Email),
	)

	// Send email notification if email service is available
	if emailSvc := s.GetEmailService(); emailSvc != nil {
		// Get inviter name
		var inviterName string
		_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(name, email) FROM users WHERE id = ?`, userID).Scan(&inviterName)

		// Get team name
		var teamName string
		_ = s.db.QueryRowContext(ctx, `SELECT name FROM teams WHERE id = ?`, teamID).Scan(&teamName)

		appURL := s.getAppURL()

		if inviteeID != nil {
			// Existing user — send project invitation with accept link
			if err := emailSvc.SendProjectInvitation(ctx, req.Email, inviterName, teamName, acceptanceToken, appURL); err != nil {
				s.logger.Warn("Failed to send team invitation email",
					zap.String("to", req.Email),
					zap.Error(err),
				)
			}
		} else {
			// New user — auto-generate invite code and send signup link with accept token
			inviteCode, codeErr := generateTeamInviteCode()
			if codeErr == nil {
				// Create a platform invite for this user
				_, _ = s.db.ExecContext(ctx,
					`INSERT INTO invites (code, inviter_id, expires_at) VALUES (?, ?, datetime('now', '+7 days'))`,
					inviteCode, userID,
				)
				// Store invite code on the team invitation for retrieval during acceptance
				_, _ = s.db.ExecContext(ctx,
					`UPDATE team_invitations SET invite_code = ? WHERE id = ?`,
					inviteCode, invitationID,
				)
				if err := emailSvc.SendProjectInvitationNewUser(ctx, req.Email, inviterName, teamName, acceptanceToken, appURL); err != nil {
					s.logger.Warn("Failed to send team invitation email to new user",
						zap.String("to", req.Email),
						zap.Error(err),
					)
				}
			}
		}
	}

	respondJSON(w, http.StatusCreated, invitation)
}

// HandleGetMyInvitations returns all pending invitations for the current user
func (s *Server) HandleGetMyInvitations(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	email := r.Context().Value(UserEmailKey).(string)

	// Get pending invitations for this user
	query := `
		SELECT ti.id, ti.team_id, t.name, ti.inviter_id, u.name, ti.invitee_email,
		       ti.invitee_id, ti.status, ti.created_at, ti.responded_at
		FROM team_invitations ti
		JOIN teams t ON ti.team_id = t.id
		JOIN users u ON ti.inviter_id = u.id
		WHERE (ti.invitee_id = ? OR ti.invitee_email = ?) AND ti.status = 'pending'
		ORDER BY ti.created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID, email)
	if err != nil {
		s.logger.Error("Failed to get invitations", zap.Error(err), zap.Int64("user_id", userID))
		respondError(w, http.StatusInternalServerError, "failed to fetch invitations", "internal_error")
		return
	}
	defer rows.Close()

	invitations := []TeamInvitation{}
	for rows.Next() {
		var inv TeamInvitation
		if err := rows.Scan(&inv.ID, &inv.TeamID, &inv.TeamName, &inv.InviterID, &inv.InviterName,
			&inv.InviteeEmail, &inv.InviteeID, &inv.Status, &inv.CreatedAt, &inv.RespondedAt); err != nil {
			s.logger.Error("Failed to scan invitation", zap.Error(err))
			respondError(w, http.StatusInternalServerError, "failed to scan invitation", "internal_error")
			return
		}
		invitations = append(invitations, inv)
	}

	respondJSON(w, http.StatusOK, invitations)
}

// HandleAcceptInvitation accepts a team invitation
func (s *Server) HandleAcceptInvitation(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	email := r.Context().Value(UserEmailKey).(string)

	invitationID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid invitation ID", "invalid_input")
		return
	}

	// Get invitation and verify it's for this user
	var inv struct {
		TeamID       int64
		InviteeEmail string
		InviteeID    *int64
		Status       string
	}

	query := `SELECT team_id, invitee_email, invitee_id, status FROM team_invitations WHERE id = ?`
	err = s.db.QueryRowContext(ctx, query, invitationID).Scan(
		&inv.TeamID, &inv.InviteeEmail, &inv.InviteeID, &inv.Status,
	)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "invitation not found", "not_found")
		return
	} else if err != nil {
		s.logger.Error("Failed to get invitation", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to fetch invitation", "internal_error")
		return
	}

	// Verify invitation is for this user
	if inv.InviteeEmail != email && (inv.InviteeID == nil || *inv.InviteeID != userID) {
		respondError(w, http.StatusForbidden, "invitation is not for you", "forbidden")
		return
	}

	// Check if invitation is still pending
	if inv.Status != "pending" {
		respondError(w, http.StatusConflict, "invitation already responded to", "already_responded")
		return
	}

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to process invitation", "internal_error")
		return
	}
	defer tx.Rollback()

	// Update invitation status
	updateInvQuery := `
		UPDATE team_invitations
		SET status = 'accepted', invitee_id = ?, responded_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = tx.ExecContext(ctx, updateInvQuery, userID, invitationID)
	if err != nil {
		s.logger.Error("Failed to update invitation", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to update invitation", "internal_error")
		return
	}

	// Add user to team
	addMemberQuery := `
		INSERT INTO team_members (team_id, user_id, role, status, joined_at)
		VALUES (?, ?, 'member', 'active', CURRENT_TIMESTAMP)
	`
	_, err = tx.ExecContext(ctx, addMemberQuery, inv.TeamID, userID)
	if err != nil {
		s.logger.Error("Failed to add team member", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to add team member", "internal_error")
		return
	}

	// Add user to all existing team projects
	addToProjectsQuery := `
		INSERT INTO project_members (project_id, user_id, role, granted_by, granted_at)
		SELECT p.id, ?, 'member', p.owner_id, CURRENT_TIMESTAMP
		FROM projects p
		WHERE p.team_id = ?
	`
	_, err = tx.ExecContext(ctx, addToProjectsQuery, userID, inv.TeamID)
	if err != nil {
		s.logger.Error("Failed to add user to team projects", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to add to team projects", "internal_error")
		return
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to process invitation", "internal_error")
		return
	}

	s.logger.Info("Invitation accepted",
		zap.Int64("invitation_id", invitationID),
		zap.Int64("user_id", userID),
		zap.Int64("team_id", inv.TeamID),
	)

	respondJSON(w, http.StatusOK, map[string]string{"message": "invitation accepted"})
}

// HandleRejectInvitation rejects a team invitation
func (s *Server) HandleRejectInvitation(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	email := r.Context().Value(UserEmailKey).(string)

	invitationID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid invitation ID", "invalid_input")
		return
	}

	// Get invitation and verify it's for this user
	var inv struct {
		InviteeEmail string
		InviteeID    *int64
		Status       string
	}

	query := `SELECT invitee_email, invitee_id, status FROM team_invitations WHERE id = ?`
	err = s.db.QueryRowContext(ctx, query, invitationID).Scan(&inv.InviteeEmail, &inv.InviteeID, &inv.Status)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "invitation not found", "not_found")
		return
	} else if err != nil {
		s.logger.Error("Failed to get invitation", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to fetch invitation", "internal_error")
		return
	}

	// Verify invitation is for this user
	if inv.InviteeEmail != email && (inv.InviteeID == nil || *inv.InviteeID != userID) {
		respondError(w, http.StatusForbidden, "invitation is not for you", "forbidden")
		return
	}

	// Check if invitation is still pending
	if inv.Status != "pending" {
		respondError(w, http.StatusConflict, "invitation already responded to", "already_responded")
		return
	}

	// Update invitation status
	updateQuery := `
		UPDATE team_invitations
		SET status = 'rejected', invitee_id = ?, responded_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = s.db.ExecContext(ctx, updateQuery, userID, invitationID)
	if err != nil {
		s.logger.Error("Failed to reject invitation", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to reject invitation", "internal_error")
		return
	}

	s.logger.Info("Invitation rejected",
		zap.Int64("invitation_id", invitationID),
		zap.Int64("user_id", userID),
	)

	respondJSON(w, http.StatusOK, map[string]string{"message": "invitation rejected"})
}

// HandleRemoveTeamMember removes a member from the team
func (s *Server) HandleRemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)

	memberID, err := strconv.ParseInt(chi.URLParam(r, "memberId"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid member ID", "invalid_input")
		return
	}

	// Get user's team ID
	teamID, err := s.getUserTeamID(ctx, userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "no active team found", "not_found")
		return
	}

	// Check if user is owner or admin
	role, err := s.getUserTeamRole(ctx, userID, teamID)
	if err != nil || (role != "owner" && role != "admin") {
		respondError(w, http.StatusForbidden, "only team owners and admins can remove members", "forbidden")
		return
	}

	// Get member to remove
	var memberUserID int64
	var memberRole string
	getMemberQuery := `SELECT user_id, role FROM team_members WHERE id = ? AND team_id = ?`
	err = s.db.QueryRowContext(ctx, getMemberQuery, memberID, teamID).Scan(&memberUserID, &memberRole)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "member not found", "not_found")
		return
	} else if err != nil {
		s.logger.Error("Failed to get member", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to get member", "internal_error")
		return
	}

	// Cannot remove team owner
	if memberRole == "owner" {
		respondError(w, http.StatusForbidden, "cannot remove team owner", "forbidden")
		return
	}

	// Delete team member
	deleteQuery := `DELETE FROM team_members WHERE id = ? AND team_id = ?`
	_, err = s.db.ExecContext(ctx, deleteQuery, memberID, teamID)
	if err != nil {
		s.logger.Error("Failed to remove team member", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to remove member", "internal_error")
		return
	}

	s.logger.Info("Team member removed",
		zap.Int64("member_id", memberID),
		zap.Int64("user_id", memberUserID),
		zap.Int64("team_id", teamID),
	)

	respondJSON(w, http.StatusOK, map[string]string{"message": "member removed"})
}

// Helper functions

func (s *Server) getUserTeamID(ctx context.Context, userID int64) (int64, error) {
	var teamID int64
	query := `
		SELECT team_id FROM team_members
		WHERE user_id = ? AND status = 'active'
		LIMIT 1
	`
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&teamID)
	return teamID, err
}

func (s *Server) getUserTeamRole(ctx context.Context, userID, teamID int64) (string, error) {
	var role string
	query := `SELECT role FROM team_members WHERE user_id = ? AND team_id = ? AND status = 'active'`
	err := s.db.QueryRowContext(ctx, query, userID, teamID).Scan(&role)
	return role, err
}

func (s *Server) getInvitation(ctx context.Context, invitationID int64) (*TeamInvitation, error) {
	query := `
		SELECT ti.id, ti.team_id, t.name, ti.inviter_id, u.name, ti.invitee_email,
		       ti.invitee_id, ti.status, ti.created_at, ti.responded_at
		FROM team_invitations ti
		JOIN teams t ON ti.team_id = t.id
		JOIN users u ON ti.inviter_id = u.id
		WHERE ti.id = ?
	`

	var inv TeamInvitation
	err := s.db.QueryRowContext(ctx, query, invitationID).Scan(
		&inv.ID, &inv.TeamID, &inv.TeamName, &inv.InviterID, &inv.InviterName,
		&inv.InviteeEmail, &inv.InviteeID, &inv.Status, &inv.CreatedAt, &inv.RespondedAt,
	)
	if err != nil {
		return nil, err
	}

	return &inv, nil
}

// generateTeamInviteCode creates a random invite code (delegates to the shared generator)
func generateTeamInviteCode() (string, error) {
	return generateInviteCode()
}

// TokenInvitationResponse is returned by the token lookup endpoint
type TokenInvitationResponse struct {
	InvitationID int64  `json:"invitation_id"`
	TeamName     string `json:"team_name"`
	InviterName  string `json:"inviter_name"`
	InviteeEmail string `json:"invitee_email"`
	Status       string `json:"status"`
	RequiresSignup bool `json:"requires_signup"`
	InviteCode   string `json:"invite_code,omitempty"`
}

// HandleGetInvitationByToken returns invitation info for a given acceptance token (public, no auth required)
func (s *Server) HandleGetInvitationByToken(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	token := r.URL.Query().Get("token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "token is required", "invalid_input")
		return
	}

	var resp TokenInvitationResponse
	var tokenExpiresAt time.Time
	var inviteeID *int64
	var inviteCode *string

	query := `
		SELECT ti.id, t.name, COALESCE(u.name, u.email), ti.invitee_email, ti.status,
		       ti.invitee_id, ti.token_expires_at, ti.invite_code
		FROM team_invitations ti
		JOIN teams t ON ti.team_id = t.id
		JOIN users u ON ti.inviter_id = u.id
		WHERE ti.acceptance_token = ?
	`
	err := s.db.QueryRowContext(ctx, query, token).Scan(
		&resp.InvitationID, &resp.TeamName, &resp.InviterName, &resp.InviteeEmail,
		&resp.Status, &inviteeID, &tokenExpiresAt, &inviteCode,
	)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "invitation not found or token is invalid", "not_found")
		return
	} else if err != nil {
		s.logger.Error("Failed to get invitation by token", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to fetch invitation", "internal_error")
		return
	}

	// Check token expiry
	if time.Now().After(tokenExpiresAt) {
		respondError(w, http.StatusGone, "invitation link has expired", "token_expired")
		return
	}

	// Check if invitation is still pending
	if resp.Status != "pending" {
		respondError(w, http.StatusConflict, "invitation has already been "+resp.Status, "already_responded")
		return
	}

	// Determine if user needs to sign up
	resp.RequiresSignup = (inviteeID == nil)
	if resp.RequiresSignup && inviteCode != nil {
		resp.InviteCode = *inviteCode
	}

	respondJSON(w, http.StatusOK, resp)
}

// HandleAcceptInvitationByToken accepts a team invitation using the acceptance token (requires auth)
func (s *Server) HandleAcceptInvitationByToken(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	email := r.Context().Value(UserEmailKey).(string)

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		respondError(w, http.StatusBadRequest, "token is required", "invalid_input")
		return
	}

	// Find invitation by token
	var inv struct {
		ID           int64
		TeamID       int64
		InviteeEmail string
		InviteeID    *int64
		Status       string
		TokenExpires time.Time
	}

	query := `
		SELECT id, team_id, invitee_email, invitee_id, status, token_expires_at
		FROM team_invitations
		WHERE acceptance_token = ?
	`
	err := s.db.QueryRowContext(ctx, query, req.Token).Scan(
		&inv.ID, &inv.TeamID, &inv.InviteeEmail, &inv.InviteeID, &inv.Status, &inv.TokenExpires,
	)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "invitation not found or token is invalid", "not_found")
		return
	} else if err != nil {
		s.logger.Error("Failed to get invitation by token", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to fetch invitation", "internal_error")
		return
	}

	// Check token expiry
	if time.Now().After(inv.TokenExpires) {
		respondError(w, http.StatusGone, "invitation link has expired", "token_expired")
		return
	}

	// Verify invitation is for this user
	if inv.InviteeEmail != email {
		respondError(w, http.StatusForbidden, "this invitation is for a different email address", "forbidden")
		return
	}

	// Check if invitation is still pending
	if inv.Status != "pending" {
		respondError(w, http.StatusConflict, "invitation has already been "+inv.Status, "already_responded")
		return
	}

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to process invitation", "internal_error")
		return
	}
	defer tx.Rollback()

	// Update invitation status
	_, err = tx.ExecContext(ctx,
		`UPDATE team_invitations SET status = 'accepted', invitee_id = ?, responded_at = CURRENT_TIMESTAMP WHERE id = ?`,
		userID, inv.ID,
	)
	if err != nil {
		s.logger.Error("Failed to update invitation", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to update invitation", "internal_error")
		return
	}

	// Add user to team
	_, err = tx.ExecContext(ctx,
		`INSERT INTO team_members (team_id, user_id, role, status, joined_at) VALUES (?, ?, 'member', 'active', CURRENT_TIMESTAMP)`,
		inv.TeamID, userID,
	)
	if err != nil {
		s.logger.Error("Failed to add team member", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to add team member", "internal_error")
		return
	}

	// Add user to all existing team projects
	_, err = tx.ExecContext(ctx,
		`INSERT INTO project_members (project_id, user_id, role, granted_by, granted_at)
		 SELECT p.id, ?, 'member', p.owner_id, CURRENT_TIMESTAMP
		 FROM projects p WHERE p.team_id = ?`,
		userID, inv.TeamID,
	)
	if err != nil {
		s.logger.Error("Failed to add user to team projects", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to add to team projects", "internal_error")
		return
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to process invitation", "internal_error")
		return
	}

	s.logger.Info("Invitation accepted via token",
		zap.Int64("invitation_id", inv.ID),
		zap.Int64("user_id", userID),
		zap.Int64("team_id", inv.TeamID),
	)

	respondJSON(w, http.StatusOK, map[string]string{"message": "invitation accepted"})
}

func isValidEmail(email string) bool {
	// Basic email validation
	if len(email) < 3 || len(email) > 254 {
		return false
	}

	atIndex := -1
	for i, c := range email {
		if c == '@' {
			if atIndex >= 0 {
				return false // Multiple @ symbols
			}
			atIndex = i
		}
	}

	if atIndex <= 0 || atIndex >= len(email)-1 {
		return false
	}

	return true
}
