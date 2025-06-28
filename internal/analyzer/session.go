package analyzer

import (
	"loglion/internal/parser"
	"time"
)

type Session struct {
	ID             string
	Events         []*parser.LogEntry
	StartTime      time.Time
	LastEventTime  time.Time
	CompletedSteps []string
	IsComplete     bool
}

type SessionManager struct {
	sessions       map[string]*Session
	sessionKey     string
	timeoutMinutes int
}

func NewSessionManager(sessionKey string, timeoutMinutes int) *SessionManager {
	return &SessionManager{
		sessions:       make(map[string]*Session),
		sessionKey:     sessionKey,
		timeoutMinutes: timeoutMinutes,
	}
}

func (sm *SessionManager) AddEvent(entry *parser.LogEntry) {
	if entry.EventData == nil {
		return
	}

	sessionID, exists := entry.EventData[sm.sessionKey]
	if !exists {
		return
	}

	sessionIDStr, ok := sessionID.(string)
	if !ok {
		return
	}

	session, exists := sm.sessions[sessionIDStr]
	if !exists {
		session = &Session{
			ID:        sessionIDStr,
			Events:    []*parser.LogEntry{},
			StartTime: entry.Timestamp,
		}
		sm.sessions[sessionIDStr] = session
	}

	// Check if session has timed out
	if sm.isSessionTimedOut(session, entry.Timestamp) {
		// Start new session with same ID
		session = &Session{
			ID:        sessionIDStr,
			Events:    []*parser.LogEntry{},
			StartTime: entry.Timestamp,
		}
		sm.sessions[sessionIDStr] = session
	}

	session.Events = append(session.Events, entry)
	session.LastEventTime = entry.Timestamp
}

func (sm *SessionManager) isSessionTimedOut(session *Session, currentTime time.Time) bool {
	timeout := time.Duration(sm.timeoutMinutes) * time.Minute
	return currentTime.Sub(session.LastEventTime) > timeout
}

func (sm *SessionManager) GetSessions() map[string]*Session {
	return sm.sessions
}

func (sm *SessionManager) GetSessionCount() int {
	return len(sm.sessions)
}
