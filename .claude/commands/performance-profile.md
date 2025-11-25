# Performance Profile & Optimize

Profile the application and suggest optimizations using the performance-engineer agent.

## Steps

1. Add pprof HTTP endpoint if not present:
   ```go
   import _ "net/http/pprof"

   go func() {
       log.Println(http.ListenAndServe("localhost:6060", nil))
   }()
   ```

2. Build and run the application:
   ```bash
   go build -o bin/matrix-mud-prof .
   ./bin/matrix-mud-prof &
   ```

3. Generate load (simulate players):
   ```bash
   # Option 1: Manual connections
   for i in {1..10}; do
       (echo "player$i"; echo "pass"; sleep 1; echo "look"; sleep 1; echo "quit") | telnet localhost 2323 &
   done

   # Option 2: Use load testing tool (if available)
   ./scripts/load-test.sh
   ```

4. Collect CPU profile:
   ```bash
   go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
   ```

5. Analyze CPU hotspots:
   ```
   (pprof) top10
   (pprof) list functionName
   (pprof) web  # if graphviz installed
   ```

6. Collect memory profile:
   ```bash
   go tool pprof http://localhost:6060/debug/pprof/heap
   ```

7. Analyze memory usage:
   ```
   (pprof) top10
   (pprof) list functionName
   ```

8. Check goroutines:
   ```bash
   go tool pprof http://localhost:6060/debug/pprof/goroutine
   ```

9. Identify optimization opportunities:
   - **Hot paths**: Functions consuming >10% CPU
   - **Memory allocations**: Frequent allocations
   - **Lock contention**: Mutex wait times
   - **Goroutine leaks**: Growing goroutine count

10. Use performance-engineer agent to suggest optimizations for each issue

11. Implement optimizations:
    - Reduce allocations (preallocate slices, reuse buffers)
    - Optimize algorithms (better data structures)
    - Reduce lock scope (RWMutex for reads)
    - Pool frequently allocated objects

12. Create benchmarks for hot paths:
    ```go
    func BenchmarkCombatRound(b *testing.B) {
        // setup
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            // benchmark code
        }
    }
    ```

13. Run benchmarks before/after:
    ```bash
    go test -bench=. -benchmem > before.txt
    # apply optimizations
    go test -bench=. -benchmem > after.txt
    benchcmp before.txt after.txt
    ```

14. Generate performance report:
    - Current performance metrics
    - Identified bottlenecks
    - Proposed optimizations
    - Expected improvements
    - Benchmark comparisons

## Example Output

```
Performance Profile Report
=========================

CPU Hotspots:
1. world.ResolveCombatRound: 35% (optimization opportunity)
2. world.mutex.Lock: 20% (lock contention)
3. json.Marshal: 15% (frequent serialization)

Memory Allocations:
1. strings.Fields: 10k allocs/sec (can be reduced)
2. fmt.Sprintf: 8k allocs/sec (use strings.Builder)

Recommendations:
1. Optimize ResolveCombatRound: Use pre-allocated buffers
2. Reduce lock scope: RWMutex for reads
3. Cache JSON marshaling: Marshal once, reuse

Expected Improvement: 40% CPU reduction, 60% fewer allocations
```

## Notes

- Run profiling under realistic load
- Profile in production-like environment
- Focus on high-impact optimizations first
- Verify improvements with benchmarks
- Don't optimize prematurely
