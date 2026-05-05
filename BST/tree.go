package main

import ("fmt"
		"go-trees/BST/data"
		"go-trees/BST/traversal"
	"go-trees/BST/insert"
	"go-trees/BST/search"
	"go-trees/BST/count")


func main() {
	var root *data.Node // Start with an empty tree

	  // Insert some numbers
    values := []int{10, 5, 15, 3, 7, 12, 20}
	// values := []int{20, 15, 10, 7, 5, 3}
	// for some reason 20, 15, 10, 7, 5, 3 is not working as expected, it is creating a left skewed tree instead of a balanced tree. This is because the insert function is designed to maintain the binary search tree property, which means that all values less than the root go to the left and all values greater than the root go to the right. When you insert values in descending order, each new value is less than the previous one, so it always goes to the left, resulting in a left skewed tree.
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

	// Try some searches
    testValues := []int{7, 99, 12, 0}
    for _, v := range testValues {
        found := search.Search(root, v)
        fmt.Printf("Search for %d: %v\n", v, found)
    }

	fmt.Println("Total nodes:", count.CountNodes(root))  // Should print 7
}