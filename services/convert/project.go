// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package convert

import (
	"context"

	project_model "code.gitea.io/gitea/models/project"
	user_model "code.gitea.io/gitea/models/user"
	api "code.gitea.io/gitea/modules/structs"
)

// ToAPIProject converts a project_model.Project to api.Project
func ToAPIProject(ctx context.Context, project *project_model.Project) *api.Project {
	apiProject := &api.Project{
		ID:           project.ID,
		Title:        project.Title,
		Description:  project.Description,
		TemplateType: uint8(project.TemplateType),
		CardType:     uint8(project.CardType),
		IsClosed:     project.IsClosed,
		OpenIssues:   project.NumOpenIssues,
		ClosedIssues: project.NumClosedIssues,
		TotalIssues:  project.NumIssues,
		RepositoryID: project.RepoID,
		OwnerID:      project.OwnerID,
		Created:      project.CreatedUnix.AsTime(),
		Updated:      project.UpdatedUnix.AsTime(),
	}

	if project.IsClosed && project.ClosedDateUnix > 0 {
		closed := project.ClosedDateUnix.AsTime()
		apiProject.Closed = &closed
	}

	if project.CreatorID > 0 {
		creator, err := user_model.GetUserByID(ctx, project.CreatorID)
		if err == nil {
			apiProject.Creator = ToUser(ctx, creator, nil)
		}
	}

	return apiProject
}

// ToAPIProjectList converts a slice of project_model.Project to a slice of api.Project
func ToAPIProjectList(ctx context.Context, projects []*project_model.Project) []*api.Project {
	result := make([]*api.Project, len(projects))
	for i := range projects {
		result[i] = ToAPIProject(ctx, projects[i])
	}
	return result
}

// ToAPIProjectColumn converts a project_model.Column to api.ProjectColumn
func ToAPIProjectColumn(column *project_model.Column) *api.ProjectColumn {
	return &api.ProjectColumn{
		ID:        column.ID,
		Title:     column.Title,
		Color:     column.Color,
		ProjectID: column.ProjectID,
		Default:   column.Default,
		Created:   column.CreatedUnix.AsTime(),
		Updated:   column.UpdatedUnix.AsTime(),
	}
}

// ToAPIProjectColumnList converts a slice of project_model.Column to a slice of api.ProjectColumn
func ToAPIProjectColumnList(columns []*project_model.Column) []*api.ProjectColumn {
	result := make([]*api.ProjectColumn, len(columns))
	for i := range columns {
		result[i] = ToAPIProjectColumn(columns[i])
	}
	return result
}

// ToAPIProjectColumnItem converts a project_model.ProjectIssue to api.ProjectColumnItem
func ToAPIProjectColumnItem(item *project_model.ProjectIssue, issue *api.Issue) *api.ProjectColumnItem {
	return &api.ProjectColumnItem{
		ID:        item.ID,
		IssueID:   item.IssueID,
		ProjectID: item.ProjectID,
		ColumnID:  item.ProjectColumnID,
		Sorting:   item.Sorting,
		Issue:     issue,
	}
}
