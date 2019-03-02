package main

import (
	"log"
	"testing"
)

func TestResolve(t *testing.T) {
	filepaths := []string {
		"/a",
		"/a/b",
		"/a/b/c",
		"/a/b/c/d",
		"/e",
		"/e/f",
		"/e/f/g",
		"/e/f/h/i",
	}

	tree := NewTree()
	tree.addPaths(filepaths)

	log.Printf("Tree: %v", tree)

	rootPaths := tree.GetRootPaths()
	log.Printf("Paths: %+v", rootPaths)
}
