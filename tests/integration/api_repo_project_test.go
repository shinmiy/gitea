// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"testing"

	auth_model "code.gitea.io/gitea/models/auth"
	project_model "code.gitea.io/gitea/models/project"
	repo_model "code.gitea.io/gitea/models/repo"
	"code.gitea.io/gitea/models/unittest"
	user_model "code.gitea.io/gitea/models/user"
	api "code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/tests"

	"github.com/stretchr/testify/assert"
)

func TestAPIRepoProject(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	// repo 1 has project 1 with columns 1, 2, 3
	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
	owner := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})
	project := unittest.AssertExistsAndLoadBean(t, &project_model.Project{ID: 1})

	session := loginUser(t, owner.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteIssue)

	t.Run("ListProjects", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/projects", owner.Name, repo.Name)).
			AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusOK)

		var projects []*api.Project
		DecodeJSON(t, resp, &projects)
		assert.NotEmpty(t, projects)
		// Should have at least the existing project
		found := false
		for _, p := range projects {
			if p.ID == project.ID {
				found = true
				assert.Equal(t, project.Title, p.Title)
			}
		}
		assert.True(t, found, "project should be in the list")
	})

	t.Run("GetProject", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d", owner.Name, repo.Name, project.ID)).
			AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusOK)

		var apiProject api.Project
		DecodeJSON(t, resp, &apiProject)
		assert.Equal(t, project.ID, apiProject.ID)
		assert.Equal(t, project.Title, apiProject.Title)
	})

	t.Run("CreateProject", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		req := NewRequestWithJSON(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/projects", owner.Name, repo.Name), api.CreateProjectOption{
			Title:       "Test API Project",
			Description: "Created via API",
		}).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusCreated)

		var apiProject api.Project
		DecodeJSON(t, resp, &apiProject)
		assert.Equal(t, "Test API Project", apiProject.Title)
		assert.Equal(t, "Created via API", apiProject.Description)
		assert.False(t, apiProject.IsClosed)
	})

	t.Run("EditProject", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Create a project to edit
		req := NewRequestWithJSON(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/projects", owner.Name, repo.Name), api.CreateProjectOption{
			Title: "Project to Edit",
		}).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusCreated)

		var created api.Project
		DecodeJSON(t, resp, &created)

		// Edit the project
		newTitle := "Edited Project"
		newDesc := "Updated description"
		req = NewRequestWithJSON(t, "PATCH", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d", owner.Name, repo.Name, created.ID), api.EditProjectOption{
			Title:       &newTitle,
			Description: &newDesc,
		}).AddTokenAuth(token)
		resp = MakeRequest(t, req, http.StatusOK)

		var edited api.Project
		DecodeJSON(t, resp, &edited)
		assert.Equal(t, newTitle, edited.Title)
		assert.Equal(t, newDesc, edited.Description)
	})

	t.Run("CloseAndReopenProject", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Create a project to close
		req := NewRequestWithJSON(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/projects", owner.Name, repo.Name), api.CreateProjectOption{
			Title: "Project to Close",
		}).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusCreated)

		var created api.Project
		DecodeJSON(t, resp, &created)

		// Close the project
		closedState := "closed"
		req = NewRequestWithJSON(t, "PATCH", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d", owner.Name, repo.Name, created.ID), api.EditProjectOption{
			State: &closedState,
		}).AddTokenAuth(token)
		resp = MakeRequest(t, req, http.StatusOK)

		var closed api.Project
		DecodeJSON(t, resp, &closed)
		assert.True(t, closed.IsClosed)

		// Reopen the project
		openState := "open"
		req = NewRequestWithJSON(t, "PATCH", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d", owner.Name, repo.Name, created.ID), api.EditProjectOption{
			State: &openState,
		}).AddTokenAuth(token)
		resp = MakeRequest(t, req, http.StatusOK)

		var reopened api.Project
		DecodeJSON(t, resp, &reopened)
		assert.False(t, reopened.IsClosed)
	})

	t.Run("DeleteProject", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Create a project to delete
		req := NewRequestWithJSON(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/projects", owner.Name, repo.Name), api.CreateProjectOption{
			Title: "Project to Delete",
		}).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusCreated)

		var created api.Project
		DecodeJSON(t, resp, &created)

		// Delete the project
		req = NewRequest(t, "DELETE", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d", owner.Name, repo.Name, created.ID)).
			AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNoContent)

		// Verify it's deleted
		req = NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d", owner.Name, repo.Name, created.ID)).
			AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNotFound)
	})
}

func TestAPIRepoProjectColumns(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
	owner := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})
	project := unittest.AssertExistsAndLoadBean(t, &project_model.Project{ID: 1})

	session := loginUser(t, owner.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteIssue)

	t.Run("ListColumns", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns", owner.Name, repo.Name, project.ID)).
			AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusOK)

		var columns []*api.ProjectColumn
		DecodeJSON(t, resp, &columns)
		assert.NotEmpty(t, columns)
		// Project 1 should have columns "To Do", "In Progress", "Done"
		assert.GreaterOrEqual(t, len(columns), 3)
	})

	t.Run("CreateColumn", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		req := NewRequestWithJSON(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns", owner.Name, repo.Name, project.ID), api.CreateProjectColumnOption{
			Title: "New Column",
			Color: "#ff0000",
		}).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusCreated)

		var column api.ProjectColumn
		DecodeJSON(t, resp, &column)
		assert.Equal(t, "New Column", column.Title)
		assert.Equal(t, "#ff0000", column.Color)
	})

	t.Run("EditColumn", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Create a column to edit
		req := NewRequestWithJSON(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns", owner.Name, repo.Name, project.ID), api.CreateProjectColumnOption{
			Title: "Column to Edit",
		}).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusCreated)

		var created api.ProjectColumn
		DecodeJSON(t, resp, &created)

		// Edit the column
		newTitle := "Edited Column"
		newColor := "#00ff00"
		req = NewRequestWithJSON(t, "PATCH", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns/%d", owner.Name, repo.Name, project.ID, created.ID), api.EditProjectColumnOption{
			Title: &newTitle,
			Color: &newColor,
		}).AddTokenAuth(token)
		resp = MakeRequest(t, req, http.StatusOK)

		var edited api.ProjectColumn
		DecodeJSON(t, resp, &edited)
		assert.Equal(t, newTitle, edited.Title)
		assert.Equal(t, newColor, edited.Color)
	})

	t.Run("DeleteColumn", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Create a column to delete (non-default)
		req := NewRequestWithJSON(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns", owner.Name, repo.Name, project.ID), api.CreateProjectColumnOption{
			Title: "Column to Delete",
		}).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusCreated)

		var created api.ProjectColumn
		DecodeJSON(t, resp, &created)

		// Delete the column
		req = NewRequest(t, "DELETE", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns/%d", owner.Name, repo.Name, project.ID, created.ID)).
			AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNoContent)
	})

	t.Run("DeleteDefaultColumnFails", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Get columns
		req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns", owner.Name, repo.Name, project.ID)).
			AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusOK)

		var columns []*api.ProjectColumn
		DecodeJSON(t, resp, &columns)

		// Find the default column
		var defaultColumn *api.ProjectColumn
		for _, col := range columns {
			if col.Default {
				defaultColumn = col
				break
			}
		}
		assert.NotNil(t, defaultColumn, "should have a default column")

		// Try to delete the default column (should fail)
		req = NewRequest(t, "DELETE", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns/%d", owner.Name, repo.Name, project.ID, defaultColumn.ID)).
			AddTokenAuth(token)
		MakeRequest(t, req, http.StatusForbidden)
	})

	t.Run("SetDefaultColumn", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Create a new column
		req := NewRequestWithJSON(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns", owner.Name, repo.Name, project.ID), api.CreateProjectColumnOption{
			Title: "New Default Column",
		}).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusCreated)

		var created api.ProjectColumn
		DecodeJSON(t, resp, &created)
		assert.False(t, created.Default)

		// Set it as default
		req = NewRequest(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns/%d/default", owner.Name, repo.Name, project.ID, created.ID)).
			AddTokenAuth(token)
		resp = MakeRequest(t, req, http.StatusOK)

		var updated api.ProjectColumn
		DecodeJSON(t, resp, &updated)
		assert.True(t, updated.Default)
	})
}

func TestAPIRepoProjectItems(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
	owner := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})
	project := unittest.AssertExistsAndLoadBean(t, &project_model.Project{ID: 1})

	session := loginUser(t, owner.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteIssue)

	t.Run("ListColumnItems", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Get columns first
		req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns", owner.Name, repo.Name, project.ID)).
			AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusOK)

		var columns []*api.ProjectColumn
		DecodeJSON(t, resp, &columns)
		assert.NotEmpty(t, columns)

		// List items in the first column
		req = NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns/%d/items", owner.Name, repo.Name, project.ID, columns[0].ID)).
			AddTokenAuth(token)
		resp = MakeRequest(t, req, http.StatusOK)

		var items []*api.ProjectColumnItem
		DecodeJSON(t, resp, &items)
		// Items may be empty, but the request should succeed
	})
}

func TestAPIProjectNotFound(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
	owner := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})

	session := loginUser(t, owner.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteIssue)

	t.Run("GetNonExistentProject", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/projects/99999", owner.Name, repo.Name)).
			AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNotFound)
	})

	t.Run("GetNonExistentColumn", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		project := unittest.AssertExistsAndLoadBean(t, &project_model.Project{ID: 1})

		req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns/99999/items", owner.Name, repo.Name, project.ID)).
			AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNotFound)
	})
}
