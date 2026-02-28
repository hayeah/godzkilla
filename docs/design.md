godzkilla (`hayeah/godzkilla`) is a Go CLI tool for managing AI agent skills. The name is a pun on Godzilla.

- Find `SKILL.md` files recursively within a source directory
  - Each `SKILL.md` marks an installable skill
  - Nested paths are flattened: `path1/path2` → `path1_path2`
- Install skills by symlinking into destination directories
  - Supported destinations: `~/.claude/skills`, `~/.openclaw/skills`, `~/.codex/skills`, or any specified path
- Implemented in Go

Skill names are derived from the source path, converted to a single flat, filesystem-safe segment by replacing `/` with `_`.

- `github.com/hayeah/skills` → `github.com_hayeah_skills`
- `github.com/hayeah/skills/foo` → `github.com_hayeah_skills_foo`
- `github.com/hayeah/skills/bar` → `github.com_hayeah_skills_bar`

Remote sources are cloned via partial (treeless) clone.

- Default clone location: `~/.godzkilla/<name>/`
  - Override with env var `GODZKILLA_PATH` (e.g. `GODZKILLA_PATH=~` to use `~/github.com/...`)

Subcommands:

- `install --source github.com/hayeah/skills --destination ~/.claude/skills`
  - Partial-clone the source (idempotent)
  - Symlink each discovered skill into the destination
  - Additive only — never removes existing links
- `sync --destination ~/.claude/skills --source github.com/hayeah/skills [--source ...]`
  - Accepts multiple `--source` flags
  - Makes the destination an exact mirror of the combined sources
  - Adds missing links, updates changed ones, removes stale symlinks
  - `--dry` flag for dry run (reports actions without making changes)
  - Only removes symlinks, never regular files or directories
