// Package db provides database abstraction for Matrix MUD.
// This file contains audit logging database operations.
package db

import (
	"database/sql"
	"fmt"
	"time"
)

// AuditRepository handles audit log database operations
type AuditRepository struct {
	db *DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// AuditAction defines types of auditable actions
type AuditAction string

const (
	AuditLogin       AuditAction = "LOGIN"
	AuditLogout      AuditAction = "LOGOUT"
	AuditCreate      AuditAction = "CREATE"
	AuditDelete      AuditAction = "DELETE"
	AuditLevelUp     AuditAction = "LEVEL_UP"
	AuditDeath       AuditAction = "DEATH"
	AuditKill        AuditAction = "KILL"
	AuditTrade       AuditAction = "TRADE"
	AuditPurchase    AuditAction = "PURCHASE"
	AuditSale        AuditAction = "SALE"
	AuditQuestStart  AuditAction = "QUEST_START"
	AuditQuestEnd    AuditAction = "QUEST_END"
	AuditAchievement AuditAction = "ACHIEVEMENT"
	AuditAdminAction AuditAction = "ADMIN"
	AuditChat        AuditAction = "CHAT"
	AuditPvP         AuditAction = "PVP"
	AuditBan         AuditAction = "BAN"
	AuditUnban       AuditAction = "UNBAN"
	AuditWarning     AuditAction = "WARNING"
)

// AuditEntry represents an audit log entry for creation
type AuditEntry struct {
	PlayerID   int64
	PlayerName string
	Action     AuditAction
	Details    string
	IPAddress  string
}

// Log creates an audit log entry
func (r *AuditRepository) Log(entry *AuditEntry) error {
	_, err := r.db.Exec(`
		INSERT INTO audit_log (player_id, player_name, action, details, ip_address, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`, entry.PlayerID, entry.PlayerName, string(entry.Action), entry.Details, entry.IPAddress, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

// LogAction is a convenience method for quick logging
func (r *AuditRepository) LogAction(playerID int64, playerName string, action AuditAction, details string) error {
	return r.Log(&AuditEntry{
		PlayerID:   playerID,
		PlayerName: playerName,
		Action:     action,
		Details:    details,
	})
}

// GetByPlayer returns audit logs for a specific player
func (r *AuditRepository) GetByPlayer(playerID int64, limit int) ([]*AuditLog, error) {
	rows, err := r.db.Query(`
		SELECT id, player_id, player_name, action, details, ip_address, timestamp
		FROM audit_log WHERE player_id = ?
		ORDER BY timestamp DESC LIMIT ?
	`, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// GetByPlayerName returns audit logs for a player by name
func (r *AuditRepository) GetByPlayerName(name string, limit int) ([]*AuditLog, error) {
	rows, err := r.db.Query(`
		SELECT id, player_id, player_name, action, details, ip_address, timestamp
		FROM audit_log WHERE player_name = ? COLLATE NOCASE
		ORDER BY timestamp DESC LIMIT ?
	`, name, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// GetByAction returns audit logs for a specific action type
func (r *AuditRepository) GetByAction(action AuditAction, limit int) ([]*AuditLog, error) {
	rows, err := r.db.Query(`
		SELECT id, player_id, player_name, action, details, ip_address, timestamp
		FROM audit_log WHERE action = ?
		ORDER BY timestamp DESC LIMIT ?
	`, string(action), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// GetRecent returns recent audit logs
func (r *AuditRepository) GetRecent(limit int) ([]*AuditLog, error) {
	rows, err := r.db.Query(`
		SELECT id, player_id, player_name, action, details, ip_address, timestamp
		FROM audit_log ORDER BY timestamp DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// GetInTimeRange returns audit logs within a time range
func (r *AuditRepository) GetInTimeRange(start, end time.Time, limit int) ([]*AuditLog, error) {
	rows, err := r.db.Query(`
		SELECT id, player_id, player_name, action, details, ip_address, timestamp
		FROM audit_log 
		WHERE timestamp BETWEEN ? AND ?
		ORDER BY timestamp DESC LIMIT ?
	`, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// Search searches audit logs by details text
func (r *AuditRepository) Search(query string, limit int) ([]*AuditLog, error) {
	rows, err := r.db.Query(`
		SELECT id, player_id, player_name, action, details, ip_address, timestamp
		FROM audit_log WHERE details LIKE ?
		ORDER BY timestamp DESC LIMIT ?
	`, "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// GetCount returns total count of audit logs
func (r *AuditRepository) GetCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM audit_log").Scan(&count)
	return count, err
}

// GetCountByAction returns count of logs for an action type
func (r *AuditRepository) GetCountByAction(action AuditAction) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM audit_log WHERE action = ?", string(action)).Scan(&count)
	return count, err
}

// Cleanup removes audit logs older than specified duration
func (r *AuditRepository) Cleanup(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result, err := r.db.Exec("DELETE FROM audit_log WHERE timestamp < ?", cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// scanLogs scans rows into audit log entries
func (r *AuditRepository) scanLogs(rows *sql.Rows) ([]*AuditLog, error) {
	var logs []*AuditLog
	for rows.Next() {
		log := &AuditLog{}
		var playerID sql.NullInt64
		var playerName, details, ipAddress sql.NullString

		err := rows.Scan(
			&log.ID, &playerID, &playerName, &log.Action,
			&details, &ipAddress, &log.Timestamp,
		)
		if err != nil {
			return nil, err
		}

		if playerID.Valid {
			log.PlayerID = playerID.Int64
		}
		if playerName.Valid {
			log.Details = playerName.String
		}
		if details.Valid {
			log.Details = details.String
		}
		if ipAddress.Valid {
			log.IPAddress = ipAddress.String
		}

		logs = append(logs, log)
	}
	return logs, nil
}

// Session represents a player session
type Session struct {
	ID           int64
	PlayerID     int64
	SessionToken string
	IPAddress    string
	UserAgent    string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

// CreateSession creates a new session
func (r *AuditRepository) CreateSession(session *Session) error {
	result, err := r.db.Exec(`
		INSERT INTO sessions (player_id, session_token, ip_address, user_agent, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, session.PlayerID, session.SessionToken, session.IPAddress, session.UserAgent, time.Now(), session.ExpiresAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	session.ID = id
	return nil
}

// GetSession retrieves a session by token
func (r *AuditRepository) GetSession(token string) (*Session, error) {
	session := &Session{}
	err := r.db.QueryRow(`
		SELECT id, player_id, session_token, ip_address, user_agent, created_at, expires_at
		FROM sessions WHERE session_token = ?
	`, token).Scan(
		&session.ID, &session.PlayerID, &session.SessionToken,
		&session.IPAddress, &session.UserAgent, &session.CreatedAt, &session.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return session, err
}

// DeleteSession deletes a session
func (r *AuditRepository) DeleteSession(token string) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE session_token = ?", token)
	return err
}

// DeleteExpiredSessions removes expired sessions
func (r *AuditRepository) DeleteExpiredSessions() (int64, error) {
	result, err := r.db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetPlayerSessions returns all sessions for a player
func (r *AuditRepository) GetPlayerSessions(playerID int64) ([]*Session, error) {
	rows, err := r.db.Query(`
		SELECT id, player_id, session_token, ip_address, user_agent, created_at, expires_at
		FROM sessions WHERE player_id = ?
	`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		s := &Session{}
		err := rows.Scan(&s.ID, &s.PlayerID, &s.SessionToken, &s.IPAddress, &s.UserAgent, &s.CreatedAt, &s.ExpiresAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}
