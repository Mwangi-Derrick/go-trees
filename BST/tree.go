package main

import ("fmt"
		"go-trees/BST/data"
		"go-trees/BST/traversal")


func main() {
    // Build a tiny tree manually
    root := &data.Node{Value: 10}
    root.Left = &data.Node{Value: 5}
    root.Right = &data.Node{Value: 15}

	// Add a deeper data
    root.Left.Left = &data.Node{Value: 3}
    root.Left.Right = &data.Node{Value: 7}

    // Print the tree

    fmt.Println("Root:", root.Value)
    fmt.Println("Left child:", root.Left.Value)
    fmt.Println("Right child:", root.Right.Value)

	fmt.Println("My tree:")
    traversal.PrintTree(root, 0)
}