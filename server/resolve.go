package main

import (
	"path/filepath"
	"strings"
)

type Node struct {
	path     string
	children map[string]*Node
}

func newNode(path string) *Node {
	return &Node{
		path:     path,
		children: make(map[string]*Node),
	}
}

func (n *Node) findPossibleRoots() []string {
	var rootsFromChildren []string
	for _, c := range n.children {
		if len(c.children) == 0 {
			return []string{n.path}
		}
		rootsFromChildren = append(rootsFromChildren, c.findPossibleRoots()...)
	}
	return rootsFromChildren
}

type Tree struct {
	root *Node
}

func NewTree() *Tree {
	return &Tree{
		root: newNode("/"),
	}
}

func (t *Tree) addPath(parent *Node, prefix string, paths []string) {
	if len(paths) == 0 {
		return
	}
	path := paths[0]
	fullpath := filepath.Join(prefix, path)
	if _, ok := parent.children[path]; !ok {
		parent.children[path] = newNode(fullpath)
	}

	t.addPath(parent.children[path], fullpath, paths[1:])
}

func (t *Tree) addPaths(paths []string) {
	for _, p := range paths {
		pathSplit := strings.Split(p, "/")[1:]
		t.addPath(t.root, "/", pathSplit)
	}
}

func GetRootPaths(paths []string) ([]string, bool) {
	t := NewTree()
	t.addPaths(paths)
	for _, c := range t.root.children {
		// If any of the root children have no other children then
		// files are being written at / and you cannot make the filesystem readonly
		if len(c.children) == 0 {
			return nil, false
		}
	}
	var roots []string
	for _, c := range t.root.children {
		roots = append(roots, c.findPossibleRoots()...)
	}
	return roots, true
}
