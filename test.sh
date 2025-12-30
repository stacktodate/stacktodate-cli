#!/bin/bash

# Test script for stacktodate

set -e

echo "Running tests for stacktodate..."
echo ""

# Run all tests with verbose output
go test -v ./...

echo ""
echo "All tests passed!"
