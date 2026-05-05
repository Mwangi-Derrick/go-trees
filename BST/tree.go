package main

import "fmt"

// Node is one box in the tree
//a node that replicates itself to the left and right, creating a binary tree structure. Each node contains an integer value and pointers to its left and right children, which can be nil if there are no children.
type Node struct {
    Value int
    Left  *Node  // pointer to left child (or nil)
    Right *Node  // pointer to right child (or nil)
}

func main() {
    // Build a tiny tree manually
    root := &Node{Value: 10}
    root.Left = &Node{Value: 5}
    root.Right = &Node{Value: 15}

	// Add a deeper node
    root.Left.Left = &Node{Value: 3}
    root.Left.Right = &Node{Value: 7}

    // Print the tree

    fmt.Println("Root:", root.Value)
    fmt.Println("Left child:", root.Left.Value)
    fmt.Println("Right child:", root.Right.Value)

	fmt.Println("My tree:")
    PrintTree(root, 0)
}