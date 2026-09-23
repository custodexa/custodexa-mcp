package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const serverName = "custodexa-mcp"

// bearerTransport attaches the agent token to every upstream request.
type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (b bearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	q := r.Clone(r.Context())
	q.Header.Set("Authorization", "Bearer "+b.token)
	return b.base.RoundTrip(q)
}

// validateEndpoint accepts only a plain HTTP(S) URL: credentials, query and
// fragment are rejected so the token is the only secret on the wire.
func validateEndpoint(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("CUSTODEXA_MCP_URL must be an HTTP(S) endpoint without credentials, query or fragment")
	}
	return nil
}

// relay connects to the upstream MCP endpoint, mirrors its tool list and
// serves it on transport, forwarding every call unchanged.
func relay(ctx context.Context, endpoint, token, version string, transport mcp.Transport) error {
	if err := validateEndpoint(endpoint); err != nil {
		return err
	}
	if token == "" {
		return errors.New("CUSTODEXA_AGENT_TOKEN is required")
	}
	hc := &http.Client{
		Transport: bearerTransport{token, http.DefaultTransport},
		// Never follow redirects: the bearer token must not reach another origin.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	impl := &mcp.Implementation{Name: serverName, Version: version}
	upstream, err := mcp.NewClient(impl, nil).Connect(ctx, &mcp.StreamableClientTransport{Endpoint: endpoint, HTTPClient: hc, DisableStandaloneSSE: true}, nil)
	if err != nil {
		return fmt.Errorf("connect to MCP service: %w", err)
	}
	defer upstream.Close()
	server := mcp.NewServer(impl, nil)
	for tool, err := range upstream.Tools(ctx, nil) {
		if err != nil {
			return fmt.Errorf("list remote tools: %w", err)
		}
		if err := addTool(server, tool, func(callCtx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Each transport negotiates its own protocol; forward cancellation, not downstream session context.
			requestCtx, cancel := context.WithCancel(ctx)
			stop := context.AfterFunc(callCtx, cancel)
			defer stop()
			defer cancel()
			return upstream.CallTool(requestCtx, &mcp.CallToolParams{Name: req.Params.Name, Arguments: req.Params.Arguments})
		}); err != nil {
			return err
		}
	}
	return server.Run(ctx, transport)
}

// addTool registers an upstream tool, turning the SDK's panic on a malformed
// schema into an error so the process exits with a readable message.
func addTool(server *mcp.Server, tool *mcp.Tool, h mcp.ToolHandler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("register remote tool %q: %v", tool.Name, r)
		}
	}()
	server.AddTool(tool, h)
	return nil
}
