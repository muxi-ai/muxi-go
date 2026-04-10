package muxi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type sseEvent struct {
	Event string
	Data  string
}

// parseChatSSE parses SSE stream into chat chunks
// Emits a final chunk {Type: "done"} when stream ends cleanly
func parseChatSSE(r io.Reader, out chan<- ChatChunk) error {
	emittedDone := false
	if err := parseSSE(r, func(evt sseEvent) error {
		switch evt.Event {
		case "error":
			return parseRouteStreamError(evt.Data)
		case "done":
			if evt.Data == "" {
				out <- ChatChunk{Type: "done"}
				emittedDone = true
				return nil
			}
		}

		if evt.Data == "" {
			if evt.Event != "" {
				out <- ChatChunk{Type: evt.Event}
				if evt.Event == "done" {
					emittedDone = true
				}
			}
			return nil
		}

		var chunk ChatChunk
		if err := json.Unmarshal([]byte(evt.Data), &chunk); err != nil {
			return fmt.Errorf("failed to parse chunk: %w", err)
		}

		if chunk.Type == "" && evt.Event != "" {
			chunk.Type = evt.Event
		}
		if chunk.Type == "done" {
			emittedDone = true
		}

		out <- chunk
		return nil
	}); err != nil {
		return err
	}

	if !emittedDone {
		out <- ChatChunk{Type: "done"}
	}
	return nil
}

// parseDeploySSE parses deploy/update/start/restart/rollback SSE
func parseDeploySSE(r io.Reader, out chan<- DeployEvent) error {
	return parseSSE(r, func(evt sseEvent) error {
		switch evt.Event {
		case "progress":
			var p DeployProgressEvent
			if err := json.Unmarshal([]byte(evt.Data), &p); err != nil {
				return fmt.Errorf("failed to parse progress event: %w", err)
			}
			out <- DeployEvent{Progress: &p}
		case "complete":
			var c DeployCompleteEvent
			if err := json.Unmarshal([]byte(evt.Data), &c); err != nil {
				return fmt.Errorf("failed to parse complete event: %w", err)
			}
			out <- DeployEvent{Complete: &c}
		case "error":
			var e DeployErrorEvent
			if err := json.Unmarshal([]byte(evt.Data), &e); err != nil {
				return fmt.Errorf("failed to parse error event: %w", err)
			}
			out <- DeployEvent{Error: &e}
		default:
			// unknown event type; ignore
		}
		return nil
	})
}

// parseLogSSE parses log streaming events
func parseLogSSE(r io.Reader, out chan<- LogEvent) error {
	return parseSSE(r, func(evt sseEvent) error {
		if evt.Data == "" {
			return nil
		}
		if evt.Event == "log" || evt.Event == "" {
			var ev LogEvent
			if err := json.Unmarshal([]byte(evt.Data), &ev); err != nil {
				return fmt.Errorf("failed to parse log event: %w", err)
			}
			out <- ev
		}
		return nil
	})
}

func parseSSE(r io.Reader, onEvent func(sseEvent) error) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 256*1024), 10*1024*1024)
	var dataBuf []string
	var eventType string

	flush := func() error {
		if eventType == "" && len(dataBuf) == 0 {
			return nil
		}
		payload := strings.Join(dataBuf, "\n")
		dataBuf = dataBuf[:0]
		evt := sseEvent{Event: eventType, Data: payload}
		eventType = ""
		return onEvent(evt)
	}

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			if err := flush(); err != nil {
				return err
			}
			continue
		}

		if strings.HasPrefix(line, ":") {
			continue
		}

		field, value := parseSSEField(line)
		switch field {
		case "event":
			eventType = value
		case "data":
			dataBuf = append(dataBuf, value)
		}
	}

	if err := flush(); err != nil {
		return err
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func parseSSEField(line string) (string, string) {
	field, value, found := strings.Cut(line, ":")
	if !found {
		return line, ""
	}
	if strings.HasPrefix(value, " ") {
		value = value[1:]
	}
	return field, value
}

func parseRouteStreamError(payload string) error {
	details := map[string]interface{}{}
	if payload != "" {
		if err := json.Unmarshal([]byte(payload), &details); err == nil {
			code, _ := details["type"].(string)
			if code == "" {
				code, _ = details["code"].(string)
			}
			if code == "" {
				code = "STREAM_ERROR"
			}

			message, _ := details["error"].(string)
			if message == "" {
				message, _ = details["message"].(string)
			}
			if message == "" {
				message = "stream error"
			}

			return &MuxiError{Code: code, Message: message, StatusCode: 0, Details: details}
		}
	}

	message := payload
	if message == "" {
		message = "stream error"
	}
	return &MuxiError{Code: "STREAM_ERROR", Message: message, StatusCode: 0, Details: map[string]interface{}{"error": message}}
}
