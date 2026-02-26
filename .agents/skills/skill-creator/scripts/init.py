#!/usr/bin/env python3
"""
Skill Initializer - Creates a new skill from template

Usage:
    uv run scripts/init.py <skill-name> [--ref] [--script] [--asset]

Examples:
    uv run scripts/init.py my-new-skill
    uv run scripts/init.py my-new-skill --script --ref
"""

import argparse
import re
import sys
from pathlib import Path

SKILLS_DIR = Path(__file__).parent.parent.parent

SKILL_TEMPLATE = """---
name: {skill_name}
description: [TODO: Complete and informative explanation of what the skill does and when to use it. Include WHEN to use this skill - specific scenarios, file types, or tasks that trigger it.]
---

# {skill_title}

## Overview

[TODO: 1-2 sentences explaining what this skill enables]

## Structuring This Skill

[TODO: Choose the structure that best fits this skill's purpose. Common patterns:

**1. Workflow-Based** (best for sequential processes)
- Works well when there are clear step-by-step procedures
- Example: DOCX skill with "Workflow Decision Tree" → "Reading" → "Creating" → "Editing"
- Structure: ## Overview → ## Workflow Decision Tree → ## Step 1 → ## Step 2...

**2. Task-Based** (best for tool collections)
- Works well when the skill offers different operations/capabilities
- Example: PDF skill with "Quick Start" → "Merge PDFs" → "Split PDFs" → "Extract Text"
- Structure: ## Overview → ## Quick Start → ## Task Category 1 → ## Task Category 2...

**3. Reference/Guidelines** (best for standards or specifications)
- Works well for brand guidelines, coding standards, or requirements
- Example: Brand styling with "Brand Guidelines" → "Colors" → "Typography" → "Features"
- Structure: ## Overview → ## Guidelines → ## Specifications → ## Usage...

**4. Capabilities-Based** (best for integrated systems)
- Works well when the skill provides multiple interrelated features
- Example: Product Management with "Core Capabilities" → numbered capability list
- Structure: ## Overview → ## Core Capabilities → ### 1. Feature → ### 2. Feature...

Patterns can be mixed and matched as needed. Most skills combine patterns (e.g., start with task-based, add workflow for complex operations).

Delete this entire "Structuring This Skill" section when done - it's just guidance.]

## [TODO: Replace with the first main section based on chosen structure]

[TODO: Add content here. See examples in existing skills:
- Code samples for technical skills
- Decision trees for complex workflows
- Concrete examples with realistic user requests
- References to scripts/templates/references as needed]

## Resources

If your skill needs bundled resources, create the appropriate directories:

### scripts/
Executable code (Python/Bash/etc.) that can be run directly to perform specific operations.

**Examples from other skills:**
- PDF skill: `fill_fillable_fields.py`, `extract_form_field_info.py` - utilities for PDF manipulation
- DOCX skill: `document.py`, `utilities.py` - Python modules for document processing

**Appropriate for:** Python scripts, shell scripts, or any executable code that performs automation, data processing, or specific operations.

**Note:** Scripts may be executed without loading into context, but can still be read by the agent for patching or environment adjustments.

### references/
Documentation and reference material intended to be loaded into context to inform the agent's process and thinking.

**Examples from other skills:**
- Product management: `communication.md`, `context_building.md` - detailed workflow guides
- BigQuery: API reference documentation and query examples
- Finance: Schema documentation, company policies

**Appropriate for:** In-depth documentation, API references, database schemas, comprehensive guides, or any detailed information that the agent should reference while working.

### assets/
Files not intended to be loaded into context, but rather used within the output the agent produces.

**Examples from other skills:**
- Brand styling: PowerPoint template files (.pptx), logo files
- Frontend builder: HTML/React boilerplate project directories
- Typography: Font files (.ttf, .woff2)

**Appropriate for:** Templates, boilerplate code, document templates, images, icons, fonts, or any files meant to be copied or used in the final output.

## File References

When referencing files in your skill, use markdown link syntax, not backticks:

**Correct:**
```markdown
See [the reference guide](references/REFERENCE.md) for details.

Run the extraction script:
[scripts/extract.py](scripts/extract.py)
```

**Incorrect:**
```markdown
See the reference guide in `references/REFERENCE.md`.

Run `scripts/extract.py`
```

This ensures proper linking and navigation within the skill.

---

Use flags to create directories: `uv run scripts/init.py <skill-name> --script --ref --asset`
"""


def validate_skill_name(name: str) -> bool:
    """Validate skill name format (hyphen-case, lowercase only)."""
    if not re.match(r"^[a-z][a-z0-9-]*[a-z0-9]$|^[a-z]$", name):
        return False
    if "--" in name or name.startswith("-") or name.endswith("-"):
        return False
    if len(name) > 64:
        return False
    return True


def init_skill(
    skill_name: str,
    with_ref: bool = False,
    with_script: bool = False,
    with_asset: bool = False,
) -> Path | None:
    """Initialize a new skill directory with template SKILL.md."""
    if not validate_skill_name(skill_name):
        print(f"Invalid skill name: {skill_name}")
        print("Must be hyphen-case (lowercase letters, digits, hyphens), max 64 chars")
        return None

    skill_dir = (SKILLS_DIR / skill_name).resolve()

    if skill_dir.exists():
        print(f"Skill directory already exists: {skill_dir}")
        return None

    try:
        skill_dir.mkdir(parents=True)
    except Exception as e:
        print(f"Error creating directory: {e}")
        return None

    skill_title = " ".join(word.capitalize() for word in skill_name.split("-"))
    skill_md = skill_dir / "SKILL.md"
    skill_md.write_text(
        SKILL_TEMPLATE.format(skill_name=skill_name, skill_title=skill_title)
    )
    print(f"Created: {skill_md}")

    if with_ref:
        ref_dir = skill_dir / "references"
        ref_dir.mkdir(exist_ok=True)
        print(f"Created: {ref_dir}/")

    if with_script:
        script_dir = skill_dir / "scripts"
        script_dir.mkdir(exist_ok=True)
        print(f"Created: {script_dir}/")

    if with_asset:
        asset_dir = skill_dir / "assets"
        asset_dir.mkdir(exist_ok=True)
        print(f"Created: {asset_dir}/")

    print(f"\nSkill '{skill_name}' created at: {skill_dir}")
    return skill_dir


def main():
    parser = argparse.ArgumentParser(description="Initialize a new skill")
    parser.add_argument("skill_name", help="Name of the skill (hyphen-case)")
    parser.add_argument(
        "--ref", action="store_true", help="Create references/ directory"
    )
    parser.add_argument(
        "--script", action="store_true", help="Create scripts/ directory"
    )
    parser.add_argument("--asset", action="store_true", help="Create assets/ directory")

    args = parser.parse_args()

    result = init_skill(
        args.skill_name,
        with_ref=args.ref,
        with_script=args.script,
        with_asset=args.asset,
    )

    sys.exit(0 if result else 1)


if __name__ == "__main__":
    main()
