# Create Development Log Entry

Create a new entry in DEVLOG.md documenting development progress, decisions, or insights.

## Steps

1. Ask the user for entry details:
   - **Title**: Brief title for the entry (e.g., "Implemented Async Logging System")
   - **Key points**: What should be documented?
     - What was done?
     - Why was it done?
     - How was it implemented?
     - What were the results/learnings?

2. Read the current DEVLOG.md file

3. Create a new dated entry at the top of the log (after the header):
   ```markdown
   ## YYYY-MM-DD - [Title]

   ### What Was Accomplished

   [Description of what was done]

   ### Technical Decisions

   [Key decisions made and rationale]

   ### Challenges Encountered

   [Problems faced and solutions]

   ### Results & Metrics

   [Outcomes, performance numbers, etc.]

   ### Lessons Learned

   [Key takeaways]
   ```

4. Populate each section based on user input

5. Add relevant technical details:
   - Code examples (if applicable)
   - Performance metrics (if available)
   - Links to related commits or PRs
   - References to related issues

6. Save the updated DEVLOG.md

7. Optionally offer to commit the changes:
   ```bash
   git add DEVLOG.md
   git commit -m "docs: add devlog entry for [title]"
   ```

## Example Entry

```markdown
## 2025-11-24 - Implemented Structured Logging

### What Was Accomplished

Migrated from fmt.Printf to zerolog for structured logging throughout the application.

### Technical Decisions

**Decision**: Chose zerolog over zap and logrus
**Rationale**:
- Zero allocation logger
- 10x faster than alternatives
- JSON output built-in
- Simple API

### Implementation

\```go
log := zerolog.New(os.Stdout).With().Timestamp().Logger()
log.Info().Str("player", p.Name).Msg("Player connected")
\```

### Results & Metrics

- CPU overhead: <1%
- Memory allocations: 0 per log call
- Log parsing: Easy with jq
- Performance: 10μs per log entry

### Lessons Learned

1. Structured logging is essential for production
2. JSON logs enable better monitoring
3. Context-aware logging helps debugging
```

## Notes

- Entries should be technical but readable
- Include metrics and numbers when possible
- Document the "why" not just the "what"
- Link to related resources
- Keep it concise but complete
