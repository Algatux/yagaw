package yagaw

import (
	"encoding/json"
	"fmt"
)

type Node struct {
	Path     string
	Value    string
	Subpaths []*Node
}

func NewNode(path string, value string) *Node {
	return &Node{Path: path, Value: value, Subpaths: []*Node{}}
}

type Tree struct {
	root *Node
}

func (r *Tree) Insert(path string, value string) {
	if r.root == nil {
		r.root = NewNode(path, value)
		return
	}

	node := r.root
	for {

		idx := longestCommonPathIndex(path, node.Path)

		if idx == len(node.Path) {
			path = path[idx:]
			for i, n := range node.Subpaths {
				jdx := longestCommonPathIndex(path, n.Path)
				if jdx > 0 {
					node = node.Subpaths[i]
				}
			}
			continue
		}

		if idx == 0 {
			node.Subpaths = append(node.Subpaths, NewNode(path, value))
			break
		}

		if idx > 0 {
			new := &Node{
				Path:     node.Path[idx:],
				Value:    node.Value,
				Subpaths: node.Subpaths,
			}
			node.Path = node.Path[:idx]
			node.Value = ""
			node.Subpaths = []*Node{new}
			path = path[idx:]
			continue
		}
	}

	s, _ := json.MarshalIndent(r.root, "", "  ")
	fmt.Println(string(s))

}

func NewTree() *Tree {
	return &Tree{}
}

func longestCommonPathIndex(a, b string) int {
	i := 0
	max := len(a)
	if l := len(b); l < max {
		max = l
	}
	for ; i < max; i++ {
		if a[i] != b[i] {
			break
		}
	}

	return i
}
