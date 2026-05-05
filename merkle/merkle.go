package main

import (
    "crypto/sha256"
    "fmt"
    "go-trees/merkle/data"
)

// IsLeaf returns true if this node has no children
func IsLeaf(n *data.MerkleNode) bool {
    return n.Left == nil && n.Right == nil
}

// NewLeaf creates a node from actual data
func NewLeaf(content []byte) *data.MerkleNode {
    hash := sha256.Sum256(content)
    return &data.MerkleNode{
        Hash: hash[:],
        Data: content,
    }
}

// NewInternal creates a parent node from two children
func NewInternal(left, right *data.MerkleNode) *data.MerkleNode {
    combined := append(left.Hash, right.Hash...)
    hash := sha256.Sum256(combined)
    
    return &data.MerkleNode{
        Hash:  hash[:],
        Left:  left,
        Right: right,
    }
}

// BuildTree builds a Merkle tree from a list of data blocks
func BuildTree(blocks [][]byte) *data.MerkleNode {
    if len(blocks) == 0 {
        return nil
    }
    
    var nodes []*data.MerkleNode
    for _, block := range blocks {
        nodes = append(nodes, NewLeaf(block))
    }
    
    for len(nodes) > 1 {
        var nextLevel []*data.MerkleNode
        //this loop performs a pair reduction
        for i := 0; i < len(nodes); i += 2 {
            if i+1 < len(nodes) {
				//if we arent at the end

				//Level 0: [A, B, C, D, E]  (5 nodes)

				// Pass 1:
				// - i=0: combine A+B → Parent1
				// - i=2: combine C+D → Parent2  
				// - i=4: E is odd one out → promote E

				// Result: [Parent1, Parent2, E]

				// Pass 2:
				// - i=0: combine Parent1+Parent2 → Grandparent
				// - i=2: E is odd → promote E

				// Result: [Grandparent, E]

				// Pass 3:
				// - i=0: combine Grandparent+E → Root

				// Result: [Root]
                nextLevel = append(nextLevel, NewInternal(nodes[i], nodes[i+1]))
            } else {
				//if we reach end
                nextLevel = append(nextLevel, nodes[i])
            }
        }
        nodes = nextLevel
    }
    
    return nodes[0]
}

// PrintTree shows the tree structure
func PrintTree(node *data.MerkleNode, level int) {
    if node == nil {
        return
    }
    
    for i := 0; i < level; i++ {
        fmt.Print("  ")
    }
    
    if IsLeaf(node) {
        fmt.Printf("Leaf: %x -> %q\n", node.Hash[:4], string(node.Data))
    } else {
        fmt.Printf("Internal: %x\n", node.Hash[:4])
    }
    
    PrintTree(node.Left, level+1)
    PrintTree(node.Right, level+1)
}

func main() {
	//an array of bytes
    blocks := [][]byte{
		//convert to bytes
        []byte("Hello"),
        []byte("World"),
        []byte("Merkle"),
        []byte("Tree"),
    }
    
    fmt.Println("Building Merkle tree from:", blocks)
    fmt.Println()
    
    root := BuildTree(blocks)
    
    fmt.Println("Tree structure:")
    PrintTree(root, 0)
    
    fmt.Println()
    fmt.Printf("Root hash (full): %x\n", root.Hash)
    fmt.Printf("Root hash (short): %x...\n", root.Hash[:8])
}