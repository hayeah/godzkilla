"""Godzkilla CLI — skill management for AI agents."""

from __future__ import annotations

import json
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import Annotated, Optional

import typer

from hayeah.core.fzfmatch import parse_matcher

from .linker import Linker, print_results
from .namer import skill_name
from .source import Found, is_remote, resolve

app = typer.Typer(help="Skill management CLI — find, install, and update AI agent skills")


def _filter_skills(skills: list[Found], pattern: str) -> list[Found]:
    """Filter skills using fzfmatch pattern against their rel_path (or basename)."""
    matcher = parse_matcher(pattern)
    paths = [s.rel_path or s.skill_dir.name for s in skills]
    matched = set(matcher.match(paths))
    return [s for s, p in zip(skills, paths) if p in matched]


def _expand_path(p: str) -> Path:
    """Expand ~ in path strings."""
    return Path(p).expanduser()


# ---------------------------------------------------------------------------
# Sync directives
# ---------------------------------------------------------------------------


@dataclass
class Directive:
    """A parsed sync directive from JSON."""

    sources: list[str]
    destinations: list[Path]
    select: str | None = None
    name: str | None = None


def _parse_directives(raw: list[dict]) -> list[Directive]:
    directives: list[Directive] = []
    for i, entry in enumerate(raw):
        if "from" not in entry:
            raise typer.BadParameter(f"directive {i}: missing 'from'")
        if "to" not in entry:
            raise typer.BadParameter(f"directive {i}: missing 'to'")

        frm = entry["from"]
        sources = [frm] if isinstance(frm, str) else list(frm)

        to = entry["to"]
        destinations = [_expand_path(to)] if isinstance(to, str) else [_expand_path(d) for d in to]

        directives.append(Directive(
            sources=sources,
            destinations=destinations,
            select=entry.get("select"),
            name=entry.get("name"),
        ))
    return directives


def _collect_desired(directives: list[Directive]) -> dict[Path, dict[str, str]]:
    """Process directives into a per-destination desired map."""
    # dest_path → {skill_name → abs_target}
    desired: dict[Path, dict[str, str]] = {}

    for directive in directives:
        for src in directive.sources:
            resolved = resolve(src)
            resolved.ensure_cloned()

            if not resolved.local_dir.exists():
                typer.echo(f"source directory not found: {resolved.local_dir}", err=True)
                raise typer.Exit(1)

            skills = resolved.find_skills()
            if directive.select is not None:
                skills = _filter_skills(skills, directive.select)

            base_name = directive.name or resolved.name
            typer.echo(f"found {len(skills)} skill(s) in {resolved.local_dir} (base name: {base_name})")

            # Add to each destination's desired map
            for dest in directive.destinations:
                dest_map = desired.setdefault(dest, {})
                for s in skills:
                    sname = skill_name(base_name, s.rel_path)
                    dest_map[sname] = str(s.skill_dir.resolve())

    return desired


# ---------------------------------------------------------------------------
# Commands
# ---------------------------------------------------------------------------


@app.command()
def install(
    source: Annotated[str, typer.Option("--source", "-s", help="Skill source (GitHub path or local directory)")],
    destination: Annotated[Path, typer.Option("--destination", "-d", help="Destination directory for skill symlinks")],
    name: Annotated[Optional[str], typer.Option("--name", help="Override base name for skill naming")] = None,
    select: Annotated[Optional[str], typer.Option("--select", help="Filter skills with fzf pattern")] = None,
) -> None:
    """Install skills from a source into a destination directory."""
    resolved = resolve(source)
    base_name = name or resolved.name
    resolved.ensure_cloned()

    if not resolved.local_dir.exists():
        typer.echo(f"source directory not found: {resolved.local_dir}", err=True)
        raise typer.Exit(1)

    skills = resolved.find_skills()
    if select is not None:
        skills = _filter_skills(skills, select)

    if not skills:
        typer.echo("no SKILL.md files found")
        return

    typer.echo(f"found {len(skills)} skill(s) in {resolved.local_dir} (base name: {base_name})")
    linker = Linker(destination)
    results = linker.install(base_name, skills)
    print_results(results)

    errors = sum(1 for r in results if r.error)
    if errors:
        raise typer.Exit(1)


@app.command()
def sync(
    json_input: Annotated[Optional[str], typer.Option("--json", help="JSON directives (string or '-' for stdin)")] = None,
    dry: Annotated[bool, typer.Option("--dry", help="Dry run — show what would happen")] = False,
) -> None:
    """Sync skills into destination directories using JSON directives.

    Each directive: {"from": string|string[], "to": string|string[], "select"?: string, "name"?: string}
    """
    if json_input is None:
        typer.echo("provide --json '<directives>' or --json - for stdin", err=True)
        raise typer.Exit(1)

    if json_input == "-":
        raw_json = sys.stdin.read()
    else:
        raw_json = json_input

    raw = json.loads(raw_json)
    if not isinstance(raw, list):
        typer.echo("JSON must be an array of directives", err=True)
        raise typer.Exit(1)

    directives = _parse_directives(raw)
    desired = _collect_desired(directives)

    errors = 0
    for dest, dest_desired in sorted(desired.items(), key=lambda kv: str(kv[0])):
        if len(desired) > 1:
            typer.echo(f"\n→ {dest}")

        linker = Linker(dest, dry=dry)
        results = linker.sync(dest_desired)
        print_results(results)
        errors += sum(1 for r in results if r.error)

    if errors:
        raise typer.Exit(1)


@app.command("list")
def list_skills(
    destination: Annotated[Path, typer.Option("--destination", "-d", help="Destination directory to inspect")],
) -> None:
    """List installed skills in a destination directory."""
    if not destination.exists():
        typer.echo(f"destination {destination} does not exist")
        return

    count = 0
    for entry in sorted(destination.iterdir()):
        if not entry.is_symlink():
            continue
        target = entry.readlink()
        if not (Path(target) / "SKILL.md").exists():
            continue
        typer.echo(f"  {entry.name:<50s} → {target}")
        count += 1

    if count == 0:
        typer.echo("no skills installed")
    else:
        typer.echo(f"\n{count} skill(s) installed")


@app.command()
def update(
    source: Annotated[str, typer.Option("--source", "-s", help="Remote source to update")],
) -> None:
    """Update a previously cloned remote source."""
    if not is_remote(source):
        typer.echo("update only applies to remote sources; use git pull for local repos", err=True)
        raise typer.Exit(1)

    resolved = resolve(source)
    if not resolved.repo_dir.exists():
        typer.echo(f"source not yet cloned; run install first: {resolved.repo_dir}", err=True)
        raise typer.Exit(1)

    resolved.fetch()
