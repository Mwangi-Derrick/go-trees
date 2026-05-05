package insert

import "go-trees/BST/data"


// Insert adds a value to the correct position
func Insert(root *data.Node, value int) *data.Node {
    // Empty spot? Create a new node here
    if root == nil {
        return &data.Node{Value: value}
    }
    
    // Go left if value is smaller
    if value < root.Value {
		//the left is always less than the root hence the < checks, and the right is always greater than the root hence the > checks. This maintains the binary search tree property.
        root.Left = Insert(root.Left, value)
     // Go right if value is larger
	} else if value > root.Value {
        root.Right = Insert(root.Right, value)
    }
    // Equal? Do nothing (no duplicates)
    
    return root
}
