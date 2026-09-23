// Command custodexa-mcp relays MCP tool calls from a stdio host to the
// Custodexa MCP endpoint. All authorization, auditing and masking happen on
// the Custodexa server; this program only forwards.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const defaultEndpoint = "http://localhost:8080/api/v1/mcp"

func endpointFromEnv(getenv func(string) string) string {
	if e := strings.TrimSpace(getenv("CUSTODEXA_MCP_URL")); e != "" {
		return e
	}
	return defaultEndpoint
}

// tokenFromEnv trims the token like the endpoint, so a blank value counts as missing.
func tokenFromEnv(getenv func(string) string) string {
	return strings.TrimSpace(getenv("CUSTODEXA_AGENT_TOKEN"))
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := relay(ctx, endpointFromEnv(os.Getenv), tokenFromEnv(os.Getenv), buildVersion(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
