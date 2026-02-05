# modules/mcp

Model Context Protocol (MCP) server for Gitea. Exposes Gitea's REST API as MCP tools over a JSON-RPC 2.0 stdio transport, allowing LLM-based agents to interact with Gitea repositories.

## Usage

```
gitea mcp --url https://gitea.example.com --token <token> [--owner <owner>] [--repo <repo>]
```

All flags can also be set via environment variables: `GITEA_URL`, `GITEA_TOKEN`, `GITEA_OWNER`, `GITEA_REPO`.

### Claude Code

```
claude mcp add gitea -- ./gitea mcp --url https://gitea.example.com --token <token> --owner <owner> --repo <repo>
```

## Architecture

```
cmd/mcp.go          CLI entry point (gitea mcp subcommand)
modules/mcp/
  protocol.go       JSON-RPC 2.0 and MCP type definitions
  client.go         HTTP client for Gitea REST API (GET/POST/PATCH/DELETE)
  server.go         Stdio JSON-RPC server: reads lines, dispatches to handlers
  tools.go          Tool registry, dispatcher, and parameter helpers
  tools_issue.go    Issue and issue comment tools
  tools_label.go    Label tools
  tools_milestone.go Milestone tools
  tools_project.go  Project and project column tools
```

The server reads newline-delimited JSON-RPC messages from stdin and writes responses to stdout. No external MCP SDK is used.

## Tools (28)

### Issues
| Tool | Description |
|------|-------------|
| `list_issues` | List and search issues in a repository |
| `get_issue` | Get a single issue by index |
| `create_issue` | Create a new issue |
| `edit_issue` | Edit an existing issue |

### Issue Comments
| Tool | Description |
|------|-------------|
| `list_issue_comments` | List comments on an issue |
| `create_issue_comment` | Add a comment to an issue |
| `edit_issue_comment` | Edit an existing comment |
| `delete_issue_comment` | Delete a comment |

### Labels
| Tool | Description |
|------|-------------|
| `list_labels` | List labels in a repository |
| `get_label` | Get a single label |
| `create_label` | Create a new label |
| `edit_label` | Edit an existing label |
| `delete_label` | Delete a label |

### Milestones
| Tool | Description |
|------|-------------|
| `list_milestones` | List milestones in a repository |
| `get_milestone` | Get a single milestone |
| `create_milestone` | Create a new milestone |
| `edit_milestone` | Edit an existing milestone |
| `delete_milestone` | Delete a milestone |

### Projects
| Tool | Description |
|------|-------------|
| `list_projects` | List projects in a repository |
| `get_project` | Get a single project |
| `create_project` | Create a new project |
| `edit_project` | Edit an existing project |
| `delete_project` | Delete a project |
| `list_project_columns` | List columns in a project |
| `create_project_column` | Create a new project column |
| `edit_project_column` | Edit a project column |
| `delete_project_column` | Delete a project column |
| `move_project_column` | Move/reorder a project column |

## Adding a New Tool

1. Create a handler function: `func handleFoo(client *Client, params map[string]any) (any, error)`
2. Register it in the appropriate `register*Tools()` method with a `ToolDef` containing the tool name, description, and input schema
3. Use the param helpers (`stringParam`, `intParam`, `stringSliceParam`, `intSliceParam`) to extract typed parameters
4. Use `resolveOwnerRepo(params)` to get owner/repo (auto-populated from defaults when omitted)
