// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package mcp

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	json "code.gitea.io/gitea/modules/json"
)

// Server is an MCP server that communicates over stdio using JSON-RPC 2.0.
type Server struct {
	client       *Client
	registry     *Registry
	defaultOwner string
	defaultRepo  string
	reader       io.Reader
	writer       io.Writer
}

// NewServer creates a new MCP server.
func NewServer(client *Client, defaultOwner, defaultRepo string, reader io.Reader, writer io.Writer) *Server {
	return &Server{
		client:       client,
		registry:     NewRegistry(),
		defaultOwner: defaultOwner,
		defaultRepo:  defaultRepo,
		reader:       reader,
		writer:       writer,
	}
}

// Run starts the server's stdio read loop.
func (s *Server) Run() error {
	scanner := bufio.NewScanner(s.reader)
	// Allow large messages (up to 10MB)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.writeResponse(Response{
				JSONRPC: "2.0",
				Error:   &Error{Code: CodeParseError, Message: "Parse error"},
			})
			continue
		}

		resp := s.handleRequest(&req)
		if resp != nil {
			s.writeResponse(*resp)
		}
	}

	return scanner.Err()
}

func (s *Server) handleRequest(req *Request) *Response {
	// Notifications have no ID and expect no response
	if req.ID == nil {
		return nil
	}

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: CodeMethodNotFound, Message: "Method not found: " + req.Method},
		}
	}
}

func (s *Server) handleInitialize(req *Request) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: InitializeResult{
			ProtocolVersion: "2025-03-26",
			ServerInfo: ServerInfo{
				Name:    "gitea-mcp",
				Version: "0.1.0",
			},
			Capabilities: Capabilities{
				Tools: &ToolsCapability{},
			},
		},
	}
}

func (s *Server) handleToolsList(req *Request) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ToolsListResult{
			Tools: s.registry.ListTools(),
		},
	}
}

func (s *Server) handleToolsCall(req *Request) *Response {
	// Re-marshal params to decode as ToolCallParams
	paramsBytes, err := json.Marshal(req.Params)
	if err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: CodeInvalidParams, Message: "Invalid tool call params"},
		}
	}

	var callParams ToolCallParams
	if err := json.Unmarshal(paramsBytes, &callParams); err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: CodeInvalidParams, Message: "Invalid tool call params"},
		}
	}

	// Convert arguments to map and inject defaults
	args := s.resolveArguments(callParams.Arguments)

	result, err := s.registry.Call(s.client, callParams.Name, args)
	if err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &Error{Code: CodeInternalError, Message: err.Error()},
		}
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func (s *Server) resolveArguments(args any) map[string]any {
	params, ok := args.(map[string]any)
	if !ok {
		params = map[string]any{}
	}

	if _, ok := params["owner"]; !ok && s.defaultOwner != "" {
		params["owner"] = s.defaultOwner
	}
	if _, ok := params["repo"]; !ok && s.defaultRepo != "" {
		params["repo"] = s.defaultRepo
	}

	return params
}

func (s *Server) writeResponse(resp Response) {
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	fmt.Fprintf(s.writer, "%s\n", data)
}
