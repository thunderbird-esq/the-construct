# Refactor Code

Perform code refactoring using specialized agents with quality validation.

## Steps

1. Ask the user what needs refactoring:
   - Specific file or package name
   - Type of refactoring needed:
     - Structure (extract functions, split files, create packages)
     - Performance (optimize algorithms, reduce allocations)
     - Style (improve readability, naming, Go idioms)
     - All of the above

2. Use the golang-pro agent to perform refactoring:
   - Analyze current code structure
   - Identify improvement opportunities
   - Apply Go best practices
   - Maintain backward compatibility
   - Preserve functionality

3. Use the code-reviewer agent to validate changes:
   - Verify code quality improvements
   - Check for introduced bugs
   - Ensure tests still pass
   - Review Go idioms usage

4. Show proposed changes:
   - Display diff of changes
   - Explain rationale for each change
   - Highlight key improvements

5. Run tests to verify nothing broke:
   ```bash
   make test
   make lint
   ```

6. If tests pass:
   - Summarize improvements made
   - Suggest next steps if applicable

7. If tests fail:
   - Analyze failures
   - Fix issues
   - Re-run tests

## Example Workflow

User wants to refactor `world.go`:
1. golang-pro analyzes the file
2. Suggests extracting combat logic into separate package
3. Proposes better error handling
4. code-reviewer validates changes
5. Tests confirm functionality preserved
6. Summary of improvements provided

## Notes

- Always run tests after refactoring
- Maintain git history (commit often)
- Focus on one aspect at a time
- Don't over-engineer
