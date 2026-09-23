package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const testToken = "test-token"

// fakeUpstream is an MCP endpoint that requires the bearer token and records
// every request it receives.
type fakeUpstream struct {
	srv      *httptest.Server
	requests atomic.Int64
	mu       sync.Mutex
	auth     []string
}

func (f *fakeUpstream) authHeaders() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.auth...)
}

func newFakeUpstream(t *testing.T, tools []*mcp.Tool, h mcp.ToolHandler) *fakeUpstream {
	t.Helper()
	server := mcp.NewServer(&mcp.Implementation{Name: "upstream", Version: "1"}, nil)
	for _, tool := range tools {
		server.AddTool(tool, h)
	}
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{JSONResponse: true})
	f := &fakeUpstream{}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.requests.Add(1)
		f.mu.Lock()
		f.auth = append(f.auth, r.Header.Get("Authorization"))
		f.mu.Unlock()
		if r.Header.Get("Authorization") != "Bearer "+testToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(f.srv.Close)
	return f
}

// startRelay runs relay on one end of an in-memory transport pair and returns
// a host session connected to the other end.
func startRelay(t *testing.T, ctx context.Context, endpoint, token, ver string) *mcp.ClientSession {
	t.Helper()
	return startRelayWithProtocol(t, ctx, endpoint, token, ver, "")
}

// startRelayWithProtocol pins the host's protocol version. Parity checks pin it
// to the version the direct connection negotiated, because the result envelope
// differs between protocol versions while the forwarded content must not.
//
// The host handshake races against relay itself: if relay returns before the
// host is connected (for example because the upstream rejected it), nothing
// will ever answer initialize, so the test fails at once with relay's error
// instead of waiting for the context deadline.
func startRelayWithProtocol(t *testing.T, ctx context.Context, endpoint, token, ver, protocol string) *mcp.ClientSession {
	t.Helper()
	hostSide, relaySide := mcp.NewInMemoryTransports()
	relayErr := make(chan error, 1)
	go func() { relayErr <- relay(ctx, endpoint, token, ver, relaySide) }()
	var opts *mcp.ClientSessionOptions
	if protocol != "" {
		opts = &mcp.ClientSessionOptions{ProtocolVersion: protocol}
	}
	// The session stays bound to connectCtx, so it is cancelled only on the
	// failure path below or when the test ends.
	connectCtx, cancelConnect := context.WithCancel(ctx)
	t.Cleanup(cancelConnect)
	type connectResult struct {
		session *mcp.ClientSession
		err     error
	}
	connected := make(chan connectResult, 1)
	go func() {
		s, err := mcp.NewClient(&mcp.Implementation{Name: "host", Version: "1"}, nil).Connect(connectCtx, hostSide, opts)
		connected <- connectResult{s, err}
	}()
	select {
	case rerr := <-relayErr:
		cancelConnect()
		t.Fatalf("relay exited before the host connected: %v", rerr)
	case res := <-connected:
		if res.err != nil {
			select {
			case rerr := <-relayErr:
				t.Fatalf("relay failed: %v", rerr)
			default:
			}
			t.Fatalf("connect to relay: %v", res.err)
		}
		t.Cleanup(func() { res.session.Close() })
		return res.session
	}
	panic("unreachable")
}

func connectDirect(t *testing.T, ctx context.Context, endpoint string) *mcp.ClientSession {
	t.Helper()
	hc := &http.Client{Transport: bearerTransport{testToken, http.DefaultTransport}}
	session, err := mcp.NewClient(&mcp.Implementation{Name: "direct", Version: "1"}, nil).Connect(ctx, &mcp.StreamableClientTransport{Endpoint: endpoint, HTTPClient: hc, DisableStandaloneSSE: true}, nil)
	if err != nil {
		t.Fatalf("connect direct: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func assertJSONEqual(t *testing.T, label string, want, got any) {
	t.Helper()
	var w, g any
	if err := json.Unmarshal([]byte(mustJSON(t, want)), &w); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(mustJSON(t, got)), &g); err != nil {
		t.Fatal(err)
	}
	if mustJSON(t, w) != mustJSON(t, g) {
		t.Fatalf("%s differs\nwant: %s\ngot:  %s", label, mustJSON(t, w), mustJSON(t, g))
	}
}

// schemaTools covers the JSON Schema constructs the Custodexa tool surface uses.
func schemaTools() []*mcp.Tool {
	return []*mcp.Tool{
		{
			Name:        "required_and_enum",
			Description: "Line one.\nLine two with non-ASCII: 資產 アセット.",
			InputSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []any{"key", "count"},
				"properties": map[string]any{
					"key":   map[string]any{"type": "string", "enum": []any{"ctrl-c", "ctrl-d", "enter"}, "description": "Key name."},
					"count": map[string]any{"type": "integer", "minimum": 1, "maximum": 3600},
				},
			},
		},
		{
			Name:        "array_of_objects",
			Description: "Nested structures.",
			InputSchema: map[string]any{
				"type":     "object",
				"required": []any{"items"},
				"properties": map[string]any{
					"items": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type":                 "object",
							"additionalProperties": false,
							"required":             []any{"asset_id", "accounts"},
							"properties": map[string]any{
								"asset_id": map[string]any{"type": "integer"},
								"accounts": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
							},
						},
					},
					"meta": map[string]any{
						"type":       "object",
						"properties": map[string]any{"reason": map[string]any{"type": "string"}},
					},
				},
			},
			OutputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"status": map[string]any{"type": "string"}},
			},
		},
		{
			Name:        "error_result",
			Description: "Returns a tool error.",
			InputSchema: map[string]any{"type": "object"},
		},
	}
}

func echoHandler(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args map[string]any
	if len(req.Params.Arguments) > 0 {
		if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
			return nil, err
		}
	}
	switch req.Params.Name {
	case "error_result":
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "request_id is required"}}}, nil
	case "array_of_objects":
		return &mcp.CallToolResult{
			Content:           []mcp.Content{&mcp.TextContent{Text: `{"status":"queued"}`}},
			StructuredContent: map[string]any{"status": "queued", "echo": args},
		}, nil
	default:
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "ok: " + mustJSONNoT(args)}}}, nil
	}
}

func mustJSONNoT(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func TestRelayRejectsInvalidEndpointWithoutRequests(t *testing.T) {
	up := newFakeUpstream(t, schemaTools(), echoHandler)
	host := strings.TrimPrefix(up.srv.URL, "http://")
	cases := map[string]string{
		"scheme":   "ftp://" + host + "/mcp",
		"no host":  "http:///mcp",
		"userinfo": "http://user:secret@" + host + "/mcp",
		"query":    up.srv.URL + "/mcp?token=x",
		"fragment": up.srv.URL + "/mcp#frag",
	}
	for name, endpoint := range cases {
		t.Run(name, func(t *testing.T) {
			_, relaySide := mcp.NewInMemoryTransports()
			err := relay(testContext(t), endpoint, testToken, "1.0.0", relaySide)
			if err == nil || !strings.Contains(err.Error(), "CUSTODEXA_MCP_URL") {
				t.Fatalf("want endpoint validation error, got %v", err)
			}
		})
	}
	if n := up.requests.Load(); n != 0 {
		t.Fatalf("upstream received %d requests, want 0", n)
	}
}

func TestRelayRequiresTokenWithoutRequests(t *testing.T) {
	up := newFakeUpstream(t, schemaTools(), echoHandler)
	_, relaySide := mcp.NewInMemoryTransports()
	err := relay(testContext(t), up.srv.URL, "", "1.0.0", relaySide)
	if err == nil || !strings.Contains(err.Error(), "CUSTODEXA_AGENT_TOKEN") {
		t.Fatalf("want missing token error, got %v", err)
	}
	if n := up.requests.Load(); n != 0 {
		t.Fatalf("upstream received %d requests, want 0", n)
	}
}

func TestBlankTokenFromEnvIsMissingWithoutRequests(t *testing.T) {
	up := newFakeUpstream(t, schemaTools(), echoHandler)
	blank := func(string) string { return "  \t  " }
	_, relaySide := mcp.NewInMemoryTransports()
	err := relay(testContext(t), up.srv.URL, tokenFromEnv(blank), "1.0.0", relaySide)
	if err == nil || !strings.Contains(err.Error(), "CUSTODEXA_AGENT_TOKEN") {
		t.Fatalf("want missing token error, got %v", err)
	}
	if n := up.requests.Load(); n != 0 {
		t.Fatalf("upstream received %d requests, want 0", n)
	}
	if got := tokenFromEnv(func(string) string { return " " + testToken + "\n" }); got != testToken {
		t.Fatalf("trimmed: got %q", got)
	}
}

func TestEndpointFromEnv(t *testing.T) {
	env := func(v string) func(string) string { return func(string) string { return v } }
	if got := endpointFromEnv(env("")); got != defaultEndpoint {
		t.Fatalf("empty: got %q", got)
	}
	if got := endpointFromEnv(env("   ")); got != defaultEndpoint {
		t.Fatalf("blank: got %q", got)
	}
	if got := endpointFromEnv(env(" https://bastion.example/api/v1/mcp ")); got != "https://bastion.example/api/v1/mcp" {
		t.Fatalf("trimmed: got %q", got)
	}
}

func TestRelaySendsBearerOnEveryRequest(t *testing.T) {
	ctx := testContext(t)
	up := newFakeUpstream(t, schemaTools(), echoHandler)
	session := startRelay(t, ctx, up.srv.URL, testToken, "1.0.0")
	if _, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "error_result", Arguments: map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	auth := up.authHeaders()
	if len(auth) == 0 {
		t.Fatal("upstream saw no requests")
	}
	for i, h := range auth {
		if h != "Bearer "+testToken {
			t.Fatalf("request %d Authorization = %q", i, h)
		}
	}
}

func TestRelayDoesNotFollowRedirects(t *testing.T) {
	target := newFakeUpstream(t, schemaTools(), echoHandler)
	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.srv.URL, http.StatusTemporaryRedirect)
	}))
	t.Cleanup(redirector.Close)
	_, relaySide := mcp.NewInMemoryTransports()
	err := relay(testContext(t), redirector.URL, testToken, "1.0.0", relaySide)
	if err == nil {
		t.Fatal("want error when upstream redirects")
	}
	if n := target.requests.Load(); n != 0 {
		t.Fatalf("redirect target received %d requests, want 0", n)
	}
}

func TestRelayToolListMatchesUpstream(t *testing.T) {
	ctx := testContext(t)
	up := newFakeUpstream(t, schemaTools(), echoHandler)
	direct := connectDirect(t, ctx, up.srv.URL)
	relayed := startRelayWithProtocol(t, ctx, up.srv.URL, testToken, "1.0.0", direct.InitializeResult().ProtocolVersion)
	want, err := direct.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := relayed.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(want.Tools) != len(schemaTools()) {
		t.Fatalf("upstream listed %d tools, want %d", len(want.Tools), len(schemaTools()))
	}
	assertJSONEqual(t, "tools/list", want.Tools, got.Tools)
}

func TestRelayCallResultsMatchUpstream(t *testing.T) {
	ctx := testContext(t)
	up := newFakeUpstream(t, schemaTools(), echoHandler)
	direct := connectDirect(t, ctx, up.srv.URL)
	relayed := startRelayWithProtocol(t, ctx, up.srv.URL, testToken, "1.0.0", direct.InitializeResult().ProtocolVersion)
	calls := []*mcp.CallToolParams{
		{Name: "required_and_enum", Arguments: map[string]any{"key": "enter", "count": 2}},
		{Name: "array_of_objects", Arguments: map[string]any{"items": []any{map[string]any{"asset_id": 7, "accounts": []any{"deploy"}}}, "meta": map[string]any{"reason": "資產"}}},
		{Name: "error_result", Arguments: map[string]any{}},
	}
	for _, p := range calls {
		want, err := direct.CallTool(ctx, p)
		if err != nil {
			t.Fatalf("%s direct: %v", p.Name, err)
		}
		got, err := relayed.CallTool(ctx, p)
		if err != nil {
			t.Fatalf("%s relayed: %v", p.Name, err)
		}
		if p.Name == "error_result" && !got.IsError {
			t.Fatal("error_result: isError not forwarded")
		}
		if p.Name == "array_of_objects" && got.StructuredContent == nil {
			t.Fatal("array_of_objects: structuredContent not forwarded")
		}
		assertJSONEqual(t, p.Name, want, got)
	}
}

func TestRelayPropagatesCancellation(t *testing.T) {
	ctx := testContext(t)
	started := make(chan struct{})
	cancelled := make(chan struct{})
	tools := []*mcp.Tool{{Name: "wait", InputSchema: map[string]any{"type": "object"}}}
	up := newFakeUpstream(t, tools, func(callCtx context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		close(started)
		<-callCtx.Done()
		close(cancelled)
		return nil, callCtx.Err()
	})
	session := startRelay(t, ctx, up.srv.URL, testToken, "1.0.0")
	callCtx, cancelCall := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() {
		_, err := session.CallTool(callCtx, &mcp.CallToolParams{Name: "wait", Arguments: map[string]any{}})
		done <- err
	}()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("upstream call never started")
	}
	cancelCall()
	select {
	case <-cancelled:
	case <-time.After(5 * time.Second):
		t.Fatal("upstream call was not cancelled")
	}
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Logf("host call returned %v", err)
	}
}

func TestRelayReportsOwnVersion(t *testing.T) {
	ctx := testContext(t)
	up := newFakeUpstream(t, schemaTools(), echoHandler)
	session := startRelay(t, ctx, up.srv.URL, testToken, "1.2.3")
	info := session.InitializeResult().ServerInfo
	if info.Name != "custodexa-mcp" || info.Version != "1.2.3" {
		t.Fatalf("serverInfo = %+v", info)
	}
}

func TestResolveVersion(t *testing.T) {
	cases := []struct{ injected, module, want string }{
		{"1.0.0", "", "1.0.0"},
		{"1.0.0", "v0.9.0", "1.0.0"},
		{"", "v1.0.0", "1.0.0"},
		{"", "(devel)", "devel"},
		{"", "", "devel"},
		{"", "v0.0.0-20260923000000-abcdef123456+dirty", "0.0.0-20260923000000-abcdef123456+dirty"},
	}
	for _, c := range cases {
		if got := resolveVersion(c.injected, c.module); got != c.want {
			t.Errorf("resolveVersion(%q, %q) = %q, want %q", c.injected, c.module, got, c.want)
		}
	}
}
