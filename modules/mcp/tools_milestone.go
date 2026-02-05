// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package mcp

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

func (r *Registry) registerMilestoneTools() {
	r.Register(ToolDef{
		Tool: Tool{
			Name:        "list_milestones",
			Description: "List milestones in a repository",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
					"state": {Type: "string", Description: "Filter by state", Enum: []string{"open", "closed", "all"}},
					"page":  {Type: "integer", Description: "Page number"},
					"limit": {Type: "integer", Description: "Page size"},
				},
			},
		},
		Handler: handleListMilestones,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "get_milestone",
			Description: "Get a single milestone by ID or name",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
					"id":    {Type: "string", Description: "Milestone ID or name"},
				},
				Required: []string{"id"},
			},
		},
		Handler: handleGetMilestone,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "create_milestone",
			Description: "Create a new milestone in a repository",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":       {Type: "string", Description: "Repository owner"},
					"repo":        {Type: "string", Description: "Repository name"},
					"title":       {Type: "string", Description: "Milestone title"},
					"description": {Type: "string", Description: "Milestone description"},
					"due_on":      {Type: "string", Description: "Due date (ISO 8601 format)"},
					"state":       {Type: "string", Description: "Milestone state", Enum: []string{"open", "closed"}},
				},
				Required: []string{"title"},
			},
		},
		Handler: handleCreateMilestone,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "edit_milestone",
			Description: "Edit an existing milestone",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":       {Type: "string", Description: "Repository owner"},
					"repo":        {Type: "string", Description: "Repository name"},
					"id":          {Type: "string", Description: "Milestone ID or name"},
					"title":       {Type: "string", Description: "New title"},
					"description": {Type: "string", Description: "New description"},
					"due_on":      {Type: "string", Description: "New due date (ISO 8601 format)"},
					"state":       {Type: "string", Description: "New state", Enum: []string{"open", "closed"}},
				},
				Required: []string{"id"},
			},
		},
		Handler: handleEditMilestone,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "delete_milestone",
			Description: "Delete a milestone from a repository",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
					"id":    {Type: "string", Description: "Milestone ID or name"},
				},
				Required: []string{"id"},
			},
		},
		Handler: handleDeleteMilestone,
	})
}

func handleListMilestones(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	if v := stringParam(params, "state"); v != "" {
		query.Set("state", v)
	}
	if v := intParam(params, "page"); v > 0 {
		query.Set("page", strconv.FormatInt(v, 10))
	}
	if v := intParam(params, "limit"); v > 0 {
		query.Set("limit", strconv.FormatInt(v, 10))
	}

	return client.Get(fmt.Sprintf("/repos/%s/%s/milestones", owner, repo), query)
}

func handleGetMilestone(client *Client, params map[string]any) (any, error) {
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

	return client.Get(fmt.Sprintf("/repos/%s/%s/milestones/%s", owner, repo, id), nil)
}

func handleCreateMilestone(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	title := stringParam(params, "title")
	if title == "" {
		return nil, errors.New("title is required")
	}

	body := map[string]any{"title": title}
	if v := stringParam(params, "description"); v != "" {
		body["description"] = v
	}
	if v := stringParam(params, "due_on"); v != "" {
		body["due_on"] = v
	}
	if v := stringParam(params, "state"); v != "" {
		body["state"] = v
	}

	return client.Post(fmt.Sprintf("/repos/%s/%s/milestones", owner, repo), body)
}

func handleEditMilestone(client *Client, params map[string]any) (any, error) {
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

	body := map[string]any{}
	if v := stringParam(params, "title"); v != "" {
		body["title"] = v
	}
	if v := stringParam(params, "description"); v != "" {
		body["description"] = v
	}
	if v := stringParam(params, "due_on"); v != "" {
		body["due_on"] = v
	}
	if v := stringParam(params, "state"); v != "" {
		body["state"] = v
	}

	return client.Patch(fmt.Sprintf("/repos/%s/%s/milestones/%s", owner, repo, id), body)
}

func handleDeleteMilestone(client *Client, params map[string]any) (any, error) {
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

	err = client.Delete(fmt.Sprintf("/repos/%s/%s/milestones/%s", owner, repo, id))
	if err != nil {
		return nil, err
	}

	return map[string]string{"status": "deleted"}, nil
}
