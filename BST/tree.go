package main

import ("fmt"
		"go-trees/BST/data"
		"go-trees/BST/traversal"
	"go-trees/BST/insert")


func main() {
	var root *data.Node // Start with an empty tree

	  // Insert some numbers
    values := []int{10, 5, 15, 3, 7, 12, 20}
    for _, v := range values {
        root = insert.Insert(root, v)
    }
    
    fmt.Println("Tree after inserts:")

    // Print the tree

    fmt.Println("Root:", root.Value)
    fmt.Println("Left child:", root.Left.Value)
    fmt.Println("Right child:", root.Right.Value)

	fmt.Println("My tree:")
    traversal.PrintTree(root, 0)
}