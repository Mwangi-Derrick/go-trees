package main

import "fmt"



// PrintTree walks the tree and prints every node
func PrintTree(node *Node, level int) {
	// Base case: if node is nil, return
    if node == nil {
        return
    }
    
    // Print indentation to show depth
    for i := 0; i < level; i++ {
        fmt.Print("  ")
    }
    fmt.Println(node.Value)
    
    // Recursively print children (one level deeper)
    PrintTree(node.Left, level+1)
    PrintTree(node.Right, level+1)
}
