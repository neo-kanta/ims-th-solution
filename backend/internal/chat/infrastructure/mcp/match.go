package mcp

import "strings"

// matchGlob reports whether name matches a simple glob pattern supporting '*'
// as a wildcard for any run of characters. Matching is case-sensitive on the
// tool name, which is conventionally lower_snake_case. Patterns are lowered
// so "GET_*" still matches "get_x".
func matchGlob(pattern, name string) bool {
	pattern = strings.ToLower(pattern)
	name = strings.ToLower(name)

	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == name // no wildcard → exact match
	}

	// Anchor the first segment.
	if parts[0] != "" {
		if !strings.HasPrefix(name, parts[0]) {
			return false
		}
		name = name[len(parts[0]):]
	}

	// Anchor the last segment.
	last := parts[len(parts)-1]
	if last != "" {
		if !strings.HasSuffix(name, last) {
			return false
		}
		name = name[:len(name)-len(last)]
	}

	// Each middle segment must appear in order.
	for _, seg := range parts[1 : len(parts)-1] {
		if seg == "" {
			continue
		}
		idx := strings.Index(name, seg)
		if idx < 0 {
			return false
		}
		name = name[idx+len(seg):]
	}
	return true
}

// toolAllowed applies the allow/deny policy for a server. Deny wins. When
// Allow is empty, all non-denied tools are allowed.
func toolAllowed(cfg ServerConfig, toolName string) bool {
	for _, d := range cfg.Deny {
		if matchGlob(d, toolName) {
			return false
		}
	}
	if len(cfg.Allow) == 0 {
		return true
	}
	for _, a := range cfg.Allow {
		if matchGlob(a, toolName) {
			return true
		}
	}
	return false
}
