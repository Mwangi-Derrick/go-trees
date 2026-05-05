package search

import "go-trees/BST/data"

// search returns true if value exists in tree
func search(root *data.Node, value int) bool {
    if root == nil {
        return false  // hit a dead end
    }
    
    if value == root.Value {
        return true   // found it!
    }
    //this createa a recursive search function that traverses the binary search tree. It checks if the current node is nil (base case), if the current node's value matches the target value, or if it should continue searching in the left or right subtree based on the comparison of the target value with the current node's value.
    if value < root.Value {
        return search(root.Left, value)   // go left
    }
    //if the left doesnt checkout the serach treverses to the right and checks if the value is there, if not it will return false
    return search(root.Right, value)      // go right
}