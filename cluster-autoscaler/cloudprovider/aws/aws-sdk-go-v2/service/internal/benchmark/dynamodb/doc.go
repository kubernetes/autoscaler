// Package benchmark provides benchmarks for the v2 SDK's Amazon DynamoDB API
// client.
//
// Includes benchmarks for the client's customizations, and compares the v2 SDK
// against the v1 SDK's legacy performance.
//
// Example command to run the benchmark
//
//	go test -bench "/default" -run NONE -v -benchtime=10s -benchmem
package benchmark
