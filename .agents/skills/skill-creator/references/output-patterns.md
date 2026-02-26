# Output Patterns

Use these patterns when skills need to produce consistent, high-quality output.

## Template Pattern

Provide templates for output format.

**For strict requirements (like API responses or data formats):**

```markdown
## Report structure

ALWAYS use this exact template structure:

# [Analysis Title]

## Executive summary
[One-paragraph overview]

## Key findings
- Finding 1
- Finding 2
```

**For flexible guidance:**

```markdown
## Report structure

Here is a sensible default format, but use your best judgment:

# [Analysis Title]

## Executive summary
[Overview]

## Key findings
[Adapt sections as needed]

Adjust based on the specific analysis type.
```

## Examples Pattern

For skills where output quality depends on seeing examples, provide input/output pairs:

```markdown
## Commit message format

**Example:**
Input: Added user authentication
Output:
```
feat(auth): implement JWT authentication

Add login endpoint and token validation
```

Follow: type(scope): brief description
```
