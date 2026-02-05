// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package mcp

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

func (r *Registry) registerProjectTools() {
	r.Register(ToolDef{
		Tool: Tool{
			Name:        "list_projects",
			Description: "List projects in a repository",
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
		Handler: handleListProjects,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "get_project",
			Description: "Get a single project by ID",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
					"id":    {Type: "integer", Description: "Project ID"},
				},
				Required: []string{"id"},
			},
		},
		Handler: handleGetProject,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "create_project",
			Description: "Create a new project in a repository",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":         {Type: "string", Description: "Repository owner"},
					"repo":          {Type: "string", Description: "Repository name"},
					"title":         {Type: "string", Description: "Project title"},
					"description":   {Type: "string", Description: "Project description"},
					"template_type": {Type: "integer", Description: "Project template type (0=none, 1=basic kanban, 2=bug triage)"},
					"card_type":     {Type: "integer", Description: "Card type (0=text only, 1=images and text)"},
				},
				Required: []string{"title"},
			},
		},
		Handler: handleCreateProject,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "edit_project",
			Description: "Edit an existing project",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":       {Type: "string", Description: "Repository owner"},
					"repo":        {Type: "string", Description: "Repository name"},
					"id":          {Type: "integer", Description: "Project ID"},
					"title":       {Type: "string", Description: "New title"},
					"description": {Type: "string", Description: "New description"},
					"card_type":   {Type: "integer", Description: "Card type (0=text only, 1=images and text)"},
					"state":       {Type: "string", Description: "New state", Enum: []string{"open", "closed"}},
				},
				Required: []string{"id"},
			},
		},
		Handler: handleEditProject,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "delete_project",
			Description: "Delete a project from a repository",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
					"id":    {Type: "integer", Description: "Project ID"},
				},
				Required: []string{"id"},
			},
		},
		Handler: handleDeleteProject,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "list_project_columns",
			Description: "List columns in a project board",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":      {Type: "string", Description: "Repository owner"},
					"repo":       {Type: "string", Description: "Repository name"},
					"project_id": {Type: "integer", Description: "Project ID"},
				},
				Required: []string{"project_id"},
			},
		},
		Handler: handleListProjectColumns,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "create_project_column",
			Description: "Create a new column in a project board",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":      {Type: "string", Description: "Repository owner"},
					"repo":       {Type: "string", Description: "Repository name"},
					"project_id": {Type: "integer", Description: "Project ID"},
					"title":      {Type: "string", Description: "Column title"},
					"color":      {Type: "string", Description: "Column color (hex code)"},
				},
				Required: []string{"project_id", "title"},
			},
		},
		Handler: handleCreateProjectColumn,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "edit_project_column",
			Description: "Edit an existing project board column",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":      {Type: "string", Description: "Repository owner"},
					"repo":       {Type: "string", Description: "Repository name"},
					"project_id": {Type: "integer", Description: "Project ID"},
					"column_id":  {Type: "integer", Description: "Column ID"},
					"title":      {Type: "string", Description: "New column title"},
					"color":      {Type: "string", Description: "New column color (hex code)"},
				},
				Required: []string{"project_id", "column_id"},
			},
		},
		Handler: handleEditProjectColumn,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "delete_project_column",
			Description: "Delete a column from a project board",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":      {Type: "string", Description: "Repository owner"},
					"repo":       {Type: "string", Description: "Repository name"},
					"project_id": {Type: "integer", Description: "Project ID"},
					"column_id":  {Type: "integer", Description: "Column ID"},
				},
				Required: []string{"project_id", "column_id"},
			},
		},
		Handler: handleDeleteProjectColumn,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "move_project_column",
			Description: "Reorder a column in a project board",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":      {Type: "string", Description: "Repository owner"},
					"repo":       {Type: "string", Description: "Repository name"},
					"project_id": {Type: "integer", Description: "Project ID"},
					"column_id":  {Type: "integer", Description: "Column ID"},
					"sorting":    {Type: "integer", Description: "New sort position"},
				},
				Required: []string{"project_id", "column_id", "sorting"},
			},
		},
		Handler: handleMoveProjectColumn,
	})
}

func handleListProjects(client *Client, params map[string]any) (any, error) {
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

	return client.Get(fmt.Sprintf("/repos/%s/%s/projects", owner, repo), query)
}

func handleGetProject(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	id := intParam(params, "id")
	if id == 0 {
		return nil, errors.New("id is required")
	}

	return client.Get(fmt.Sprintf("/repos/%s/%s/projects/%d", owner, repo, id), nil)
}

func handleCreateProject(client *Client, params map[string]any) (any, error) {
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
	if v, ok := params["template_type"]; ok {
		if n, ok := v.(float64); ok {
			body["template_type"] = int64(n)
		}
	}
	if v, ok := params["card_type"]; ok {
		if n, ok := v.(float64); ok {
			body["card_type"] = int64(n)
		}
	}

	return client.Post(fmt.Sprintf("/repos/%s/%s/projects", owner, repo), body)
}

func handleEditProject(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	id := intParam(params, "id")
	if id == 0 {
		return nil, errors.New("id is required")
	}

	body := map[string]any{}
	if v := stringParam(params, "title"); v != "" {
		body["title"] = v
	}
	if v := stringParam(params, "description"); v != "" {
		body["description"] = v
	}
	if v, ok := params["card_type"]; ok {
		if n, ok := v.(float64); ok {
			body["card_type"] = int64(n)
		}
	}
	if v := stringParam(params, "state"); v != "" {
		body["state"] = v
	}

	return client.Patch(fmt.Sprintf("/repos/%s/%s/projects/%d", owner, repo, id), body)
}

func handleDeleteProject(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	id := intParam(params, "id")
	if id == 0 {
		return nil, errors.New("id is required")
	}

	err = client.Delete(fmt.Sprintf("/repos/%s/%s/projects/%d", owner, repo, id))
	if err != nil {
		return nil, err
	}

	return map[string]string{"status": "deleted"}, nil
}

func handleListProjectColumns(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	projectID := intParam(params, "project_id")
	if projectID == 0 {
		return nil, errors.New("project_id is required")
	}

	return client.Get(fmt.Sprintf("/repos/%s/%s/projects/%d/columns", owner, repo, projectID), nil)
}

func handleCreateProjectColumn(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	projectID := intParam(params, "project_id")
	if projectID == 0 {
		return nil, errors.New("project_id is required")
	}

	title := stringParam(params, "title")
	if title == "" {
		return nil, errors.New("title is required")
	}

	body := map[string]any{"title": title}
	if v := stringParam(params, "color"); v != "" {
		body["color"] = v
	}

	return client.Post(fmt.Sprintf("/repos/%s/%s/projects/%d/columns", owner, repo, projectID), body)
}

func handleEditProjectColumn(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	projectID := intParam(params, "project_id")
	columnID := intParam(params, "column_id")
	if projectID == 0 || columnID == 0 {
		return nil, errors.New("project_id and column_id are required")
	}

	body := map[string]any{}
	if v := stringParam(params, "title"); v != "" {
		body["title"] = v
	}
	if v := stringParam(params, "color"); v != "" {
		body["color"] = v
	}

	return client.Patch(fmt.Sprintf("/repos/%s/%s/projects/%d/columns/%d", owner, repo, projectID, columnID), body)
}

func handleDeleteProjectColumn(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	projectID := intParam(params, "project_id")
	columnID := intParam(params, "column_id")
	if projectID == 0 || columnID == 0 {
		return nil, errors.New("project_id and column_id are required")
	}

	err = client.Delete(fmt.Sprintf("/repos/%s/%s/projects/%d/columns/%d", owner, repo, projectID, columnID))
	if err != nil {
		return nil, err
	}

	return map[string]string{"status": "deleted"}, nil
}

func handleMoveProjectColumn(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	projectID := intParam(params, "project_id")
	columnID := intParam(params, "column_id")
	if projectID == 0 || columnID == 0 {
		return nil, errors.New("project_id and column_id are required")
	}

	sorting := intParam(params, "sorting")

	body := map[string]any{"sorting": sorting}

	return client.Post(fmt.Sprintf("/repos/%s/%s/projects/%d/columns/%d/move", owner, repo, projectID, columnID), body)
}
