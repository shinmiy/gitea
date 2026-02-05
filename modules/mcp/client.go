// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package mcp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	json "code.gitea.io/gitea/modules/json"
)

// Client is an HTTP client for the Gitea REST API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Gitea API client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{},
	}
}

// Get performs a GET request to the given API path with optional query parameters.
func (c *Client) Get(path string, query url.Values) (any, error) {
	return c.do(http.MethodGet, path, query, nil)
}

// Post performs a POST request to the given API path with a JSON body.
func (c *Client) Post(path string, body any) (any, error) {
	return c.do(http.MethodPost, path, nil, body)
}

// Patch performs a PATCH request to the given API path with a JSON body.
func (c *Client) Patch(path string, body any) (any, error) {
	return c.do(http.MethodPatch, path, nil, body)
}

// Delete performs a DELETE request to the given API path.
func (c *Client) Delete(path string) error {
	_, err := c.do(http.MethodDelete, path, nil, nil)
	return err
}

func (c *Client) do(method, path string, query url.Values, body any) (any, error) {
	u := c.baseURL + "/api/v1" + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "token "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	// DELETE with 204 No Content returns no body
	if resp.StatusCode == http.StatusNoContent || len(respBody) == 0 {
		return nil, nil
	}

	var result any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return result, nil
}
