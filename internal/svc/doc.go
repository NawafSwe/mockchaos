// Package svc provides functions to run HTTP and gRPC mock servers from configuration files.
//
// The package loads mock handlers from JSON files in specified directories and starts
// servers that respond to requests according to the configured handlers, status codes,
// and latencies.
package svc
