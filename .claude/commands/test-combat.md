# Test Combat System

Run comprehensive tests for the combat system with race detection and failure analysis.

## Steps

1. Run combat-specific unit tests (when they exist):
   ```bash
   go test -v -run TestCombat ./tests/unit/
   ```

2. Check for race conditions:
   ```bash
   go test -race -run TestCombat ./tests/unit/
   ```

3. Run integration tests that involve combat:
   ```bash
   go test -v ./tests/integration/
   ```

4. If any tests fail:
   - Analyze the failure output
   - Identify the root cause
   - Suggest fixes based on the error messages
   - Provide code examples if needed

5. Report summary:
   - Total tests run
   - Passed / Failed count
   - Any race conditions detected
   - Performance metrics if available

## Notes

- Currently, test infrastructure is in place but tests are marked as skipped
- This command will work fully once combat tests are implemented
- For now, it will report that tests are pending implementation
