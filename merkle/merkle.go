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
    
    node := &data.MerkleNode{
        Hash:  hash[:],
        Left:  left,
        Right: right,
    }
    
    // Set parent pointers
    left.Parent = node
    right.Parent = node
    
    return node
}

// RecomputeHash updates this node's hash based on its children or data
func RecomputeHash(n *data.MerkleNode) {
    if IsLeaf(n) {
        hash := sha256.Sum256(n.Data)
        n.Hash = hash[:]
    } else {
        combined := append(n.Left.Hash, n.Right.Hash...)
        hash := sha256.Sum256(combined)
        n.Hash = hash[:]
    }
}

// UpdateLeaf changes a leaf's data and propagates the hash change up to the root
func UpdateLeaf(n *data.MerkleNode, newData []byte) {
    if !IsLeaf(n) {
        return
    }
    
    n.Data = newData
    
    // Propagate changes up
    current := n
    for current != nil {
        RecomputeHash(current)
        current = current.Parent
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
        for i := 0; i < len(nodes); i += 2 {
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
            if i+1 < len(nodes) {//not in the end
                nextLevel = append(nextLevel, NewInternal(nodes[i], nodes[i+1]))
            } else {//the end
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

// VerifyIntegrity checks if the tree is consistent
func VerifyIntegrity(node *data.MerkleNode) bool {
    if node == nil {
        return true
    }
    
    if IsLeaf(node) {
        return true
    }
    
    combined := append(node.Left.Hash, node.Right.Hash...)
    expectedHash := sha256.Sum256(combined)
    
    for i := 0; i < 32; i++ {
        if node.Hash[i] != expectedHash[i] {
            return false
        }
    }
    
    return VerifyIntegrity(node.Left) && VerifyIntegrity(node.Right)
}

func main() {
    blocks := [][]byte{
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

    fmt.Println("Before tampering:")
    fmt.Printf("Root hash: %x\n", root.Hash[:8])
    fmt.Printf("Integrity check: %v\n\n", VerifyIntegrity(root))

    // Find a leaf and corrupt it
    var leaf *data.MerkleNode
    var findLeaf func(*data.MerkleNode)
    findLeaf = func(n *data.MerkleNode) {
        if n == nil {
            return
        }
        if IsLeaf(n) {
            leaf = n
            return
        }
        findLeaf(n.Left)
        if leaf == nil {
            findLeaf(n.Right)
        }
    }
    findLeaf(root)
    
    if leaf != nil {
        fmt.Printf("Found leaf with data: %q\n", string(leaf.Data))
        
        // Use UpdateLeaf to see hash propagation
        fmt.Printf("Updating leaf to: %q...\n", "Hacked!")
        UpdateLeaf(leaf, []byte("Hacked!"))
        
        fmt.Printf("After update:\n")
        fmt.Printf("Root hash: %x\n", root.Hash[:8])
        fmt.Printf("Integrity check: %v\n", VerifyIntegrity(root))
    }
}