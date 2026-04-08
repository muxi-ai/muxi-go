package muxi

import (
	"strings"
	"testing"
)

func collectChatChunks(t *testing.T, stream string) ([]ChatChunk, error) {
	t.Helper()

	out := make(chan ChatChunk, 8)
	err := parseChatSSE(strings.NewReader(stream), out)
	close(out)

	var chunks []ChatChunk
	for chunk := range out {
		chunks = append(chunks, chunk)
	}

	return chunks, err
}

func TestParseChatSSEIgnoresKeepalivesAndSurfacesDone(t *testing.T) {
	chunks, err := collectChatChunks(t, ""+
		": keepalive\n\n"+
		"event: planning\n"+
		"data: {\"steps\":[\"inspect\",\"respond\"]}\n\n"+
		"event: done\n\n",
	)
	if err != nil {
		t.Fatalf("parseChatSSE returned error: %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("len(chunks) = %d, want 2", len(chunks))
	}
	if chunks[0].Type != "planning" {
		t.Fatalf("chunks[0].Type = %q, want planning", chunks[0].Type)
	}
	steps, ok := chunks[0].Raw["steps"].([]interface{})
	if !ok || len(steps) != 2 {
		t.Fatalf("chunks[0].Raw[steps] = %#v, want two-step payload", chunks[0].Raw["steps"])
	}
	if chunks[1].Type != "done" {
		t.Fatalf("chunks[1].Type = %q, want done", chunks[1].Type)
	}
}

func TestParseChatSSESurfacesRouteErrors(t *testing.T) {
	chunks, err := collectChatChunks(t, ""+
		": keepalive\n\n"+
		"event: error\n"+
		"data: {\"error\":\"boom\",\"type\":\"RUNTIME_ERROR\"}\n\n",
	)
	if err == nil {
		t.Fatalf("expected route-level error, got nil")
	}
	if len(chunks) != 0 {
		t.Fatalf("len(chunks) = %d, want 0", len(chunks))
	}

	muxiErr, ok := err.(*MuxiError)
	if !ok {
		t.Fatalf("err = %T, want *MuxiError", err)
	}
	if muxiErr.Code != "RUNTIME_ERROR" {
		t.Fatalf("err.Code = %q, want RUNTIME_ERROR", muxiErr.Code)
	}
	if muxiErr.Message != "boom" {
		t.Fatalf("err.Message = %q, want boom", muxiErr.Message)
	}
}

func TestParseChatSSEPreservesUnknownChunkTypes(t *testing.T) {
	chunks, err := collectChatChunks(t, ""+
		"event: progress\n"+
		"data: {\"progress\":1,\n"+
		"data: \"message\":\"still-working\"}\n\n",
	)
	if err != nil {
		t.Fatalf("parseChatSSE returned error: %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("len(chunks) = %d, want 2", len(chunks))
	}
	if chunks[0].Type != "progress" {
		t.Fatalf("chunks[0].Type = %q, want progress", chunks[0].Type)
	}
	if _, ok := chunks[0].Raw["progress"].(float64); !ok {
		t.Fatalf("chunks[0].Raw[progress] = %#v, want numeric progress", chunks[0].Raw["progress"])
	}
	if got, ok := chunks[0].Raw["message"].(string); !ok || got != "still-working" {
		t.Fatalf("chunks[0].Raw[message] = %#v, want still-working", chunks[0].Raw["message"])
	}
	if chunks[1].Type != "done" {
		t.Fatalf("chunks[1].Type = %q, want done", chunks[1].Type)
	}
}
