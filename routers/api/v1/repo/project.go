// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"net/http"

	"code.gitea.io/gitea/models/db"
	issues_model "code.gitea.io/gitea/models/issues"
	project_model "code.gitea.io/gitea/models/project"
	"code.gitea.io/gitea/modules/optional"
	api "code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/modules/web"
	"code.gitea.io/gitea/routers/api/v1/utils"
	"code.gitea.io/gitea/services/context"
	"code.gitea.io/gitea/services/convert"
	project_service "code.gitea.io/gitea/services/projects"
)

// ListProjects list all projects for a repository
func ListProjects(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/projects project projectListProjects
	// ---
	// summary: Get all projects for a repository
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: state
	//   in: query
	//   description: state of the projects (open, closed, all). Defaults to "open"
	//   type: string
	// - name: page
	//   in: query
	//   description: page number of results to return (1-based)
	//   type: integer
	// - name: limit
	//   in: query
	//   description: page size of results
	//   type: integer
	// responses:
	//   "200":
	//     "$ref": "#/responses/ProjectList"
	//   "404":
	//     "$ref": "#/responses/notFound"

	listOptions := utils.GetListOptions(ctx)

	var isClosed optional.Option[bool]
	switch ctx.FormString("state") {
	case "closed":
		isClosed = optional.Some(true)
	case "all":
		isClosed = optional.None[bool]()
	default:
		isClosed = optional.Some(false)
	}

	projects, total, err := db.FindAndCount[project_model.Project](ctx, project_model.SearchOptions{
		ListOptions: listOptions,
		RepoID:      ctx.Repo.Repository.ID,
		IsClosed:    isClosed,
		Type:        project_model.TypeRepository,
	})
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	if err := project_service.LoadIssueNumbersForProjects(ctx, projects, ctx.Doer); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	ctx.SetTotalCountHeader(total)
	ctx.JSON(http.StatusOK, convert.ToAPIProjectList(ctx, projects))
}

// GetProject get a project by ID
func GetProject(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/projects/{id} project projectGetProject
	// ---
	// summary: Get a project
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/Project"
	//   "404":
	//     "$ref": "#/responses/notFound"

	project, err := project_model.GetProjectForRepoByID(ctx, ctx.Repo.Repository.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	if err := project_service.LoadIssueNumbersForProject(ctx, project, ctx.Doer); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	ctx.JSON(http.StatusOK, convert.ToAPIProject(ctx, project))
}

// CreateProject create a project for a repository
func CreateProject(ctx *context.APIContext) {
	// swagger:operation POST /repos/{owner}/{repo}/projects project projectCreateProject
	// ---
	// summary: Create a project
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/CreateProjectOption"
	// responses:
	//   "201":
	//     "$ref": "#/responses/Project"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"
	//   "412":
	//     "$ref": "#/responses/error"

	form := web.GetForm(ctx).(*api.CreateProjectOption)

	project := &project_model.Project{
		RepoID:       ctx.Repo.Repository.ID,
		Title:        form.Title,
		Description:  form.Description,
		CreatorID:    ctx.Doer.ID,
		TemplateType: project_model.TemplateType(form.TemplateType),
		CardType:     project_model.CardType(form.CardType),
		Type:         project_model.TypeRepository,
	}

	if err := project_model.NewProject(ctx, project); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	ctx.JSON(http.StatusCreated, convert.ToAPIProject(ctx, project))
}

// EditProject modify a project
func EditProject(ctx *context.APIContext) {
	// swagger:operation PATCH /repos/{owner}/{repo}/projects/{id} project projectEditProject
	// ---
	// summary: Update a project
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: integer
	//   format: int64
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/EditProjectOption"
	// responses:
	//   "200":
	//     "$ref": "#/responses/Project"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	form := web.GetForm(ctx).(*api.EditProjectOption)

	project, err := project_model.GetProjectForRepoByID(ctx, ctx.Repo.Repository.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	if form.Title != nil {
		project.Title = *form.Title
	}
	if form.Description != nil {
		project.Description = *form.Description
	}
	if form.CardType != nil {
		project.CardType = project_model.CardType(*form.CardType)
	}

	if err := project_model.UpdateProject(ctx, project); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	if form.State != nil {
		isClosed := *form.State == "closed"
		if project.IsClosed != isClosed {
			if err := project_model.ChangeProjectStatus(ctx, project, isClosed); err != nil {
				ctx.APIErrorInternal(err)
				return
			}
		}
	}

	if err := project_service.LoadIssueNumbersForProject(ctx, project, ctx.Doer); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	ctx.JSON(http.StatusOK, convert.ToAPIProject(ctx, project))
}

// DeleteProject delete a project
func DeleteProject(ctx *context.APIContext) {
	// swagger:operation DELETE /repos/{owner}/{repo}/projects/{id} project projectDeleteProject
	// ---
	// summary: Delete a project
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "204":
	//     "$ref": "#/responses/empty"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	project, err := project_model.GetProjectForRepoByID(ctx, ctx.Repo.Repository.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	if err := project_model.DeleteProjectByID(ctx, project.ID); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// ListProjectColumns list all columns for a project
func ListProjectColumns(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/projects/{id}/columns project projectListProjectColumns
	// ---
	// summary: Get all columns for a project
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/ProjectColumnList"
	//   "404":
	//     "$ref": "#/responses/notFound"

	project, err := project_model.GetProjectForRepoByID(ctx, ctx.Repo.Repository.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	columns, err := project.GetColumns(ctx)
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	ctx.JSON(http.StatusOK, convert.ToAPIProjectColumnList(columns))
}

// CreateProjectColumn create a column for a project
func CreateProjectColumn(ctx *context.APIContext) {
	// swagger:operation POST /repos/{owner}/{repo}/projects/{id}/columns project projectCreateProjectColumn
	// ---
	// summary: Create a column for a project
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: integer
	//   format: int64
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/CreateProjectColumnOption"
	// responses:
	//   "201":
	//     "$ref": "#/responses/ProjectColumn"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	form := web.GetForm(ctx).(*api.CreateProjectColumnOption)

	project, err := project_model.GetProjectForRepoByID(ctx, ctx.Repo.Repository.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	column := &project_model.Column{
		ProjectID: project.ID,
		Title:     form.Title,
		Color:     form.Color,
		CreatorID: ctx.Doer.ID,
	}

	if err := project_model.NewColumn(ctx, column); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	ctx.JSON(http.StatusCreated, convert.ToAPIProjectColumn(column))
}

// EditProjectColumn modify a column
func EditProjectColumn(ctx *context.APIContext) {
	// swagger:operation PATCH /repos/{owner}/{repo}/projects/{id}/columns/{columnId} project projectEditProjectColumn
	// ---
	// summary: Update a column
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: integer
	//   format: int64
	//   required: true
	// - name: columnId
	//   in: path
	//   description: id of the column
	//   type: integer
	//   format: int64
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/EditProjectColumnOption"
	// responses:
	//   "200":
	//     "$ref": "#/responses/ProjectColumn"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	form := web.GetForm(ctx).(*api.EditProjectColumnOption)

	project, err := project_model.GetProjectForRepoByID(ctx, ctx.Repo.Repository.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	column, err := project_model.GetColumnByIDAndProjectID(ctx, ctx.PathParamInt64("columnId"), project.ID)
	if err != nil {
		if project_model.IsErrProjectColumnNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	if form.Title != nil {
		column.Title = *form.Title
	}
	if form.Color != nil {
		column.Color = *form.Color
	}

	if err := project_model.UpdateColumn(ctx, column); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	ctx.JSON(http.StatusOK, convert.ToAPIProjectColumn(column))
}

// DeleteProjectColumn delete a column
func DeleteProjectColumn(ctx *context.APIContext) {
	// swagger:operation DELETE /repos/{owner}/{repo}/projects/{id}/columns/{columnId} project projectDeleteProjectColumn
	// ---
	// summary: Delete a column
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: integer
	//   format: int64
	//   required: true
	// - name: columnId
	//   in: path
	//   description: id of the column
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "204":
	//     "$ref": "#/responses/empty"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	project, err := project_model.GetProjectForRepoByID(ctx, ctx.Repo.Repository.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	column, err := project_model.GetColumnByIDAndProjectID(ctx, ctx.PathParamInt64("columnId"), project.ID)
	if err != nil {
		if project_model.IsErrProjectColumnNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	if column.Default {
		ctx.APIError(http.StatusForbidden, "cannot delete the default column")
		return
	}

	if err := project_model.DeleteColumnByID(ctx, column.ID); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// MoveProjectColumn move a column
func MoveProjectColumn(ctx *context.APIContext) {
	// swagger:operation POST /repos/{owner}/{repo}/projects/{id}/columns/{columnId}/move project projectMoveProjectColumn
	// ---
	// summary: Move a column
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: integer
	//   format: int64
	//   required: true
	// - name: columnId
	//   in: path
	//   description: id of the column
	//   type: integer
	//   format: int64
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/MoveProjectColumnOption"
	// responses:
	//   "200":
	//     "$ref": "#/responses/ProjectColumn"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	form := web.GetForm(ctx).(*api.MoveProjectColumnOption)

	project, err := project_model.GetProjectForRepoByID(ctx, ctx.Repo.Repository.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	column, err := project_model.GetColumnByIDAndProjectID(ctx, ctx.PathParamInt64("columnId"), project.ID)
	if err != nil {
		if project_model.IsErrProjectColumnNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	sortedColumnIDs := map[int64]int64{
		form.Sorting: column.ID,
	}

	if err := project_model.MoveColumnsOnProject(ctx, project, sortedColumnIDs); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	// Reload the column to get updated sorting
	column, err = project_model.GetColumnByIDAndProjectID(ctx, column.ID, project.ID)
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	ctx.JSON(http.StatusOK, convert.ToAPIProjectColumn(column))
}

// SetDefaultProjectColumn set a column as default
func SetDefaultProjectColumn(ctx *context.APIContext) {
	// swagger:operation POST /repos/{owner}/{repo}/projects/{id}/columns/{columnId}/default project projectSetDefaultProjectColumn
	// ---
	// summary: Set a column as the default
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: integer
	//   format: int64
	//   required: true
	// - name: columnId
	//   in: path
	//   description: id of the column
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/ProjectColumn"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	project, err := project_model.GetProjectForRepoByID(ctx, ctx.Repo.Repository.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	columnID := ctx.PathParamInt64("columnId")
	_, err = project_model.GetColumnByIDAndProjectID(ctx, columnID, project.ID)
	if err != nil {
		if project_model.IsErrProjectColumnNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	if err := project_model.SetDefaultColumn(ctx, project.ID, columnID); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	// Reload the column to get updated default status
	column, err := project_model.GetColumnByIDAndProjectID(ctx, columnID, project.ID)
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	ctx.JSON(http.StatusOK, convert.ToAPIProjectColumn(column))
}

// ListColumnItems list all items in a column
func ListColumnItems(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/projects/{id}/columns/{columnId}/items project projectListColumnItems
	// ---
	// summary: Get all items in a column
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: integer
	//   format: int64
	//   required: true
	// - name: columnId
	//   in: path
	//   description: id of the column
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/ProjectColumnItemList"
	//   "404":
	//     "$ref": "#/responses/notFound"

	project, err := project_model.GetProjectForRepoByID(ctx, ctx.Repo.Repository.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	column, err := project_model.GetColumnByIDAndProjectID(ctx, ctx.PathParamInt64("columnId"), project.ID)
	if err != nil {
		if project_model.IsErrProjectColumnNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	issues, err := issues_model.LoadIssuesFromColumn(ctx, column, &issues_model.IssuesOptions{
		RepoIDs: []int64{ctx.Repo.Repository.ID},
	})
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	projectIssues, err := column.GetIssues(ctx)
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	issueMap := make(map[int64]*api.Issue, len(issues))
	for _, issue := range issues {
		issueMap[issue.ID] = convert.ToAPIIssue(ctx, ctx.Doer, issue)
	}

	items := make([]*api.ProjectColumnItem, 0, len(projectIssues))
	for _, pi := range projectIssues {
		issue := issueMap[pi.IssueID]
		items = append(items, convert.ToAPIProjectColumnItem(pi, issue))
	}

	ctx.JSON(http.StatusOK, items)
}

// AddColumnItem add an issue to a column
func AddColumnItem(ctx *context.APIContext) {
	// swagger:operation POST /repos/{owner}/{repo}/projects/{id}/columns/{columnId}/items project projectAddColumnItem
	// ---
	// summary: Add an issue to a column
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: integer
	//   format: int64
	//   required: true
	// - name: columnId
	//   in: path
	//   description: id of the column
	//   type: integer
	//   format: int64
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/AddProjectColumnItemOption"
	// responses:
	//   "201":
	//     "$ref": "#/responses/ProjectColumnItem"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	form := web.GetForm(ctx).(*api.AddProjectColumnItemOption)

	project, err := project_model.GetProjectForRepoByID(ctx, ctx.Repo.Repository.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	columnID := ctx.PathParamInt64("columnId")
	_, err = project_model.GetColumnByIDAndProjectID(ctx, columnID, project.ID)
	if err != nil {
		if project_model.IsErrProjectColumnNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	issue, err := issues_model.GetIssueByID(ctx, form.IssueID)
	if err != nil {
		if issues_model.IsErrIssueNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	if issue.RepoID != ctx.Repo.Repository.ID {
		ctx.APIError(http.StatusForbidden, "issue does not belong to this repository")
		return
	}

	if err := issues_model.IssueAssignOrRemoveProject(ctx, issue, ctx.Doer, project.ID, columnID); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	// Reload the project issue to return
	var projectIssue project_model.ProjectIssue
	has, err := db.GetEngine(ctx).Where("project_id = ? AND issue_id = ?", project.ID, issue.ID).Get(&projectIssue)
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}
	if !has {
		ctx.APIErrorInternal(nil)
		return
	}

	apiIssue := convert.ToAPIIssue(ctx, ctx.Doer, issue)
	ctx.JSON(http.StatusCreated, convert.ToAPIProjectColumnItem(&projectIssue, apiIssue))
}

// DeleteProjectItem remove an issue from a project
func DeleteProjectItem(ctx *context.APIContext) {
	// swagger:operation DELETE /repos/{owner}/{repo}/projects/{id}/items/{itemId} project projectDeleteProjectItem
	// ---
	// summary: Remove an issue from a project
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: integer
	//   format: int64
	//   required: true
	// - name: itemId
	//   in: path
	//   description: id of the project item (not the issue id)
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "204":
	//     "$ref": "#/responses/empty"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	project, err := project_model.GetProjectForRepoByID(ctx, ctx.Repo.Repository.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	itemID := ctx.PathParamInt64("itemId")

	// Find the project issue
	var projectIssue project_model.ProjectIssue
	has, err := db.GetEngine(ctx).Where("id = ? AND project_id = ?", itemID, project.ID).Get(&projectIssue)
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	if !has {
		ctx.APIErrorNotFound()
		return
	}

	// Load the issue to pass to IssueAssignOrRemoveProject
	issue, err := issues_model.GetIssueByID(ctx, projectIssue.IssueID)
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	// Remove the issue from the project by setting project ID to 0
	if err := issues_model.IssueAssignOrRemoveProject(ctx, issue, ctx.Doer, 0, 0); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// MoveProjectItem move an issue to a different column
func MoveProjectItem(ctx *context.APIContext) {
	// swagger:operation POST /repos/{owner}/{repo}/projects/{id}/items/{itemId}/move project projectMoveProjectItem
	// ---
	// summary: Move an issue to a different column
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: id of the project
	//   type: integer
	//   format: int64
	//   required: true
	// - name: itemId
	//   in: path
	//   description: id of the project item (not the issue id)
	//   type: integer
	//   format: int64
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/MoveProjectItemOption"
	// responses:
	//   "200":
	//     "$ref": "#/responses/ProjectColumnItem"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	form := web.GetForm(ctx).(*api.MoveProjectItemOption)

	project, err := project_model.GetProjectForRepoByID(ctx, ctx.Repo.Repository.ID, ctx.PathParamInt64("id"))
	if err != nil {
		if project_model.IsErrProjectNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	itemID := ctx.PathParamInt64("itemId")

	// Find the project issue
	var projectIssue project_model.ProjectIssue
	has, err := db.GetEngine(ctx).Where("id = ? AND project_id = ?", itemID, project.ID).Get(&projectIssue)
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	if !has {
		ctx.APIErrorNotFound()
		return
	}

	// Verify the target column exists
	column, err := project_model.GetColumnByIDAndProjectID(ctx, form.ColumnID, project.ID)
	if err != nil {
		if project_model.IsErrProjectColumnNotExist(err) {
			ctx.APIErrorNotFound()
		} else {
			ctx.APIErrorInternal(err)
		}
		return
	}

	// Move the issue to the new column with the specified sorting
	sortedIssueIDs := map[int64]int64{
		form.Sorting: projectIssue.IssueID,
	}

	if err := project_service.MoveIssuesOnProjectColumn(ctx, ctx.Doer, column, sortedIssueIDs); err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	// Reload the project issue
	has, err = db.GetEngine(ctx).Where("id = ? AND project_id = ?", itemID, project.ID).Get(&projectIssue)
	if err != nil || !has {
		ctx.APIErrorInternal(err)
		return
	}

	issue, err := issues_model.GetIssueByID(ctx, projectIssue.IssueID)
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	apiIssue := convert.ToAPIIssue(ctx, ctx.Doer, issue)
	ctx.JSON(http.StatusOK, convert.ToAPIProjectColumnItem(&projectIssue, apiIssue))
}
