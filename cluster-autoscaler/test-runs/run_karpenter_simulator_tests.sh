#!/bin/bash
# Script to run Karpenter Simulator tests manually.
set -e

echo "Running Karpenter Simulator unit tests..."
go test -v ./core/scaleup/orchestrator/ -run "TestKarpenterSimulator.*"

echo "Running Builder validation tests..."
go test -v ./builder/... -run "TestAutoscalerBuilder.*"

echo "All tests passed successfully!"
