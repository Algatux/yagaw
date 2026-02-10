package yagaw

import (
	"testing"
)

func TestRouter(t *testing.T) {
	router := NewTree()

	if router.root != nil {
		t.Errorf("Unexpected value in root")
	}
}

func TestRouterRootInsertion(t *testing.T) {
	router := NewTree()

	if router.root != nil {
		t.Errorf("Unexpected value in root")
	}

	router.Insert("/test", "a")

	if router.root == nil {
		t.Errorf("expected node not found")
	}

	if router.root.Path != "/test" {
		t.Errorf("Expected path /test in node not found")
	}
}

func TestRouterNodeInsertion(t *testing.T) {
	router := NewTree()
	router.Insert("/root", "a")
	router.Insert("/home", "b")
	router.Insert("/home/user", "c")
	router.Insert("/h", "d")

	if router.root.Path != "/" {
		t.Errorf("Expected path / in node not found")
	}

	if len(router.root.Subpaths) != 2 {
		t.Errorf("Expected subpath number was 2")
	}

	if router.root.Subpaths[0].Path != "root" {
		t.Errorf("Expected subpath 1 path was root")
	}

	if router.root.Subpaths[0].Value != "a" {
		t.Errorf("Expected subpath 1 value was a")
	}

	if router.root.Subpaths[1].Path != "home" {
		t.Errorf("Expected subpath 2 path was home")
	}

	if router.root.Subpaths[1].Value != "b" {
		t.Errorf("Expected subpath 2 value was b")
	}

	if router.root.Subpaths[1].Subpaths[0].Path != "/user" {
		t.Errorf("Expected subpath 1-0 path was /user")
	}

	// Expected final tree
	// {
	// 	"Path": "/",
	// 	"Value": "",
	// 	"Subpaths": [
	// 		{
	// 		"Path": "root",
	// 		"Value": "a",
	// 		"Subpaths": null
	// 		},
	// 		{
	// 		"Path": "home",
	// 		"Value": "b",
	// 		"Subpaths": []
	// 		}
	// 	]
	// }
}
