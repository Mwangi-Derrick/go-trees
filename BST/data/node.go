package data


// Node is one box in the tree
//a node that replicates itself to the left and right, creating a binary tree structure. Each node contains an integer value and pointers to its left and right children, which can be nil if there are no children.
type Node struct {
    Value int
    Left  *Node  // pointer to left child (or nil)
    Right *Node  // pointer to right child (or nil)
}
