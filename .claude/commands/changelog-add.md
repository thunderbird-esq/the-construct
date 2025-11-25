# Add Changelog Entry

Add an entry to CHANGELOG.md following Keep a Changelog format.

## Usage

```
/changelog-add [type] [message]
```

## Parameters

- **type**: One of: `added`, `changed`, `deprecated`, `removed`, `fixed`, `security`
- **message**: Brief description of the change

## Steps

1. If parameters not provided, ask the user:
   - What type of change? (added/changed/deprecated/removed/fixed/security)
   - What is the description of the change?

2. Read the current CHANGELOG.md file

3. Locate the `[Unreleased]` section

4. Add the new entry under the appropriate subsection:
   - Find or create the `### [Type]` heading
   - Add the entry as a bullet point: `- [message]`

5. Maintain alphabetical order of subsections if needed:
   - Added
   - Changed
   - Deprecated
   - Removed
   - Fixed
   - Security

6. Save the updated CHANGELOG.md

7. Confirm the addition to the user

## Example

Input:
```
/changelog-add fixed "Combat system race condition in player state updates"
```

Result in CHANGELOG.md:
```markdown
## [Unreleased]

### Fixed
- Combat system race condition in player state updates
```

## Notes

- Entries go under [Unreleased] until a version is released
- Keep entries concise but descriptive
- Reference PR numbers if applicable: `[#123]`
- Follow the existing format in the file
