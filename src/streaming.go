package muxi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// parseChatSSE parses SSE stream into chat chunks
// Emits a final chunk {Type: "done"} when stream ends cleanly
func parseChatSSE(r io.Reader, out chan<- ChatChunk) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 256*1024), 10*1024*1024)
	var dataBuf []string
	var eventType string

	flush := func() error {
		if len(dataBuf) == 0 {
			return nil
		}
		payload := strings.Join(dataBuf, "")
		dataBuf = dataBuf[:0]

		var chunk ChatChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return fmt.Errorf("failed to parse chunk: %w", err)
		}

		// If SSE event defined a type and payload did not, use it
		if chunk.Type == "" && eventType != "" {
			chunk.Type = eventType
		}

		out <- chunk
		eventType = ""
		return nil
	}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}

		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			dataBuf = append(dataBuf, data)
		}

		// empty line ends event
		if line == "" {
			if err := flush(); err != nil {
				return err
			}
		}
	}

	// flush remaining
	if err := flush(); err != nil {
		return err
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// signal completion
	out <- ChatChunk{Type: "done"}
	return nil
}

// parseDeploySSE parses deploy/update/start/restart/rollback SSE
func parseDeploySSE(r io.Reader, out chan<- DeployEvent) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 256*1024), 10*1024*1024)
	var dataBuf []string
	var eventType string

	flush := func() error {
		if len(dataBuf) == 0 {
			return nil
		}
		payload := strings.Join(dataBuf, "")
		dataBuf = dataBuf[:0]

		switch eventType {
		case "progress":
			var p DeployProgressEvent
			if err := json.Unmarshal([]byte(payload), &p); err != nil {
				return fmt.Errorf("failed to parse progress event: %w", err)
			}
			out <- DeployEvent{Progress: &p}
		case "complete":
			var c DeployCompleteEvent
			if err := json.Unmarshal([]byte(payload), &c); err != nil {
				return fmt.Errorf("failed to parse complete event: %w", err)
			}
			out <- DeployEvent{Complete: &c}
		case "error":
			var e DeployErrorEvent
			if err := json.Unmarshal([]byte(payload), &e); err != nil {
				return fmt.Errorf("failed to parse error event: %w", err)
			}
			out <- DeployEvent{Error: &e}
		default:
			// unknown event type; ignore
		}

		eventType = ""
		return nil
	}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}

		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			dataBuf = append(dataBuf, data)
		}

		if line == "" {
			if err := flush(); err != nil {
				return err
			}
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

// parseLogSSE parses log streaming events
func parseLogSSE(r io.Reader, out chan<- LogEvent) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 256*1024), 10*1024*1024)
	var dataBuf []string
	var eventType string

	flush := func() error {
		if len(dataBuf) == 0 {
			return nil
		}
		payload := strings.Join(dataBuf, "")
		dataBuf = dataBuf[:0]

		if eventType == "log" || eventType == "" {
			var ev LogEvent
			if err := json.Unmarshal([]byte(payload), &ev); err != nil {
				return fmt.Errorf("failed to parse log event: %w", err)
			}
			out <- ev
		}
		eventType = ""
		return nil
	}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}

		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			dataBuf = append(dataBuf, data)
		}

		if line == "" {
			if err := flush(); err != nil {
				return err
			}
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
