// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package swagger

import (
	api "code.gitea.io/gitea/modules/structs"
)

// Project
// swagger:response Project
type swaggerResponseProject struct {
	// in:body
	Body api.Project `json:"body"`
}

// ProjectList
// swagger:response ProjectList
type swaggerResponseProjectList struct {
	// in:body
	Body []api.Project `json:"body"`
}

// ProjectColumn
// swagger:response ProjectColumn
type swaggerResponseProjectColumn struct {
	// in:body
	Body api.ProjectColumn `json:"body"`
}

// ProjectColumnList
// swagger:response ProjectColumnList
type swaggerResponseProjectColumnList struct {
	// in:body
	Body []api.ProjectColumn `json:"body"`
}

// ProjectColumnItem
// swagger:response ProjectColumnItem
type swaggerResponseProjectColumnItem struct {
	// in:body
	Body api.ProjectColumnItem `json:"body"`
}

// ProjectColumnItemList
// swagger:response ProjectColumnItemList
type swaggerResponseProjectColumnItemList struct {
	// in:body
	Body []api.ProjectColumnItem `json:"body"`
}
