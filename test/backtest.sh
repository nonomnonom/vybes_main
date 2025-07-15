#!/bin/bash

# This script runs the main test script 1000 times to simulate load.

echo "--- Starting Backtest ---"
echo "Running test script 1000 times..."

for i in {1..100}
do
   echo "Running test iteration: $i"
   # Run in background to speed up the process
   bash test/tester.sh &> /dev/null &
done

# Wait for all background jobs to finish
wait

echo "--- Backtest Complete ---"