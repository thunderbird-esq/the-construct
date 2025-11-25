#!/bin/bash
# Simple load testing script for Matrix MUD

NUM_PLAYERS=${1:-10}

echo "🔥 Running load test with $NUM_PLAYERS concurrent players..."

for i in $(seq 1 $NUM_PLAYERS); do
    (
        echo "player$i"
        echo "password$i"
        sleep 1
        echo "look"
        sleep 1
        echo "north"
        sleep 1
        echo "quit"
    ) | telnet localhost 2323 > /dev/null 2>&1 &
done

wait
echo "✅ Load test complete"
