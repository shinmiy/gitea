// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_26

import (
	"xorm.io/xorm"
)

// AddStatusChangeToProjectColumn adds status_change column to project_board table
func AddStatusChangeToProjectColumn(x *xorm.Engine) error {
	type ProjectBoard struct {
		StatusChange int8 `xorm:"NOT NULL DEFAULT 0"`
	}

	return x.Sync(new(ProjectBoard))
}
