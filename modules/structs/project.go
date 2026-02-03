// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package structs

import (
	"time"
)

// Project represents a project
type Project struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	TemplateType uint8  `json:"template_type"`
	CardType     uint8  `json:"card_type"`
	IsClosed     bool   `json:"closed"`
	OpenIssues   int64  `json:"open_issues"`
	ClosedIssues int64  `json:"closed_issues"`
	TotalIssues  int64  `json:"total_issues"`
	Creator      *User  `json:"creator"`
	RepositoryID int64  `json:"repo_id"`
	OwnerID      int64  `json:"owner_id"`
	// swagger:strfmt date-time
	Created time.Time `json:"created_at"`
	// swagger:strfmt date-time
	Updated time.Time `json:"updated_at"`
	// swagger:strfmt date-time
	Closed *time.Time `json:"closed_at,omitempty"`
}

// CreateProjectOption options for creating a project
type CreateProjectOption struct {
	// required: true
	Title string `json:"title" binding:"Required"`
	// Description of the project
	Description string `json:"description"`
	// Template type for the project (0: None, 1: Basic Kanban, 2: Bug Triage)
	TemplateType uint8 `json:"template_type"`
	// Card type for the project (0: Text Only, 1: Images and Text)
	CardType uint8 `json:"card_type"`
}

// EditProjectOption options for editing a project
type EditProjectOption struct {
	// Title of the project
	Title *string `json:"title"`
	// Description of the project
	Description *string `json:"description"`
	// Card type for the project (0: Text Only, 1: Images and Text)
	CardType *uint8 `json:"card_type"`
	// Whether the project is closed
	// enum: open,closed
	State *string `json:"state"`
}

// ProjectColumn represents a project column
type ProjectColumn struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Color     string `json:"color"`
	ProjectID int64  `json:"project_id"`
	Default   bool   `json:"default"`
	// swagger:strfmt date-time
	Created time.Time `json:"created_at"`
	// swagger:strfmt date-time
	Updated time.Time `json:"updated_at"`
}

// CreateProjectColumnOption options for creating a project column
type CreateProjectColumnOption struct {
	// required: true
	Title string `json:"title" binding:"Required"`
	// Color of the column (hex color code)
	Color string `json:"color"`
}

// EditProjectColumnOption options for editing a project column
type EditProjectColumnOption struct {
	// Title of the column
	Title *string `json:"title"`
	// Color of the column (hex color code)
	Color *string `json:"color"`
}

// MoveProjectColumnOption options for moving a project column
type MoveProjectColumnOption struct {
	// required: true
	// New sorting position (0-based index)
	Sorting int64 `json:"sorting"`
}

// ProjectColumnItem represents an item (issue) in a project column
type ProjectColumnItem struct {
	ID        int64  `json:"id"`
	IssueID   int64  `json:"issue_id"`
	ProjectID int64  `json:"project_id"`
	ColumnID  int64  `json:"column_id"`
	Sorting   int64  `json:"sorting"`
	Issue     *Issue `json:"issue,omitempty"`
}

// AddProjectColumnItemOption options for adding an item to a project column
type AddProjectColumnItemOption struct {
	// required: true
	// ID of the issue to add
	IssueID int64 `json:"issue_id" binding:"Required"`
}

// MoveProjectItemOption options for moving an item in a project
type MoveProjectItemOption struct {
	// required: true
	// Column ID to move the item to
	ColumnID int64 `json:"column_id" binding:"Required"`
	// New sorting position (0-based index)
	Sorting int64 `json:"sorting"`
}
