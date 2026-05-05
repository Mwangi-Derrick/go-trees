package search

import "go-trees/BST/data"

// Search returns true if value exists in tree
func Search(root *data.Node, value int) bool {
    if root == nil {
        return false  // hit a dead end
    }
    
    if value == root.Value {
        return true   // found it!
    }
    //this createa a recursive Search function that traverses the binary Search tree. It checks if the current node is nil (base case), if the current node's value matches the target value, or if it should continue searching in the left or right subtree based on the comparison of the target value with the current node's value.
    if value < root.Value {
        return Search(root.Left, value)   // go left
    }
    //if the left doesnt checkout the serach treverses to the right and checks if the value is there, if not it will return false
    return Search(root.Right, value)      // go right
}