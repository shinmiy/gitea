// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package mcp

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

func (r *Registry) registerIssueTools() {
	r.Register(ToolDef{
		Tool: Tool{
			Name:        "list_issues",
			Description: "List and search issues in a repository",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":     {Type: "string", Description: "Repository owner"},
					"repo":      {Type: "string", Description: "Repository name"},
					"state":     {Type: "string", Description: "Filter by state", Enum: []string{"open", "closed", "all"}},
					"labels":    {Type: "string", Description: "Comma-separated list of label names"},
					"q":         {Type: "string", Description: "Search query"},
					"milestone": {Type: "string", Description: "Milestone name or ID"},
					"page":      {Type: "integer", Description: "Page number"},
					"limit":     {Type: "integer", Description: "Page size"},
				},
			},
		},
		Handler: handleListIssues,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "get_issue",
			Description: "Get a single issue by its index number",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
					"index": {Type: "integer", Description: "Issue index number"},
				},
				Required: []string{"index"},
			},
		},
		Handler: handleGetIssue,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "create_issue",
			Description: "Create a new issue in a repository",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":     {Type: "string", Description: "Repository owner"},
					"repo":      {Type: "string", Description: "Repository name"},
					"title":     {Type: "string", Description: "Issue title"},
					"body":      {Type: "string", Description: "Issue body/description"},
					"assignees": {Type: "array", Description: "List of assignee usernames"},
					"labels":    {Type: "array", Description: "List of label IDs"},
					"milestone": {Type: "integer", Description: "Milestone ID"},
					"due_date":  {Type: "string", Description: "Due date (ISO 8601 format)"},
				},
				Required: []string{"title"},
			},
		},
		Handler: handleCreateIssue,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "edit_issue",
			Description: "Edit an existing issue",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":     {Type: "string", Description: "Repository owner"},
					"repo":      {Type: "string", Description: "Repository name"},
					"index":     {Type: "integer", Description: "Issue index number"},
					"title":     {Type: "string", Description: "New title"},
					"body":      {Type: "string", Description: "New body/description"},
					"state":     {Type: "string", Description: "New state", Enum: []string{"open", "closed"}},
					"assignees": {Type: "array", Description: "List of assignee usernames"},
					"milestone": {Type: "integer", Description: "Milestone ID (0 to clear)"},
					"due_date":  {Type: "string", Description: "Due date (ISO 8601 format)"},
				},
				Required: []string{"index"},
			},
		},
		Handler: handleEditIssue,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "list_issue_comments",
			Description: "List comments on an issue",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":  {Type: "string", Description: "Repository owner"},
					"repo":   {Type: "string", Description: "Repository name"},
					"index":  {Type: "integer", Description: "Issue index number"},
					"since":  {Type: "string", Description: "Only show comments updated after this date (ISO 8601 format)"},
					"before": {Type: "string", Description: "Only show comments updated before this date (ISO 8601 format)"},
				},
				Required: []string{"index"},
			},
		},
		Handler: handleListIssueComments,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "create_issue_comment",
			Description: "Add a comment to an issue",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
					"index": {Type: "integer", Description: "Issue index number"},
					"body":  {Type: "string", Description: "Comment body"},
				},
				Required: []string{"index", "body"},
			},
		},
		Handler: handleCreateIssueComment,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "edit_issue_comment",
			Description: "Edit an existing comment on an issue",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
					"id":    {Type: "integer", Description: "Comment ID"},
					"body":  {Type: "string", Description: "New comment body"},
				},
				Required: []string{"id", "body"},
			},
		},
		Handler: handleEditIssueComment,
	})

	r.Register(ToolDef{
		Tool: Tool{
			Name:        "delete_issue_comment",
			Description: "Delete a comment on an issue",
			InputSchema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Repository owner"},
					"repo":  {Type: "string", Description: "Repository name"},
					"id":    {Type: "integer", Description: "Comment ID"},
				},
				Required: []string{"id"},
			},
		},
		Handler: handleDeleteIssueComment,
	})
}

func handleListIssues(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	if v := stringParam(params, "state"); v != "" {
		query.Set("state", v)
	}
	if v := stringParam(params, "labels"); v != "" {
		query.Set("labels", v)
	}
	if v := stringParam(params, "q"); v != "" {
		query.Set("q", v)
	}
	if v := stringParam(params, "milestone"); v != "" {
		query.Set("milestones", v)
	}
	if v := intParam(params, "page"); v > 0 {
		query.Set("page", strconv.FormatInt(v, 10))
	}
	if v := intParam(params, "limit"); v > 0 {
		query.Set("limit", strconv.FormatInt(v, 10))
	}
	query.Set("type", "issues")

	return client.Get(fmt.Sprintf("/repos/%s/%s/issues", owner, repo), query)
}

func handleGetIssue(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	index := intParam(params, "index")
	if index == 0 {
		return nil, errors.New("index is required")
	}

	return client.Get(fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, index), nil)
}

func handleCreateIssue(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	title := stringParam(params, "title")
	if title == "" {
		return nil, errors.New("title is required")
	}

	body := map[string]any{"title": title}
	if v := stringParam(params, "body"); v != "" {
		body["body"] = v
	}
	if v := stringSliceParam(params, "assignees"); len(v) > 0 {
		body["assignees"] = v
	}
	if v := intSliceParam(params, "labels"); len(v) > 0 {
		body["labels"] = v
	}
	if v := intParam(params, "milestone"); v > 0 {
		body["milestone"] = v
	}
	if v := stringParam(params, "due_date"); v != "" {
		body["due_date"] = v
	}

	return client.Post(fmt.Sprintf("/repos/%s/%s/issues", owner, repo), body)
}

func handleEditIssue(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	index := intParam(params, "index")
	if index == 0 {
		return nil, errors.New("index is required")
	}

	body := map[string]any{}
	if v := stringParam(params, "title"); v != "" {
		body["title"] = v
	}
	if v := stringParam(params, "body"); v != "" {
		body["body"] = v
	}
	if v := stringParam(params, "state"); v != "" {
		body["state"] = v
	}
	if v := stringSliceParam(params, "assignees"); v != nil {
		body["assignees"] = v
	}
	if v, ok := params["milestone"]; ok {
		if n, ok := v.(float64); ok {
			body["milestone"] = int64(n)
		}
	}
	if v := stringParam(params, "due_date"); v != "" {
		body["due_date"] = v
	}

	return client.Patch(fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, index), body)
}

func handleListIssueComments(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	index := intParam(params, "index")
	if index == 0 {
		return nil, errors.New("index is required")
	}

	query := url.Values{}
	if v := stringParam(params, "since"); v != "" {
		query.Set("since", v)
	}
	if v := stringParam(params, "before"); v != "" {
		query.Set("before", v)
	}

	return client.Get(fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, repo, index), query)
}

func handleCreateIssueComment(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	index := intParam(params, "index")
	if index == 0 {
		return nil, errors.New("index is required")
	}

	body := stringParam(params, "body")
	if body == "" {
		return nil, errors.New("body is required")
	}

	return client.Post(fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, repo, index), map[string]any{"body": body})
}

func handleEditIssueComment(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	id := intParam(params, "id")
	if id == 0 {
		return nil, errors.New("id is required")
	}

	body := stringParam(params, "body")
	if body == "" {
		return nil, errors.New("body is required")
	}

	return client.Patch(fmt.Sprintf("/repos/%s/%s/issues/comments/%d", owner, repo, id), map[string]any{"body": body})
}

func handleDeleteIssueComment(client *Client, params map[string]any) (any, error) {
	owner, repo, err := resolveOwnerRepo(params)
	if err != nil {
		return nil, err
	}

	id := intParam(params, "id")
	if id == 0 {
		return nil, errors.New("id is required")
	}

	return nil, client.Delete(fmt.Sprintf("/repos/%s/%s/issues/comments/%d", owner, repo, id))
}
