"""Source resolution, git cloning, and SKILL.md discovery."""

from __future__ import annotations

import os
import subprocess
from dataclasses import dataclass, field
from pathlib import Path

KNOWN_HOSTS = ("github.com/", "gitlab.com/", "bitbucket.org/", "codeberg.org/")


def is_remote(src: str) -> bool:
    return any(src.startswith(h) for h in KNOWN_HOSTS)


def normalize_source(src: str) -> str:
    """Convert browser URLs to bare host paths.

    "https://github.com/user/repo/tree/master/skills"
    → "github.com/user/repo/skills"
    """
    for scheme in ("https://", "http://"):
        if src.startswith(scheme):
            src = src.removeprefix(scheme)
            break

    repo_path, after = parse_remote(src)
    if not after:
        return repo_path

    parts = after.split("/", 2)  # ["tree", "branch", "path..."]
    if len(parts) >= 2 and parts[0] == "tree":
        if len(parts) == 2:
            return repo_path
        return f"{repo_path}/{parts[2]}"
    return src


def parse_remote(src: str) -> tuple[str, str]:
    """Split remote source into (repo_path, sub_path).

    "github.com/user/repo"         → ("github.com/user/repo", "")
    "github.com/user/repo/foo/bar" → ("github.com/user/repo", "foo/bar")
    """
    parts = src.split("/")
    if len(parts) <= 3:
        return src, ""
    return "/".join(parts[:3]), "/".join(parts[3:])


def _storage_dir(repo_path: str) -> Path:
    """Local directory for a remote source clone."""
    base = os.environ.get("GODZKILLA_PATH", "")
    if not base:
        base = str(Path.home() / ".godzkilla")
    elif base.startswith("~/"):
        base = str(Path.home() / base[2:])
    return Path(base) / repo_path


@dataclass
class Found:
    """A discovered SKILL.md location."""

    root_dir: Path
    skill_dir: Path
    rel_path: str  # "" for root-level skill


@dataclass
class Resolved:
    """Result of resolving a source identifier."""

    local_dir: Path
    remote: bool
    name: str
    repo_path: str = ""
    repo_dir: Path = field(default_factory=lambda: Path())
    sub_path: str = ""

    def ensure_cloned(self) -> None:
        """Clone remote source if not already present."""
        if not self.remote:
            return

        git_dir = self.repo_dir / ".git"
        if not git_dir.exists():
            self.repo_dir.parent.mkdir(parents=True, exist_ok=True)
            url = f"https://{self.repo_path}.git"
            print(f"cloning {url} → {self.repo_dir}")
            if self.sub_path:
                _clone_sparse(url, self.repo_dir, self.sub_path)
            else:
                _clone_full(url, self.repo_dir)
            return

        # Already cloned — ensure subpath is checked out if needed.
        if not self.sub_path:
            return
        if (self.repo_dir / self.sub_path).exists():
            return
        print(f"adding sparse checkout path: {self.sub_path}")
        _git_run(self.repo_dir, "sparse-checkout", "add", self.sub_path)

    def fetch(self) -> None:
        """git fetch + fast-forward merge."""
        print(f"fetching {self.repo_dir}")
        _git_run(self.repo_dir, "fetch", "--all")
        _git_run(self.repo_dir, "merge", "--ff-only", "@{u}")

    def find_skills(self) -> list[Found]:
        """Discover all SKILL.md files under local_dir."""
        return find_all(self.local_dir)


def resolve(src: str) -> Resolved:
    """Resolve a source string to a Resolved."""
    src = normalize_source(src)
    if is_remote(src):
        repo_path, sub_path = parse_remote(src)
        repo_dir = _storage_dir(repo_path)
        local_dir = repo_dir / sub_path if sub_path else repo_dir
        return Resolved(
            local_dir=local_dir,
            remote=True,
            name=src.replace("/", "_"),
            repo_path=repo_path,
            repo_dir=repo_dir,
            sub_path=sub_path,
        )
    abs_path = Path(src).resolve()
    return Resolved(local_dir=abs_path, remote=False, name=abs_path.name)


def find_all(root_dir: Path) -> list[Found]:
    """Walk root_dir and return every directory containing SKILL.md."""
    found: list[Found] = []
    for skill_md in root_dir.rglob("SKILL.md"):
        skill_dir = skill_md.parent
        rel = skill_dir.relative_to(root_dir)
        rel_str = "" if str(rel) == "." else str(rel)
        found.append(Found(root_dir=root_dir, skill_dir=skill_dir, rel_path=rel_str))
    return found


def _clone_full(url: str, dest: Path) -> None:
    subprocess.run(
        ["git", "clone", "--filter=tree:0", url, str(dest)],
        check=True,
    )


def _clone_sparse(url: str, dest: Path, sub_path: str) -> None:
    subprocess.run(
        ["git", "clone", "--filter=tree:0", "--sparse", url, str(dest)],
        check=True,
    )
    _git_run(dest, "sparse-checkout", "add", sub_path)


def _git_run(repo_dir: Path, *args: str) -> None:
    subprocess.run(["git", "-C", str(repo_dir), *args], check=True)
