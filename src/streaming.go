package muxi

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "strings"
)

// parseChatSSE parses SSE stream into chat chunks
func parseChatSSE(r io.Reader, out chan<- ChatChunk) error {
    scanner := bufio.NewScanner(r)
    var dataBuf []string

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
        out <- chunk
        return nil
    }

    for scanner.Scan() {
        line := scanner.Text()

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
    return nil
}
