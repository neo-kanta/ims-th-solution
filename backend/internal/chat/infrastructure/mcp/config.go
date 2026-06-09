// Package mcp is the chat module's MCP client: it connects to one or more MCP
// servers (stdio or HTTP), aggregates their tools, and executes tool calls on
// the agent loop's behalf. It is the ONLY place in the chat module that knows
// about the MCP SDK. Business data is reached exclusively through here — the
// chat module never imports another module's internal packages.
package mcp

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Transport enumerates supported MCP transports.
type Transport string

const (
	TransportStdio Transport = "stdio"
	TransportHTTP  Transport = "http"
)

// ServerConfig declares one MCP server connection.
type ServerConfig struct {
	Name      string            `yaml:"name"`
	Transport Transport         `yaml:"transport"`
	Command   string            `yaml:"command"` // stdio: executable path
	Args      []string          `yaml:"args"`    // stdio: arguments
	Env       map[string]string `yaml:"env"`     // stdio: child environment
	URL       string            `yaml:"url"`     // http: base URL
	// Allow/Deny are tool-name glob patterns (e.g. "get_*", "*delete*").
	// A tool is exposed only if it matches an Allow pattern (when Allow is
	// non-empty) AND matches no Deny pattern. Read-only enforcement lives
	// here as defense-in-depth, independent of the server's own restraint.
	Allow []string `yaml:"allow"`
	Deny  []string `yaml:"deny"`
}

// Config is the parsed mcp-servers.yaml.
type Config struct {
	Servers []ServerConfig `yaml:"servers"`
}

// LoadConfig reads mcp-servers.yaml from path. A missing path is not an error
// — it returns an empty config so the caller can fall back to a default.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		return &Config{}, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read mcp config %q: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse mcp config %q: %w", path, err)
	}
	return &cfg, nil
}

// DefaultIMSConfig builds a single-server stdio config pointing at the
// bundled ims-mcp binary. Used when no YAML config file is present so the
// PoC works out of the box.
//
// binPath is the path to the ims-mcp executable; apiBaseURL is forwarded to
// it as IMS_API_BASE_URL so the server knows where to reach the REST API.
func DefaultIMSConfig(binPath, apiBaseURL string) *Config {
	return &Config{
		Servers: []ServerConfig{
			{
				Name:      "ims",
				Transport: TransportStdio,
				Command:   binPath,
				Env:       map[string]string{"IMS_API_BASE_URL": apiBaseURL},
				// Read-only allowlist: list_*, get_*, and calc_* (deterministic
				// read-only calculation tools) only. Everything that could
				// mutate is denied outright, belt-and-suspenders with the server
				// only registering read/calc wrappers.
				Allow: []string{"list_*", "get_*", "calc_*"},
				Deny: []string{
					"*create*", "*update*", "*delete*", "*post*",
					"*submit*", "*approve*", "*reverse*", "*cancel*",
					"*transfer*", "*write*", "*set_*",
				},
			},
		},
	}
}
