package analyzer

import (
	"regexp"
	"loglion/internal/config"
	"loglion/internal/parser"
)

type FunnelAnalyzer struct {
	config         *config.Config
	sessionManager *SessionManager
}

type FunnelResult struct {
	FunnelName      string           `json:"funnel_name"`
	TotalSessions   int              `json:"total_sessions"`
	CompletedFunnels int             `json:"completed_funnels"`
	CompletionRate  float64          `json:"completion_rate"`
	Steps           []StepResult     `json:"steps"`
	Sessions        []SessionResult  `json:"sessions"`
}

type StepResult struct {
	Name           string  `json:"name"`
	Completed      int     `json:"completed"`
	CompletionRate float64 `json:"completion_rate"`
}

type SessionResult struct {
	SessionID       string  `json:"session_id"`
	Completed       bool    `json:"completed"`
	StepsCompleted  int     `json:"steps_completed"`
	DurationMinutes float64 `json:"duration_minutes"`
}

func NewFunnelAnalyzer(cfg *config.Config) *FunnelAnalyzer {
	sessionManager := NewSessionManager(cfg.Funnel.SessionKey, cfg.Funnel.TimeoutMinutes)
	
	return &FunnelAnalyzer{
		config:         cfg,
		sessionManager: sessionManager,
	}
}

func (fa *FunnelAnalyzer) ProcessLogEntries(entries []*parser.LogEntry) {
	for _, entry := range entries {
		fa.sessionManager.AddEvent(entry)
	}
}

func (fa *FunnelAnalyzer) AnalyzeFunnel() *FunnelResult {
	sessions := fa.sessionManager.GetSessions()
	totalSessions := len(sessions)
	
	if totalSessions == 0 {
		return &FunnelResult{
			FunnelName:    fa.config.Funnel.Name,
			TotalSessions: 0,
		}
	}
	
	stepResults := make([]StepResult, len(fa.config.Funnel.Steps))
	sessionResults := make([]SessionResult, 0, totalSessions)
	completedFunnels := 0
	
	// Initialize step results
	for i, step := range fa.config.Funnel.Steps {
		stepResults[i] = StepResult{
			Name:      step.Name,
			Completed: 0,
		}
	}
	
	// Analyze each session
	for _, session := range sessions {
		sessionResult := fa.analyzeSession(session)
		sessionResults = append(sessionResults, sessionResult)
		
		if sessionResult.Completed {
			completedFunnels++
		}
		
		// Update step completion counts
		for i := 0; i < sessionResult.StepsCompleted; i++ {
			stepResults[i].Completed++
		}
	}
	
	// Calculate completion rates
	for i := range stepResults {
		stepResults[i].CompletionRate = float64(stepResults[i].Completed) / float64(totalSessions)
	}
	
	completionRate := float64(completedFunnels) / float64(totalSessions)
	
	return &FunnelResult{
		FunnelName:       fa.config.Funnel.Name,
		TotalSessions:    totalSessions,
		CompletedFunnels: completedFunnels,
		CompletionRate:   completionRate,
		Steps:            stepResults,
		Sessions:         sessionResults,
	}
}

func (fa *FunnelAnalyzer) analyzeSession(session *Session) SessionResult {
	stepsCompleted := 0
	
	// Check each step in order
	for _, step := range fa.config.Funnel.Steps {
		if fa.sessionCompletedStep(session, step) {
			stepsCompleted++
		} else {
			break // Funnel steps must be completed in order
		}
	}
	
	isComplete := stepsCompleted == len(fa.config.Funnel.Steps)
	duration := session.LastEventTime.Sub(session.StartTime).Minutes()
	
	return SessionResult{
		SessionID:       session.ID,
		Completed:       isComplete,
		StepsCompleted:  stepsCompleted,
		DurationMinutes: duration,
	}
}

func (fa *FunnelAnalyzer) sessionCompletedStep(session *Session, step config.Step) bool {
	eventRegex, err := regexp.Compile(step.EventPattern)
	if err != nil {
		return false
	}
	
	for _, event := range session.Events {
		if event.EventData == nil {
			continue
		}
		
		// Check if event matches the pattern
		if eventRegex.MatchString(event.Message) {
			// Check required properties
			if fa.checkRequiredProperties(event.EventData, step.RequiredProperties) {
				return true
			}
		}
	}
	
	return false
}

func (fa *FunnelAnalyzer) checkRequiredProperties(eventData map[string]interface{}, requiredProps map[string]string) bool {
	for key, pattern := range requiredProps {
		value, exists := eventData[key]
		if !exists {
			return false
		}
		
		valueStr, ok := value.(string)
		if !ok {
			return false
		}
		
		matched, err := regexp.MatchString(pattern, valueStr)
		if err != nil || !matched {
			return false
		}
	}
	
	return true
}