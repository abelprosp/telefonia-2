# Code navigation

This repository has a local Graft structural index in `graft/` and a Graft MCP server configured in Codex.

- When available, use targeted Graft MCP queries for symbol relationships, callers, and file signatures to keep context small.
- Use `rg` for literal text searches and to verify completeness; graph rankings are not exhaustive.
- Verify relevant source lines before editing. Source code and tests take precedence over graph summaries.
- Keep queries focused; request full definitions only when excerpts are insufficient.
- Keep Graft in structural mode. Do not enable `--deep`, external model calls, or cloud synchronization unless the user requests them.
- Treat `graft/` as a regenerable local cache; do not commit it.
- If Graft is unavailable, continue with normal repository tools.
