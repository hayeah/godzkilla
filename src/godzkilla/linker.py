"""Symlink management for skill directories."""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

from .namer import skill_name
from .source import Found


@dataclass
class LinkResult:
    name: str
    target: str = ""
    action: str = ""  # "create", "update", "skip", "remove"
    error: str = ""


class Linker:
    def __init__(self, dest_dir: Path, *, dry: bool = False):
        self.dest_dir = dest_dir
        self.dry = dry

    def install(self, base_name: str, skills: list[Found]) -> list[LinkResult]:
        """Additive-only: create symlinks for all found skills."""
        self.dest_dir.mkdir(parents=True, exist_ok=True)
        return [self._install_one(base_name, s) for s in skills]

    def _install_one(self, base_name: str, s: Found) -> LinkResult:
        name = skill_name(base_name, s.rel_path)
        link_path = self.dest_dir / name
        abs_target = str(s.skill_dir.resolve())

        # Check existing symlink.
        if link_path.is_symlink():
            existing = str(link_path.readlink())
            if existing == abs_target:
                return LinkResult(name=name, target=abs_target, action="skip")
            link_path.unlink()
            link_path.symlink_to(abs_target)
            return LinkResult(name=name, target=abs_target, action="update")

        if link_path.exists():
            return LinkResult(name=name, error=f"unexpected file at {link_path}")

        link_path.symlink_to(abs_target)
        return LinkResult(name=name, target=abs_target, action="create")

    def sync(self, desired: dict[str, str]) -> list[LinkResult]:
        """Declarative: make dest_dir match desired exactly."""
        if not self.dry:
            self.dest_dir.mkdir(parents=True, exist_ok=True)

        # Scan current symlinks.
        current: dict[str, str] = {}
        if self.dest_dir.exists():
            for entry in self.dest_dir.iterdir():
                if entry.is_symlink():
                    current[entry.name] = str(entry.readlink())

        results: list[LinkResult] = []

        # Creates and updates.
        for name, target in desired.items():
            link_path = self.dest_dir / name
            existing = current.get(name)

            if existing == target:
                results.append(LinkResult(name=name, target=target, action="skip"))
                continue

            verb = "update" if existing is not None else "create"

            if not self.dry:
                if existing is not None:
                    link_path.unlink()
                link_path.symlink_to(target)

            results.append(LinkResult(name=name, target=target, action=verb))

        # Removals.
        for name, target in current.items():
            if name in desired:
                continue
            if not self.dry:
                (self.dest_dir / name).unlink()
            results.append(LinkResult(name=name, target=target, action="remove"))

        return results


def print_results(results: list[LinkResult]) -> None:
    results.sort(key=lambda r: r.name)
    for r in results:
        if r.error:
            print(f"  error    {r.name}: {r.error}")
        else:
            print(f"  {r.action:<12s} {r.name} → {r.target}")
