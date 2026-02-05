// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package mcp

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

func (r *Registry) registerLabelTools() {
	r.Register(ToolDef{
		Tool: Tool{
			Name:        "list_labels",
			Description: "List labels in a repository",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
					"page":  {Type: "integer", Description: "Page number"},
					"limit": {Type: "integer", Description: "Page size"},
				},
			},
		},
		Handler: handleListLabels,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "get_label",
			Description: "Get a single label by ID or name",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
					"id":    {Type: "string", Description: "Label ID or name"},
				},
				Required: []string{"id"},
			},
		},
		Handler: handleGetLabel,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "create_label",
			Description: "Create a new label in a repository",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":       {Type: "string", Description: "Repository owner"},
					"repo":        {Type: "string", Description: "Repository name"},
					"name":        {Type: "string", Description: "Label name"},
					"color":       {Type: "string", Description: "Label color (hex code, e.g. '#00aabb')"},
					"description": {Type: "string", Description: "Label description"},
				},
				Required: []string{"name", "color"},
			},
		},
		Handler: handleCreateLabel,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "edit_label",
			Description: "Edit an existing label",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":       {Type: "string", Description: "Repository owner"},
					"repo":        {Type: "string", Description: "Repository name"},
					"id":          {Type: "integer", Description: "Label ID"},
					"name":        {Type: "string", Description: "New label name"},
					"color":       {Type: "string", Description: "New label color (hex code)"},
					"description": {Type: "string", Description: "New label description"},
				},
				Required: []string{"id"},
			},
		},
		Handler: handleEditLabel,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "delete_label",
			Description: "Delete a label from a repository",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
					"id":    {Type: "integer", Description: "Label ID"},
				},
				Required: []string{"id"},
			},
		},
		Handler: handleDeleteLabel,
	})
}

func handleListLabels(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	if v := intParam(params, "page"); v > 0 {
		query.Set("page", strconv.FormatInt(v, 10))
	}
	if v := intParam(params, "limit"); v > 0 {
		query.Set("limit", strconv.FormatInt(v, 10))
	}

	return client.Get(fmt.Sprintf("/repos/%s/%s/labels", owner, repo), query)
}

func handleGetLabel(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	id := stringParam(params, "id")
	if id == "" {
		if v := intParam(params, "id"); v > 0 {
			id = strconv.FormatInt(v, 10)
		}
	}
	if id == "" {
		return nil, errors.New("id is required")
	}

	return client.Get(fmt.Sprintf("/repos/%s/%s/labels/%s", owner, repo, id), nil)
}

func handleCreateLabel(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	name := stringParam(params, "name")
	color := stringParam(params, "color")
	if name == "" || color == "" {
		return nil, errors.New("name and color are required")
	}

	body := map[string]any{
		"name":  name,
		"color": color,
	}
	if v := stringParam(params, "description"); v != "" {
		body["description"] = v
	}

	return client.Post(fmt.Sprintf("/repos/%s/%s/labels", owner, repo), body)
}

func handleEditLabel(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	id := intParam(params, "id")
	if id == 0 {
		return nil, errors.New("id is required")
	}

	body := map[string]any{}
	if v := stringParam(params, "name"); v != "" {
		body["name"] = v
	}
	if v := stringParam(params, "color"); v != "" {
		body["color"] = v
	}
	if v := stringParam(params, "description"); v != "" {
		body["description"] = v
	}

	return client.Patch(fmt.Sprintf("/repos/%s/%s/labels/%d", owner, repo, id), body)
}

func handleDeleteLabel(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	id := intParam(params, "id")
	if id == 0 {
		return nil, errors.New("id is required")
	}

	err = client.Delete(fmt.Sprintf("/repos/%s/%s/labels/%d", owner, repo, id))
	if err != nil {
		return nil, err
	}

	return map[string]string{"status": "deleted"}, nil
}
