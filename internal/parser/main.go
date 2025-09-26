package main

import (
	"fmt"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_hledger "github.com/yeldirium/tree-sitter-hledger/bindings/go"
)

func main() {
	code := []byte(`account foo:ba
`)

	parser := tree_sitter.NewParser()
	defer parser.Close()
	_ = parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_hledger.Language()))

	tree := parser.Parse(code, nil)
	defer tree.Close()

	printTree(tree, code)

	fmt.Printf("Full expression:\n%s\n\n", tree.RootNode().ToSexp())

	printNodeAtCursor(tree, code, 0, 10)
}

func printNodeAtCursor(tree *tree_sitter.Tree, code []byte, row, column uint) {
	point := tree_sitter.Point{
		Row: row,
		Column: column,
	}
	node := tree.RootNode().DescendantForPointRange(point, point)

	fmt.Printf("node at cursor position %d:%d:\n", row, column)
	if node != nil {
		printNode(node, code)
	} else {
		fmt.Printf("no node found at cursor!\n")
	}
	fmt.Printf("\n")
}

func printTree(tree *tree_sitter.Tree, code []byte) {
	fmt.Printf("Each Node:\n")

	cursor := tree.Walk()
	defer cursor.Close()

outer:
	for {
		currentNode := cursor.Node()
		printNode(currentNode, code)

		if cursor.GotoFirstChild() {
			continue
		}
		if cursor.GotoNextSibling() {
			continue
		}
		for {
			if !cursor.GotoParent() {
				break outer
			}
			if cursor.GotoNextSibling() {
				break
			}
		}
		break
	}

	fmt.Printf("\n")
}

func printNode(node *tree_sitter.Node, code []byte) {
	nodeRange := node.Range()
	fmt.Printf(
		"%s: %s (%d:%d - %d:%d)\n",
		node.Kind(),
		code[nodeRange.StartByte:nodeRange.EndByte],
		nodeRange.StartPoint.Row,
		nodeRange.StartPoint.Column,
		nodeRange.EndPoint.Row,
		nodeRange.EndPoint.Column,
	)
}
