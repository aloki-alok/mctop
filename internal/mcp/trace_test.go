package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestTraceOverStreamableHTTP is the load-bearing check for the trace feature:
// wrapping the transport in a LoggingTransport must not break a real multi-call
// streamable-HTTP session, and every frame must land in the recorder.
func TestTraceOverStreamableHTTP(t *testing.T) {
	srv := sdk.NewServer(&sdk.Implementation{Name: "test", Version: "1"}, nil)
	sdk.AddTool(srv, &sdk.Tool{Name: "ping", Description: "reply ok"},
		func(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, any, error) {
			return &sdk.CallToolResult{Content: []sdk.Content{&sdk.TextContent{Text: "ok"}}}, nil, nil
		})
	handler := sdk.NewStreamableHTTPHandler(func(*http.Request) *sdk.Server { return srv }, nil)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := Connect(ctx, ts.URL, Options{})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer c.Close()

	tools, err := c.Tools(ctx)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(tools) != 1 || tools[0].Name != "ping" {
		t.Fatalf("tools = %v, want [ping]", tools)
	}
	res, err := c.Call(ctx, "ping", map[string]any{})
	if err != nil {
		t.Fatalf("call after a wrapped multi-step session failed: %v", err)
	}
	if res.IsError {
		t.Fatalf("call returned isError")
	}

	frames := c.Trace()
	var sawInit, sawList, sawCall, sawRecv bool
	for _, f := range frames {
		switch f.Dir {
		case Received:
			sawRecv = true
		case Sent:
			switch {
			case strings.Contains(f.Data, "initialize"):
				sawInit = true
			case strings.Contains(f.Data, "tools/list"):
				sawList = true
			case strings.Contains(f.Data, "tools/call"):
				sawCall = true
			}
		}
	}
	if !sawInit || !sawList || !sawCall || !sawRecv {
		t.Fatalf("trace missing frames: init=%v list=%v call=%v recv=%v (have %d frames)",
			sawInit, sawList, sawCall, sawRecv, len(frames))
	}
}
