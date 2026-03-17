---
name: godzkilla
description: Skill management CLI — install, sync, and update AI agent skills from GitHub repos or local directories. Use when the user wants to install, list, sync, or update skills for Claude, Codex, or OpenClaw.
---

# godzkilla

Equip agents with superpowers. **RAWR.**

![GODZKILLA](banner.jpg)

Install AI agent skills from GitHub repos or local directories by symlinking them into agent skill directories.

## Install

```bash
uv tool install -e .
```

## Usage

```bash
# Install all skills from a GitHub source
godzkilla install -s github.com/hayeah/skills -d ~/.claude/skills

# Install from a URL (tree/branch segments are stripped automatically)
godzkilla install -s https://github.com/anthropics/skills/tree/main/skills/frontend-design -d ~/.claude/skills

# Install with fzf pattern filter
godzkilla install -s github.com/hayeah/dotfiles/skills -d ~/.claude/skills --select "python;golang"

# List installed skills
godzkilla list -d ~/.claude/skills

# Update a previously cloned remote source
godzkilla update -s github.com/hayeah/skills
```

## Sync with JSON Directives

`godzkilla sync` takes a JSON array of directives via `--json`:

```bash
godzkilla sync --json '<directives>'
godzkilla sync --json -          # read from stdin
godzkilla sync --json '...' --dry  # dry run
```

Each directive is an object:

```jsonc
{
  "from": "string | string[]",   // source(s) — GitHub path, URL, or local dir
  "to": "string | string[]",     // destination dir(s) to sync into
  "select": "string",            // optional fzfmatch pattern to filter skills
  "name": "string"               // optional base name override
}
```

| Field    | Type                 | Required | Description                                  |
|----------|----------------------|----------|----------------------------------------------|
| `from`   | `string \| string[]` | yes      | Source(s) — GitHub paths, URLs, or local dirs |
| `to`     | `string \| string[]` | yes      | Destination dir(s), `~` is expanded           |
| `select` | `string`             | no       | fzfmatch pattern to filter discovered skills  |
| `name`   | `string`             | no       | Override base name for skill naming            |

Multiple directives targeting the same destination are merged (union of skills). Each destination is then synced declaratively: create missing, update changed, remove stale.

### Example

```bash
godzkilla sync --json '[
  {
    "from": [
      "github.com/hayeah/dotfiles/skills",
      "github.com/hayeah/devportv2",
      "github.com/hayeah/godzkilla",
      "github.com/hayeah/pymake"
    ],
    "to": [
      "~/.claude/skills",
      "~/.codex/skills",
      "~/.openclaw/skills"
    ]
  },
  {
    "from": "github.com/anthropics/skills",
    "to": "~/.claude/skills",
    "select": "frontend-design"
  }
]'
```

This syncs all skills from four repos into three agent dirs, plus cherry-picks `frontend-design` from anthropics/skills into claude only.

## Notes

- Skill discovery: walks the source directory recursively; every directory containing a `SKILL.md` is an installable skill
- Skill naming: source path with `/` replaced by `_`, subpath appended
  - `github.com/hayeah/skills` → `github.com_hayeah_skills`
  - `github.com/hayeah/skills/foo` → `github.com_hayeah_skills_foo`
- Remote sources are cloned with `git clone --filter=tree:0` (treeless partial clone)
  - Default clone location: `$GODZKILLA_PATH/<repo-path>/`
  - `GODZKILLA_PATH` defaults to `~/.godzkilla` if unset
- URL normalization: `https://github.com/user/repo/tree/branch/path` → `github.com/user/repo/path`
- `install` is additive and idempotent
- `sync` is declarative: only removes symlinks, never regular files or directories
- `--select` uses [fzfmatch](https://github.com/hayeah/dotfiles/blob/master/hayeah/src/hayeah/core/fzfmatch.py) extended-search syntax: `"a b"` (AND), `"a;b"` (OR), `"!a"` (NOT), `"^a"` / `"a$"` (anchors)
