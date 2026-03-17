"""Skill naming: source path → flat filesystem-safe name."""


def skill_name(src: str, rel_path: str) -> str:
    """Derive a flat, filesystem-safe skill name.

    Both src and rel_path have "/" replaced with "_".

    Examples:
        ("github.com/hayeah/skills", "")    → "github.com_hayeah_skills"
        ("github.com/hayeah/skills", "foo") → "github.com_hayeah_skills_foo"
        ("github.com/hayeah/skills", "a/b") → "github.com_hayeah_skills_a_b"
    """
    base = src.replace("/", "_")
    if not rel_path:
        return base
    sub = rel_path.replace("/", "_")
    return f"{base}_{sub}"
