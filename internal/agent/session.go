package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/siddharth-bhatnagar/anvil/internal/llm"
)

// Session represents a conversation session
type Session struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Messages    []llm.Message `json:"messages"`
	Metadata    SessionMeta  `json:"metadata"`
}

// SessionMeta holds session metadata
type SessionMeta struct {
	Model       string            `json:"model"`
	Provider    string            `json:"provider"`
	TotalTokens int               `json:"total_tokens"`
	WorkingDir  string            `json:"working_dir"`
	Tags        []string          `json:"tags,omitempty"`
	Custom      map[string]string `json:"custom,omitempty"`
}

// SessionStore manages session persistence
type SessionStore struct {
	baseDir string
}

// NewSessionStore creates a new session store
func NewSessionStore(baseDir string) (*SessionStore, error) {
	// Ensure the directory exists
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	return &SessionStore{
		baseDir: baseDir,
	}, nil
}

// DefaultSessionStore returns a session store using the default location
func DefaultSessionStore() (*SessionStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".anvil", "sessions")
	return NewSessionStore(baseDir)
}

// Save saves a session to disk
func (ss *SessionStore) Save(session *Session) error {
	session.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	filename := ss.sessionFilename(session.ID)
	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// Load loads a session from disk
func (ss *SessionStore) Load(id string) (*Session, error) {
	filename := ss.sessionFilename(id)

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session %s not found", id)
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// Delete removes a session from disk
func (ss *SessionStore) Delete(id string) error {
	filename := ss.sessionFilename(id)

	if err := os.Remove(filename); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session %s not found", id)
		}
		return fmt.Errorf("failed to delete session file: %w", err)
	}

	return nil
}

// List returns all sessions, sorted by update time (most recent first)
func (ss *SessionStore) List() ([]*Session, error) {
	entries, err := os.ReadDir(ss.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Session{}, nil
		}
		return nil, fmt.Errorf("failed to read session directory: %w", err)
	}

	var sessions []*Session

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".json")
		session, err := ss.Load(id)
		if err != nil {
			// Skip corrupted sessions
			continue
		}

		sessions = append(sessions, session)
	}

	// Sort by update time (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

// ListSummaries returns session summaries without loading full messages
func (ss *SessionStore) ListSummaries() ([]SessionSummary, error) {
	sessions, err := ss.List()
	if err != nil {
		return nil, err
	}

	summaries := make([]SessionSummary, len(sessions))
	for i, session := range sessions {
		summaries[i] = SessionSummary{
			ID:           session.ID,
			Name:         session.Name,
			CreatedAt:    session.CreatedAt,
			UpdatedAt:    session.UpdatedAt,
			MessageCount: len(session.Messages),
			TotalTokens:  session.Metadata.TotalTokens,
			Preview:      getSessionPreview(session),
		}
	}

	return summaries, nil
}

// SessionSummary is a lightweight session representation for listing
type SessionSummary struct {
	ID           string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	MessageCount int
	TotalTokens  int
	Preview      string // First user message or title
}

// Search searches sessions by name or content
func (ss *SessionStore) Search(query string) ([]*Session, error) {
	sessions, err := ss.List()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var matches []*Session

	for _, session := range sessions {
		if strings.Contains(strings.ToLower(session.Name), query) {
			matches = append(matches, session)
			continue
		}

		// Search in messages
		for _, msg := range session.Messages {
			if strings.Contains(strings.ToLower(msg.Content), query) {
				matches = append(matches, session)
				break
			}
		}
	}

	return matches, nil
}

// sessionFilename returns the filename for a session
func (ss *SessionStore) sessionFilename(id string) string {
	return filepath.Join(ss.baseDir, id+".json")
}

// getSessionPreview extracts a preview from a session
func getSessionPreview(session *Session) string {
	if session.Name != "" && session.Name != session.ID {
		return session.Name
	}

	// Find first user message
	for _, msg := range session.Messages {
		if msg.Role == llm.RoleUser {
			preview := msg.Content
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			return preview
		}
	}

	return "Empty session"
}

// CreateSession creates a new session with a unique ID
func CreateSession(name string) *Session {
	id := generateSessionID()

	return &Session{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  make([]llm.Message, 0),
		Metadata: SessionMeta{
			Custom: make(map[string]string),
		},
	}
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	// Format: YYYYMMDD-HHMMSS-RANDOM
	now := time.Now()
	return fmt.Sprintf("%s-%06d",
		now.Format("20060102-150405"),
		now.UnixNano()%1000000,
	)
}

// SessionManager provides high-level session management
type SessionManager struct {
	store          *SessionStore
	currentSession *Session
	agent          *Agent
}

// NewSessionManager creates a new session manager
func NewSessionManager(store *SessionStore) *SessionManager {
	return &SessionManager{
		store: store,
	}
}

// SetAgent sets the agent to sync with
func (sm *SessionManager) SetAgent(agent *Agent) {
	sm.agent = agent
}

// NewSession creates and starts a new session
func (sm *SessionManager) NewSession(name string) (*Session, error) {
	session := CreateSession(name)
	sm.currentSession = session

	// Sync with agent context if available
	if sm.agent != nil {
		sm.agent.Reset()
	}

	return session, nil
}

// LoadSession loads and resumes a session
func (sm *SessionManager) LoadSession(id string) (*Session, error) {
	session, err := sm.store.Load(id)
	if err != nil {
		return nil, err
	}

	sm.currentSession = session

	// Sync with agent context if available
	if sm.agent != nil {
		sm.agent.Reset()
		for _, msg := range session.Messages {
			sm.agent.context.AddMessage(msg)
		}
	}

	return session, nil
}

// SaveCurrentSession saves the current session
func (sm *SessionManager) SaveCurrentSession() error {
	if sm.currentSession == nil {
		return fmt.Errorf("no active session")
	}

	// Sync messages from agent if available
	if sm.agent != nil {
		sm.currentSession.Messages = sm.agent.context.GetMessages()
	}

	return sm.store.Save(sm.currentSession)
}

// CurrentSession returns the current session
func (sm *SessionManager) CurrentSession() *Session {
	return sm.currentSession
}

// AddMessage adds a message to the current session
func (sm *SessionManager) AddMessage(msg llm.Message) {
	if sm.currentSession != nil {
		sm.currentSession.Messages = append(sm.currentSession.Messages, msg)
		sm.currentSession.UpdatedAt = time.Now()
	}
}

// AutoSave saves the session if enough time has passed since last save
func (sm *SessionManager) AutoSave(minInterval time.Duration) error {
	if sm.currentSession == nil {
		return nil
	}

	// Check if enough time has passed
	if time.Since(sm.currentSession.UpdatedAt) < minInterval {
		return nil
	}

	return sm.SaveCurrentSession()
}

// Close saves and closes the current session
func (sm *SessionManager) Close() error {
	if sm.currentSession == nil {
		return nil
	}

	err := sm.SaveCurrentSession()
	sm.currentSession = nil
	return err
}
