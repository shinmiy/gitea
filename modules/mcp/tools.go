// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package mcp

import (
	"errors"
	"fmt"

	json "code.gitea.io/gitea/modules/json"
)

// ToolHandler is a function that handles a tool call.
type ToolHandler func(client *Client, params map[string]any) (any, error)

// ToolDef groups a tool definition with its handler.
type ToolDef struct {
	Tool    Tool
	Handler ToolHandler
}

// Registry holds all registered tool definitions.
type Registry struct {
	tools []ToolDef
}

// NewRegistry creates a new tool registry and registers all tools.
func NewRegistry() *Registry {
	r := &Registry{}
	r.registerIssueTools()
	r.registerLabelTools()
	r.registerMilestoneTools()
	r.registerProjectTools()
	return r
}

// Register adds a tool definition to the registry.
func (r *Registry) Register(def ToolDef) {
	r.tools = append(r.tools, def)
}

// ListTools returns all registered tool definitions.
func (r *Registry) ListTools() []Tool {
	tools := make([]Tool, len(r.tools))
	for i, def := range r.tools {
		tools[i] = def.Tool
	}
	return tools
}

// Call dispatches a tool call to the appropriate handler.
func (r *Registry) Call(client *Client, name string, args map[string]any) (*ToolResult, error) {
	for _, def := range r.tools {
		if def.Tool.Name == name {
			result, err := def.Handler(client, args)
			if err != nil {
				return &ToolResult{
					Content: []Content{{Type: "text", Text: "Error: " + err.Error()}},
					IsError: true,
				}, nil
			}
			text, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return &ToolResult{
					Content: []Content{{Type: "text", Text: "Error marshaling result: " + err.Error()}},
					IsError: true,
				}, nil
			}
			return &ToolResult{
				Content: []Content{{Type: "text", Text: string(text)}},
			}, nil
		}
	}
	return nil, fmt.Errorf("unknown tool: %s", name)
}

// Param helpers

func stringParam(params map[string]any, key string) string {
	if v, ok := params[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func intParam(params map[string]any, key string) int64 {
	if v, ok := params[key]; ok {
		if n, ok := v.(float64); ok {
			return int64(n)
		}
	}
	return 0
}

func stringSliceParam(params map[string]any, key string) []string {
	if v, ok := params[key]; ok {
		if arr, ok := v.([]any); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}

func intSliceParam(params map[string]any, key string) []int64 {
	if v, ok := params[key]; ok {
		if arr, ok := v.([]any); ok {
			result := make([]int64, 0, len(arr))
			for _, item := range arr {
				if n, ok := item.(float64); ok {
					result = append(result, int64(n))
				}
			}
			return result
		}
	}
	return nil
}

// resolveOwnerRepo returns owner and repo from the params map.
func resolveOwnerRepo(params map[string]any) (string, string, error) {
	owner := stringParam(params, "owner")
	repo := stringParam(params, "repo")
	if owner == "" || repo == "" {
		return "", "", errors.New("owner and repo are required (set via parameters or GITEA_OWNER/GITEA_REPO)")
	}
	return owner, repo, nil
}
