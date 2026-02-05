# modules/mcp — Agent Instructions

## Key constraints

- Use `code.gitea.io/gitea/modules/json` instead of `encoding/json` — the linter enforces this
- Avoid `json.RawMessage` / `[]byte` fields in structs — they get base64-encoded by the JSON encoder; use `any` instead
- All tool files follow the pattern `tools_<resource>.go` with a `register<Resource>Tools()` method called from `NewRegistry()` in `tools.go`
- `owner` and `repo` parameters are auto-injected by `Server.resolveArguments()` from `GITEA_OWNER`/`GITEA_REPO` defaults when not provided by the caller

## Adding tools

1. Add the handler function in the appropriate `tools_*.go` file
2. Register it in the corresponding `register*Tools()` method
3. For a new resource category, create `tools_<resource>.go`, add `register<Resource>Tools()`, and call it from `NewRegistry()` in `tools.go`
4. Use param helpers from `tools.go`: `stringParam`, `intParam`, `stringSliceParam`, `intSliceParam`, `resolveOwnerRepo`
5. The `Client` provides `Get`, `Post`, `Patch`, and `Delete` methods — `Delete` returns only `error` (no body)

## Testing changes

```
make fmt && make lint-go && make backend
```
