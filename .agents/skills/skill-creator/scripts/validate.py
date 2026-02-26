#!/usr/bin/env python3
"""
Skill Validator - Checks skill structure and frontmatter

Usage:
    uv run scripts/validate.py <skill-directory> [--allow-todos]
"""

import argparse
import re
import sys
import yaml
from pathlib import Path

ALLOWED_PROPS = {
    "name",
    "description",
    "license",
    "compatibility",
    "allowed-tools",
    "metadata",
}


def validate_skill(skill_path: Path, allow_todos: bool = False) -> tuple[bool, str]:
    """Validate a skill directory."""
    skill_md = skill_path / "SKILL.md"
    if not skill_md.exists():
        return False, "SKILL.md not found"

    content = skill_md.read_text()
    if not content.startswith("---"):
        return False, "No YAML frontmatter found"

    match = re.match(r"^---\n(.*?)\n---", content, re.DOTALL)
    if not match:
        return False, "Invalid frontmatter format"

    try:
        frontmatter = yaml.safe_load(match.group(1))
        if not isinstance(frontmatter, dict):
            return False, "Frontmatter must be a YAML dictionary"
    except yaml.YAMLError as e:
        return False, f"Invalid YAML: {e}"

    unexpected = set(frontmatter.keys()) - ALLOWED_PROPS
    if unexpected:
        return False, f"Unexpected keys: {', '.join(sorted(unexpected))}"

    name = frontmatter.get("name", "")
    if not isinstance(name, str) or not name:
        return False, "Missing or invalid 'name' in frontmatter"

    if skill_path.name != name:
        return (
            False,
            f"Directory name '{skill_path.name}' must match skill name '{name}'",
        )

    if not re.match(r"^[a-z][a-z0-9-]*[a-z0-9]$|^[a-z]$", name):
        return False, "Name must be hyphen-case (lowercase, digits, hyphens)"
    if "--" in name or name.startswith("-") or name.endswith("-"):
        return (
            False,
            "Name cannot contain consecutive hyphens or start/end with hyphen",
        )
    if len(name) > 64:
        return False, "Name exceeds 64 characters"

    desc = frontmatter.get("description", "")
    if not isinstance(desc, str):
        return False, f"Description must be a string, got {type(desc).__name__}"
    if not desc.strip():
        return False, "Description is empty"
    if len(desc) > 1024:
        return False, "Description exceeds 1024 characters"
    if "<" in desc or ">" in desc:
        return False, "Description cannot contain angle brackets"

    if "[TODO:" in desc:
        return False, "Description contains unresolved TODO placeholder"

    if not allow_todos:
        body = content.split("---", 2)[2] if len(content.split("---", 2)) > 2 else ""
        if "[TODO:" in body:
            return False, "SKILL.md body contains unresolved TODO placeholder"

    if len(list(skill_path.iterdir())) == 1:
        suffix = " (Allowed TODOs)" if allow_todos else ""
        return True, f"Valid (minimal skill){suffix}"

    for d in ["scripts", "references", "assets"]:
        dir_path = skill_path / d
        if dir_path.exists() and not dir_path.is_dir():
            return False, f"'{d}' exists but is not a directory"

    suffix = " (Allowed TODOs)" if allow_todos else ""
    return True, f"Valid{suffix}"


def main():
    parser = argparse.ArgumentParser(description="Validate a skill directory")
    parser.add_argument("skill_path", help="Path to the skill directory")
    parser.add_argument(
        "--allow-todos",
        action="store_true",
        help="Allow TODO placeholders in SKILL.md body (for skills like code-style that legitimately need TODO syntax)",
    )
    args = parser.parse_args()

    valid, msg = validate_skill(Path(args.skill_path), allow_todos=args.allow_todos)
    print(msg)
    sys.exit(0 if valid else 1)


if __name__ == "__main__":
    main()
