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

        // If SSE event was "error" and chunk type not set, set it
        if chunk.Type == "" && eventType == "error" {
            chunk.Type = "error"
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
