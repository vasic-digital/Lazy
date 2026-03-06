# Course: Lazy Loading Patterns in Go

## Module Overview

This course covers the `digital.vasic.lazy` module, teaching lazy initialization patterns using Go generics and `sync.Once`. You will learn to defer expensive computations, build lazy service registries, and understand the trade-offs of deferred initialization.

## Prerequisites

- Basic Go knowledge (functions, error handling, goroutines)
- Understanding of Go generics syntax
- Go 1.21+ installed

## Lessons

| # | Title | Duration |
|---|-------|----------|
| 1 | Lazy Values with sync.Once | 30 min |
| 2 | Lazy Services and Practical Patterns | 30 min |

## Source Files

- `pkg/lazy/lazy.go` -- Value[T] and Service[T] implementations
