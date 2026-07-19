package event

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
)

const maxContentRunes = 180

// Notification is the normalized subset of Codex's task-completion payload
// used by the notifier. It deliberately excludes arbitrary event fields.
type Notification struct {
	Kind                 string
	ThreadID             string
	CWD                  string
	LastAssistantMessage string
	InputMessages        []string
	IsSubagent           bool
}

type rawEvent struct {
	Type                 string          `json:"type"`
	HookEventName        string          `json:"hook_event_name"`
	SessionID            string          `json:"session_id"`
	ThreadID             string          `json:"thread-id"`
	ThreadIDAlt          string          `json:"thread_id"`
	CWD                  string          `json:"cwd"`
	LastAssistantMessage string          `json:"last-assistant-message"`
	LastAssistantAlt     string          `json:"last_assistant_message"`
	InputMessages        []string        `json:"input-messages"`
	InputMessagesAlt     []string        `json:"input_messages"`
	ParentThreadID       string          `json:"parent_thread_id"`
	ThreadSource         string          `json:"thread_source"`
	Source               json.RawMessage `json:"source"`
}

type sourceMetadata struct {
	Subagent bool `json:"subagent"`
}

// Parse normalizes the two Codex notification payload forms used by the
// current and previous desktop integrations.
func Parse(raw string) (Notification, error) {
	var input rawEvent
	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		return Notification{}, errors.New("invalid Codex event JSON")
	}

	kind := strings.TrimSpace(input.HookEventName)
	if kind == "" {
		kind = strings.TrimSpace(input.Type)
	}
	if kind != "Stop" && kind != "SubagentStop" && kind != "agent-turn-complete" {
		return Notification{}, errors.New("unsupported Codex event")
	}

	threadID := strings.TrimSpace(input.ThreadID)
	if threadID == "" {
		threadID = strings.TrimSpace(input.ThreadIDAlt)
	}
	if threadID == "" {
		threadID = strings.TrimSpace(input.SessionID)
	}

	lastMessage := input.LastAssistantMessage
	if lastMessage == "" {
		lastMessage = input.LastAssistantAlt
	}
	messages := input.InputMessages
	if len(messages) == 0 {
		messages = input.InputMessagesAlt
	}

	isSubagent := input.ParentThreadID != "" || input.ThreadSource == "subagent" || kind == "SubagentStop"
	if len(input.Source) > 0 {
		var source sourceMetadata
		if json.Unmarshal(input.Source, &source) == nil && source.Subagent {
			isSubagent = true
		}
	}

	return Notification{
		Kind:                 kind,
		ThreadID:             threadID,
		CWD:                  input.CWD,
		LastAssistantMessage: lastMessage,
		InputMessages:        messages,
		IsSubagent:           isSubagent,
	}, nil
}

// BuildContent produces privacy-minimal content for a Bark notification.
func BuildContent(notification Notification) (string, string) {
	cwd := strings.TrimRight(strings.TrimSpace(notification.CWD), `/\`)
	projectName := ""
	if cwd != "" {
		projectName = filepath.Base(cwd)
	}

	title := "Codex 任务已结束"
	if projectName != "" && projectName != "." && projectName != string(filepath.Separator) {
		title = "Codex - " + projectName
	}

	content := strings.TrimSpace(notification.LastAssistantMessage)
	prefix := "结果："
	if content == "" && len(notification.InputMessages) > 0 {
		content = strings.TrimSpace(notification.InputMessages[len(notification.InputMessages)-1])
		prefix = "任务："
	}
	if content == "" {
		return title, "请回来检查。"
	}

	runes := []rune(content)
	if len(runes) > maxContentRunes {
		runes = runes[:maxContentRunes-3]
		content = string(runes) + "..."
	}
	return title, prefix + content
}
