---
name: godzkilla
description: Skill management CLI — install and update AI agent skills from GitHub repos or local directories. Use when the user wants to install, list, or update skills for Claude, Codex, or OpenClaw.
---

# godzkilla

Install AI agent skills from GitHub repos or local directories by symlinking them into agent skill directories.

## Usage

```bash
# Install all skills from a GitHub source
godzkilla install --source github.com/hayeah/skills --destination ~/.claude/skills

# Install from a local directory
godzkilla install --source ./my-skills --destination ~/.claude/skills

# Override the base name used for skill naming
godzkilla install --source ./my-skills --name myskills --destination ~/.claude/skills

# List installed skills in a destination
godzkilla list --destination ~/.claude/skills

# Update a previously cloned remote source
godzkilla update --source github.com/hayeah/skills
```

## Notes

- Skill discovery: walks the source directory recursively; every directory containing a `SKILL.md` is an installable skill
- Skill naming: source path with `/` replaced by `_`, subpath appended
  - `github.com/hayeah/skills` → `github.com_hayeah_skills`
  - `github.com/hayeah/skills/foo` → `github.com_hayeah_skills_foo`
- Remote sources are cloned with `git clone --filter=tree:0` (treeless partial clone)
  - Default clone location: `~/.skilla/<name>/`
  - Override with `SKILLA_PATH` env var (e.g. `SKILLA_PATH=~`)
- Install is idempotent: existing symlinks pointing to the correct target are skipped
