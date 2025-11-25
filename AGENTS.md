# Multi-Agent Development Architecture

Comprehensive guide to using specialized AI agents for Matrix MUD development, including patterns, coordination strategies, and resource management.

---

## Table of Contents

1. [Overview](#overview)
2. [Agent Types & Capabilities](#agent-types--capabilities)
3. [Memory Management](#memory-management)
4. [Coordination Patterns](#coordination-patterns)
5. [Workflow Examples](#workflow-examples)
6. [Best Practices](#best-practices)
7. [Troubleshooting](#troubleshooting)

---

## Overview

### What is Multi-Agent Development?

Multi-agent development uses specialized AI agents, each with domain expertise, to collaboratively work on different aspects of a project. This approach:

- **Increases velocity**: Parallel execution of independent tasks
- **Improves quality**: Domain experts for each task
- **Reduces context switching**: Focused agents with narrow scope
- **Enables complexity**: Coordinate complex workflows

### Key Principles

1. **Specialization**: Each agent has a specific role and expertise
2. **Coordination**: Agents communicate and build on each other's work
3. **Resource Awareness**: Memory and CPU budgeting
4. **Fault Tolerance**: Graceful degradation if agents fail

### Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│              Orchestration Layer                    │
│   (Task Decomposition, Scheduling, Coordination)    │
└──────────────────┬──────────────────────────────────┘
                   │
      ┌────────────┼────────────┐
      │            │            │
┌─────▼─────┐ ┌───▼──────┐ ┌──▼────────┐
│  Agent 1  │ │ Agent 2  │ │  Agent 3  │
│  (Code)   │ │ (Test)   │ │  (Docs)   │
└───────────┘ └──────────┘ └───────────┘
      │            │            │
      └────────────┼────────────┘
                   │
      ┌────────────▼────────────┐
      │   Shared Resources      │
      │ (Files, Git, Memory)    │
      └─────────────────────────┘
```

---

## Agent Types & Capabilities

### Code Development Agents

#### golang-pro
**Expertise**: Idiomatic Go code, performance optimization, Go patterns

**Best For**:
- Writing new Go code
- Refactoring existing code
- Performance optimization
- Implementing Go best practices
- Complex concurrency patterns

**Memory Usage**: ~2GB

**Example**:
```
Use golang-pro to refactor the world.go file into smaller, focused packages.
Focus on clean interfaces and proper error handling.
```

#### test-engineer
**Expertise**: Test design, coverage analysis, test frameworks

**Best For**:
- Unit test implementation
- Test coverage improvement
- Test strategy design
- Mocking and fixtures
- Test refactoring

**Memory Usage**: ~2GB

**Example**:
```
Use test-engineer to create comprehensive unit tests for pkg/world/
Target 90%+ coverage with focus on edge cases.
```

#### test-automator
**Expertise**: Test automation, CI/CD testing, E2E tests

**Best For**:
- Integration tests
- E2E test suites
- Test automation setup
- CI/CD test configuration
- Load testing scripts

**Memory Usage**: ~2.5GB

**Example**:
```
Use test-automator to set up integration tests for the combat system.
Include scenarios for PvE combat, fleeing, and death mechanics.
```

#### code-reviewer
**Expertise**: Code quality, maintainability, security

**Best For**:
- Pull request reviews
- Code quality checks
- Identifying anti-patterns
- Security vulnerability detection
- Maintainability assessment

**Memory Usage**: ~1.5GB

**Example**:
```
Use code-reviewer to review all changes in pkg/auth/ for security issues.
Focus on authentication logic and password handling.
```

### Architecture & Design Agents

#### backend-architect
**Expertise**: System design, API architecture, scalability

**Best For**:
- System architecture design
- API endpoint design
- Database schema design
- Scalability planning
- Microservices architecture

**Memory Usage**: ~2GB

**Example**:
```
Use backend-architect to design a REST API for player management.
Include authentication, rate limiting, and versioning strategy.
```

#### database-optimizer
**Expertise**: Database design, query optimization, indexing

**Best For**:
- Database schema design
- Query optimization
- Index strategy
- Migration planning
- Database performance tuning

**Memory Usage**: ~1.5GB

**Example**:
```
Use database-optimizer to design a PostgreSQL schema for the game.
Focus on efficient queries for player lookups and world state.
```

### Documentation Agents

#### technical-writer
**Expertise**: Technical documentation, tutorials, user guides

**Best For**:
- User documentation
- Architecture documentation
- Tutorial creation
- README files
- API documentation

**Memory Usage**: ~1.5GB

**Example**:
```
Use technical-writer to create ARCHITECTURE.md documenting the system design.
Include diagrams, component descriptions, and data flow.
```

#### documentation-expert
**Expertise**: Documentation structure, code documentation, standards

**Best For**:
- Godoc comments
- Code documentation
- Documentation standards
- Documentation structure
- API reference generation

**Memory Usage**: ~1.5GB

**Example**:
```
Use documentation-expert to add comprehensive godoc comments to all exported functions.
Follow Go documentation standards.
```

### Security & Performance Agents

#### security-engineer
**Expertise**: Security audits, vulnerability detection, secure coding

**Best For**:
- Security audits
- Vulnerability scanning
- Secure code review
- Penetration testing
- Security best practices

**Memory Usage**: ~2GB

**Example**:
```
Use security-engineer to audit the authentication system.
Check for common vulnerabilities: SQL injection, XSS, password storage.
```

#### performance-engineer
**Expertise**: Performance optimization, profiling, benchmarking

**Best For**:
- Performance profiling
- Bottleneck identification
- Optimization strategies
- Benchmark creation
- Load testing

**Memory Usage**: ~2.5GB

**Example**:
```
Use performance-engineer to profile the combat system.
Identify bottlenecks and suggest optimizations for 1000+ concurrent players.
```

### Operations Agents

#### devops-engineer
**Expertise**: CI/CD, deployment, infrastructure, monitoring

**Best For**:
- CI/CD pipeline setup
- Docker configuration
- Kubernetes deployment
- Monitoring setup
- Infrastructure as Code

**Memory Usage**: ~2GB

**Example**:
```
Use devops-engineer to create a production-ready Docker deployment.
Include health checks, logging, and monitoring.
```

---

## Memory Management

### Memory Budget Guidelines

**System Limit**: 8GB total memory usage

**Per-Agent Estimates**:
| Agent Type | Typical Memory | Max Memory |
|-----------|---------------|------------|
| golang-pro | 1.5-2GB | 3GB |
| test-engineer | 1.5-2GB | 2.5GB |
| documentation-expert | 1-1.5GB | 2GB |
| backend-architect | 1.5-2GB | 2.5GB |
| security-engineer | 2-2.5GB | 3GB |
| performance-engineer | 2-3GB | 3.5GB |

### Parallel Execution Matrix

**Safe Configurations**:

| # Agents | Agent Types | Total Memory | Status |
|----------|-------------|--------------|--------|
| 1 | Any | ~2GB | ✅ Safe |
| 2 | 2× Light (docs) | ~3GB | ✅ Safe |
| 2 | 1× Heavy + 1× Light | ~4.5GB | ✅ Safe |
| 3 | 3× Light | ~4.5GB | ✅ Safe |
| 3 | 2× Medium + 1× Light | ~6GB | ⚠️ Monitor |
| 4 | 4× Light | ~6GB | ⚠️ Monitor |
| 3 | 3× Heavy | ~9GB | ❌ Exceeds Limit |

**Monitoring Commands**:

```bash
# Check current memory usage (macOS)
top -l 1 | grep PhysMem

# Check current memory usage (Linux)
free -h

# Monitor in real-time
watch -n 1 free -h
```

### Memory Optimization Strategies

#### 1. Sequential Execution
When memory constrained, run agents sequentially:

```
Phase 1: Agent A (0-2GB)
Phase 2: Agent B (0-2GB)
Phase 3: Agent C (0-2GB)
```

#### 2. Batching
Group similar tasks for single agent:

```
Instead of:
- Agent 1: Document file A
- Agent 2: Document file B
- Agent 3: Document file C

Use:
- Agent 1: Document files A, B, and C
```

#### 3. Scoped Context
Limit context to reduce memory:

```
# High memory
"Review the entire codebase"

# Lower memory
"Review pkg/world/room.go focusing on error handling"
```

#### 4. Fallback Strategy
If memory > 6GB, automatically fall back to sequential:

```bash
CURRENT_MEM=$(check_memory)
if [ $CURRENT_MEM -gt 6144 ]; then
    echo "Memory high, using sequential execution"
    PARALLEL=false
fi
```

---

## Coordination Patterns

### Pattern 1: Pipeline

Agents work sequentially, each building on previous output.

```
┌────────┐    ┌────────┐    ┌────────┐    ┌────────┐
│ Agent  │───>│ Agent  │───>│ Agent  │───>│ Agent  │
│   1    │    │   2    │    │   3    │    │   4    │
└────────┘    └────────┘    └────────┘    └────────┘
  Design        Code         Test         Review
```

**When to Use**: Tasks have dependencies, sequential workflow

**Example**:
```
1. backend-architect: Design API structure
2. golang-pro: Implement API handlers
3. test-engineer: Write API tests
4. code-reviewer: Review implementation
```

**Memory**: ~2GB (only one agent active at a time)

### Pattern 2: Parallel Independent

Multiple agents work independently on different parts.

```
        ┌────────┐
    ┌──>│ Agent  │
    │   │   1    │
    │   └────────┘
    │
┌───┴──┐
│Start │   ┌────────┐
│      │──>│ Agent  │
└───┬──┘   │   2    │
    │      └────────┘
    │
    │   ┌────────┐
    └──>│ Agent  │
        │   3    │
        └────────┘
```

**When to Use**: Independent tasks, different files/packages

**Example**:
```
Parallel:
- Agent 1 (golang-pro): Implement pkg/world/
- Agent 2 (golang-pro): Implement pkg/auth/
- Agent 3 (test-engineer): Write tests for pkg/game/
```

**Memory**: ~6GB (3 agents × 2GB each)

### Pattern 3: MapReduce

Multiple agents process different chunks, then combine results.

```
        ┌────────┐
    ┌──>│ Agent  │───┐
    │   │  1 (M) │   │
    │   └────────┘   │
    │                │   ┌────────┐
┌───┴──┐             ├──>│ Agent  │
│Split │   ┌────────┐│   │ (R)    │
│      │──>│ Agent  ││   └────────┘
└───┬──┘   │  2 (M) ││     Combine
    │      └────────┘│
    │                │
    │   ┌────────┐   │
    └──>│ Agent  │───┘
        │  3 (M) │
        └────────┘
```

**When to Use**: Large task divisible into chunks, need aggregation

**Example**:
```
Map Phase (Parallel):
- Agent 1: Test coverage for pkg/world/
- Agent 2: Test coverage for pkg/game/
- Agent 3: Test coverage for pkg/auth/

Reduce Phase:
- Agent 4: Combine coverage reports, identify gaps
```

**Memory**: Map phase: ~6GB, Reduce phase: ~2GB

### Pattern 4: Review Chain

Multiple specialized reviewers examine the same code.

```
┌─────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│Code │───>│ Golang  │───>│Security │───>│ Perf    │
│     │    │ Review  │    │ Review  │    │ Review  │
└─────┘    └─────────┘    └─────────┘    └─────────┘
            Report 1       Report 2       Report 3
                              │
                              ▼
                         ┌─────────────┐
                         │ Consolidated│
                         │   Report    │
                         └─────────────┘
```

**When to Use**: Comprehensive review needed, multiple perspectives

**Example**:
```
Sequential Reviews:
1. golang-pro: Check Go best practices
2. security-engineer: Audit for vulnerabilities
3. performance-engineer: Identify bottlenecks
4. code-reviewer: Final quality check
```

**Memory**: ~2.5GB (sequential execution)

### Pattern 5: Recursive Refinement

Agent iterates on its own output until quality threshold met.

```
┌────────┐    ┌────────┐    ┌────────┐
│ Agent  │───>│ Agent  │───>│ Agent  │
│ Draft  │    │Review &│    │Finalize│
│        │    │Refine  │    │        │
└────────┘    └────────┘    └────────┘
    │             │
    └─────────────┘
    (Loop until quality met)
```

**When to Use**: Quality iterative improvement, self-correction

**Example**:
```
Iteration 1: golang-pro writes initial code
Iteration 2: golang-pro reviews and refines
Iteration 3: golang-pro final polish
```

**Memory**: ~2GB (same agent, fresh context each iteration)

---

## Workflow Examples

### Example 1: Feature Implementation

**Goal**: Implement user authentication with bcrypt

**Workflow**:

```yaml
Phase 1: Design (Sequential)
  - backend-architect:
      task: Design authentication system
      outputs: Architecture doc, API design
      memory: ~2GB

Phase 2: Implementation (Parallel - 3 agents, ~6GB)
  - golang-pro (Agent 1):
      task: Implement authentication handlers
      files: pkg/auth/*.go
      memory: ~2GB

  - test-engineer (Agent 2):
      task: Write authentication tests
      files: tests/unit/auth_test.go
      memory: ~2GB

  - documentation-expert (Agent 3):
      task: Document authentication API
      files: docs/API.md
      memory: ~2GB

Phase 3: Security Review (Sequential)
  - security-engineer:
      task: Security audit of auth system
      focus: Password handling, session management
      memory: ~2.5GB

Phase 4: Final Review (Sequential)
  - code-reviewer:
      task: Code quality review
      checks: Tests passing, docs complete
      memory: ~1.5GB
```

**Total Time**: ~45 minutes
**Peak Memory**: ~6GB (Phase 2)

### Example 2: Documentation Sprint

**Goal**: Create comprehensive project documentation

**Workflow**:

```yaml
Phase 1: Architecture Docs (Parallel - 2 agents, ~3GB)
  - technical-writer (Agent 1):
      task: Create ARCHITECTURE.md
      sections: System design, components, data flow
      memory: ~1.5GB

  - technical-writer (Agent 2):
      task: Create DEVELOPMENT.md
      sections: Setup, debugging, workflows
      memory: ~1.5GB

Phase 2: API Docs (Parallel - 2 agents, ~3.5GB)
  - documentation-expert (Agent 1):
      task: Create docs/API.md
      content: Telnet protocol, commands, responses
      memory: ~1.5GB

  - documentation-expert (Agent 2):
      task: Add godoc comments to all packages
      files: pkg/**/*.go
      memory: ~2GB

Phase 3: Review & Polish (Sequential)
  - technical-writer:
      task: Review all docs for consistency
      checks: Links, formatting, completeness
      memory: ~1.5GB
```

**Total Time**: ~30 minutes
**Peak Memory**: ~3.5GB (Phase 2)

### Example 3: Code Quality Improvement

**Goal**: Improve test coverage from 30% to 80%

**Workflow**:

```yaml
Phase 1: Analysis (Sequential)
  - test-engineer:
      task: Analyze current test coverage
      outputs: Coverage report, gap analysis
      memory: ~2GB

Phase 2: Test Implementation (Parallel - 4 agents, ~6GB)
  - test-engineer (Agent 1):
      task: Unit tests for pkg/world/
      target: 90% coverage
      memory: ~1.5GB

  - test-engineer (Agent 2):
      task: Unit tests for pkg/game/
      target: 90% coverage
      memory: ~1.5GB

  - test-engineer (Agent 3):
      task: Unit tests for pkg/auth/
      target: 90% coverage
      memory: ~1.5GB

  - test-automator (Agent 4):
      task: Integration tests
      scenarios: Combat, trading, world gen
      memory: ~2GB

NOTE: Run Agents 1-3 in parallel (~4.5GB), then Agent 4 (~2GB)

Phase 3: Validation (Sequential)
  - test-engineer:
      task: Verify coverage targets met
      run: All tests with coverage report
      memory: ~2GB
```

**Total Time**: ~60 minutes
**Peak Memory**: ~4.5GB (Phase 2, first batch)

### Example 4: Performance Optimization

**Goal**: Reduce combat system latency by 50%

**Workflow**:

```yaml
Phase 1: Profiling (Sequential)
  - performance-engineer:
      task: Profile current combat system
      tools: pprof, CPU profiling
      outputs: Bottleneck identification
      memory: ~3GB

Phase 2: Optimization (Sequential)
  - golang-pro:
      task: Implement optimizations
      focus: Hot paths identified in profiling
      memory: ~2GB

Phase 3: Benchmarking (Sequential)
  - performance-engineer:
      task: Create benchmarks
      compare: Before/after performance
      memory: ~2.5GB

Phase 4: Load Testing (Sequential)
  - test-automator:
      task: Load test with 1000+ users
      verify: Performance targets met
      memory: ~2.5GB
```

**Total Time**: ~90 minutes
**Peak Memory**: ~3GB (Phase 1)

---

## Best Practices

### 1. Agent Selection

**Match Agent to Task Complexity**:

```
Simple task (format code):
  ❌ golang-pro (overkill)
  ✅ Direct formatting tool

Medium task (add error handling):
  ✅ golang-pro (appropriate)

Complex task (refactor architecture):
  ✅ backend-architect + golang-pro (coordinated)
```

### 2. Task Decomposition

**Break Large Tasks into Agent-Sized Chunks**:

```
Bad:
  "Improve the entire codebase"

Good:
  1. "Add error handling to pkg/world/"
  2. "Add error handling to pkg/game/"
  3. "Add error handling to pkg/auth/"
```

### 3. Context Scoping

**Provide Focused Context**:

```
Inefficient:
  "Review all Go files"
  Context: Entire codebase
  Memory: High

Efficient:
  "Review pkg/auth/auth.go for security issues"
  Context: Single file, specific focus
  Memory: Low
```

### 4. Dependency Management

**Respect Task Dependencies**:

```
Parallel (OK):
  - Implement feature A
  - Implement feature B (independent)

Sequential (Required):
  - Design API structure
  - Implement API handlers (depends on design)
  - Write API tests (depends on implementation)
```

### 5. Quality Gates

**Verify Agent Output**:

```
After each agent:
  1. Run tests
  2. Check linter
  3. Verify builds
  4. Review changes

If quality gate fails:
  - Fix issues
  - Re-run agent if needed
  - Don't proceed to next phase
```

### 6. Resource Monitoring

**Monitor Throughout Workflow**:

```bash
# Before spawning agents
make check-memory

# During execution
watch -n 5 'make check-memory'

# If memory high:
- Kill non-essential processes
- Use sequential execution
- Reduce context scope
```

---

## Troubleshooting

### Issue 1: Memory Exceeded

**Symptoms**:
- System slowdown
- Swap usage increases
- Agents become unresponsive

**Solutions**:

```bash
# Check memory usage
top -l 1 | grep PhysMem

# Kill agents
pkill -f "claude.*agent"

# Switch to sequential execution
PARALLEL_MODE=false

# Reduce agent count
MAX_AGENTS=2
```

### Issue 2: Agent Conflicts

**Symptoms**:
- Merge conflicts
- Overwritten changes
- Inconsistent code

**Solutions**:

```bash
# Use file locking
flock /tmp/agent.lock agent-command

# Coordinate file access
Agent 1: pkg/world/*.go
Agent 2: pkg/auth/*.go  (different package)
Agent 3: tests/*.go     (different directory)

# Review changes before merging
git diff HEAD
```

### Issue 3: Quality Issues

**Symptoms**:
- Tests failing
- Linter warnings
- Build errors

**Solutions**:

```bash
# Run quality checks after each agent
make test
make lint
make build

# Use review agent
code-reviewer: Validate all changes

# Iterate if needed
If quality < threshold:
  Refine and re-run
```

### Issue 4: Task Stalls

**Symptoms**:
- Agent takes too long
- No progress updates
- Stuck in loop

**Solutions**:

```bash
# Set timeouts
AGENT_TIMEOUT=300  # 5 minutes

# Break into smaller tasks
Large task → 3-4 smaller tasks

# Provide more specific instructions
Vague: "Improve code"
Specific: "Add error handling to func X, Y, Z"
```

---

## Advanced Patterns

### Dynamic Agent Selection

Choose agents based on task characteristics:

```python
def select_agent(task):
    if task.type == "code" and task.language == "go":
        return "golang-pro"
    elif task.type == "test":
        return "test-engineer"
    elif task.type == "docs":
        return "technical-writer"
    elif task.complexity == "high":
        return ["backend-architect", "golang-pro"]  # Multiple
    else:
        return "general-purpose"
```

### Adaptive Parallelization

Adjust parallelization based on resources:

```python
def get_parallel_config():
    available_memory = check_memory()

    if available_memory < 4GB:
        return {"max_agents": 1, "mode": "sequential"}
    elif available_memory < 6GB:
        return {"max_agents": 2, "mode": "limited_parallel"}
    else:
        return {"max_agents": 4, "mode": "full_parallel"}
```

### Quality-Driven Iteration

Iterate until quality threshold met:

```python
def iterative_improvement(code, threshold=0.9):
    quality = 0
    iterations = 0
    max_iterations = 3

    while quality < threshold and iterations < max_iterations:
        improved_code = agent_improve(code)
        quality = agent_evaluate(improved_code)
        code = improved_code
        iterations += 1

    return code, quality
```

---

## Metrics & Analytics

### Track Agent Performance

**Metrics to Monitor**:
- Agent execution time
- Memory usage per agent
- Success rate
- Quality scores
- Cost (if applicable)

**Logging**:

```bash
# Log agent activity
echo "$(date),golang-pro,refactor,2.1GB,45s,success" >> .claude/agent-metrics.csv

# Analyze
cat .claude/agent-metrics.csv | \
  awk -F',' '{sum+=$5; count++} END {print "Avg time:", sum/count}'
```

---

## Resources

- [CLAUDE.md](CLAUDE.md) - Claude Code integration guide
- [CONTRIBUTING.md](CONTRIBUTING.md) - Development workflow
- [DEVLOG.md](DEVLOG.md) - Development journal

---

**Last Updated**: 2025-11-24
**Agent Framework Version**: v1.0
**Recommended Agents**: 1-4 concurrent
**Memory Budget**: <8GB
