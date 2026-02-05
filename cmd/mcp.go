// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"errors"
	"os"

	"code.gitea.io/gitea/modules/mcp"

	"github.com/urfave/cli/v3"
)

// CmdMCP represents the MCP server subcommand.
var CmdMCP = &cli.Command{
	Name:  "mcp",
	Usage: "Start a Model Context Protocol server over stdio",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "url",
			Sources: cli.EnvVars("GITEA_URL"),
			Usage:   "Gitea instance base URL",
		},
		&cli.StringFlag{
			Name:    "token",
			Sources: cli.EnvVars("GITEA_TOKEN"),
			Usage:   "API access token",
		},
		&cli.StringFlag{
			Name:    "owner",
			Sources: cli.EnvVars("GITEA_OWNER"),
			Usage:   "Default repository owner",
		},
		&cli.StringFlag{
			Name:    "repo",
			Sources: cli.EnvVars("GITEA_REPO"),
			Usage:   "Default repository name",
		},
	},
	Action: runMCP,
}

func runMCP(_ context.Context, cmd *cli.Command) error {
	giteaURL := cmd.String("url")
	if giteaURL == "" {
		return errors.New("--url or GITEA_URL is required")
	}
	token := cmd.String("token")
	if token == "" {
		return errors.New("--token or GITEA_TOKEN is required")
	}

	client := mcp.NewClient(giteaURL, token)
	server := mcp.NewServer(client, cmd.String("owner"), cmd.String("repo"), os.Stdin, os.Stdout)
	return server.Run()
}
