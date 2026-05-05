package count 
import "go-trees/BST/data"

func CountNodes(root *data.Node) int {
    if root == nil {
        return 0
    }
	//the 1 returns one (current node) as it treverses the tree and finds a non null/nil node it returns 1 and then it adds the count of the left and right subtrees to it, effectively counting all nodes in the tree.
    return 1 + CountNodes(root.Left) + CountNodes(root.Right)
}