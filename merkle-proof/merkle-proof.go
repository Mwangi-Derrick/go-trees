package main

import (
    "crypto/sha256"
    "fmt"
)

type MerkleNode struct {
    Hash   []byte
    Left   *MerkleNode
    Right  *MerkleNode
    Parent *MerkleNode
    Data   []byte
    Index  int // position in leaf list (0, 1, 2, ...)
}

type Proof struct {
    LeafHash   []byte
    LeafData   []byte
    LeafIndex  int
    SiblingHashes [][]byte // hashes of siblings from leaf to root
    Directions []bool      // true = sibling is on right, false = on left
}

func NewLeaf(data []byte, index int) *MerkleNode {
	//a leaf is a node that has a hash
    hash := sha256.Sum256(data)
    return &MerkleNode{
        Hash:  hash[:],
        Data:  data,
        Index: index,
    }
}

func NewInternal(left, right *MerkleNode) *MerkleNode {
    combined := append(left.Hash, right.Hash...)
	//hash combined left and right hash
    hash := sha256.Sum256(combined)
    
    node := &MerkleNode{
        Hash:  hash[:],
        Left:  left,
        Right: right,
    }
    
    left.Parent = node
    right.Parent = node
    
    return node
}

func BuildTree(blocks [][]byte) *MerkleNode {
    if len(blocks) == 0 {
        return nil
    }
    
    var nodes []*MerkleNode
    for i, block := range blocks {
        nodes = append(nodes, NewLeaf(block, i))
    }
    
    for len(nodes) > 1 {
        var nextLevel []*MerkleNode
        
        for i := 0; i < len(nodes); i += 2 {
            if i+1 < len(nodes) {
                nextLevel = append(nextLevel, NewInternal(nodes[i], nodes[i+1]))
            } else {
                // Odd node: promote (in real systems, you might duplicate)
                nextLevel = append(nextLevel, nodes[i])
            }
        }
        nodes = nextLevel
    }
    
    return nodes[0]
}

// GenerateProof creates a proof for leaf at given index
func GenerateProof(root *MerkleNode, targetIndex int) *Proof {
    // Find the target leaf
    var targetLeaf *MerkleNode
    var findLeaf func(*MerkleNode)
    findLeaf = func(node *MerkleNode) {
        if node == nil {
            return
        }
        if node.Left == nil && node.Right == nil && node.Index == targetIndex {
            targetLeaf = node
            return
        }
        findLeaf(node.Left)
        if targetLeaf == nil {
            findLeaf(node.Right)
        }
    }
    findLeaf(root)
    
    if targetLeaf == nil {
        return nil
    }
    
    proof := &Proof{
        LeafHash:  targetLeaf.Hash,
        LeafData:  targetLeaf.Data,
        LeafIndex: targetIndex,
    }
    
    // Walk up from leaf to root, collecting siblings
    current := targetLeaf
    for current.Parent != nil {
        parent := current.Parent
        
        if parent.Left == current {
            // Current is left child → sibling is right child
            proof.SiblingHashes = append(proof.SiblingHashes, parent.Right.Hash)
            proof.Directions = append(proof.Directions, false) // sibling on right
        } else {
            // Current is right child → sibling is left child
            proof.SiblingHashes = append(proof.SiblingHashes, parent.Left.Hash)
            proof.Directions = append(proof.Directions, true) // sibling on left
        }
        
        current = parent
    }
    
    return proof
}

// VerifyProof verifies a proof against a trusted root hash
func VerifyProof(proof *Proof, trustedRoot []byte) bool {
    // Start with the leaf hash
    currentHash := proof.LeafHash
    
    // Recompute up the tree using sibling hashes
    for i, siblingHash := range proof.SiblingHashes {
        var combined []byte
        
        if proof.Directions[i] {
            // Sibling is on left: [sibling + current]
            combined = append(siblingHash, currentHash...)
        } else {
            // Sibling is on right: [current + sibling]
            combined = append(currentHash, siblingHash...)
        }
        
        hash := sha256.Sum256(combined)
        currentHash = hash[:]
    }
    
    // Compare computed root with trusted root
    if len(currentHash) != len(trustedRoot) {
        return false
    }
    
    for i := 0; i < len(currentHash); i++ {
        if currentHash[i] != trustedRoot[i] {
            return false
        }
    }
    
    return true
}

func PrintTree(node *MerkleNode, level int) {
    if node == nil {
        return
    }
    
    for i := 0; i < level; i++ {
        fmt.Print("  ")
    }
    
    if node.Left == nil && node.Right == nil {
        fmt.Printf("Leaf[%d]: %x (data: %q)\n", node.Index, node.Hash[:4], string(node.Data))
    } else {
        fmt.Printf("Internal: %x\n", node.Hash[:4])
    }
    
    PrintTree(node.Left, level+1)
    PrintTree(node.Right, level+1)
}

func main() {
    // Create 8 data blocks (like 8 chunks of a backup)
    blocks := [][]byte{
        []byte("Chunk_0: Genesis block"),
        []byte("Chunk_1: User photo 1"),
        []byte("Chunk_2: Database dump"),
        []byte("Chunk_3: Config file"),
        []byte("Chunk_4: Log file"),
        []byte("Chunk_5: Cache data"),
        []byte("Chunk_6: Index"),
        []byte("Chunk_7: Manifest"),
    }
    
    root := BuildTree(blocks)
    
    fmt.Println("=== Merkle Tree ===")
    PrintTree(root, 0)
    fmt.Printf("\nRoot hash: %x\n\n", root.Hash)
    
    // Generate a proof for chunk #2
    fmt.Println("=== Generate Proof for Chunk #2 ===")
    proof := GenerateProof(root, 2)
    
    fmt.Printf("Leaf hash: %x\n", proof.LeafHash[:4])
    fmt.Printf("Leaf data: %q\n", string(proof.LeafData))
    fmt.Printf("Number of sibling hashes: %d\n", len(proof.SiblingHashes))
    fmt.Printf("Sibling hashes (showing first 4 bytes):\n")
    for i, sibling := range proof.SiblingHashes {
        dir := "right"
        if proof.Directions[i] {
            dir = "left"
        }
        fmt.Printf("  Level %d: sibling on %s: %x\n", i+1, dir, sibling[:4])
    }
    
    // Verify the proof
    fmt.Println("\n=== Verify Proof ===")
    isValid := VerifyProof(proof, root.Hash)
    fmt.Printf("Proof valid: %v\n", isValid)
    
    // Tamper with the proof and see verification fail
    fmt.Println("\n=== Tamper with Proof (change sibling hash) ===")
    tamperedProof := &Proof{
        LeafHash:       proof.LeafHash,
        LeafData:       proof.LeafData,
        LeafIndex:      proof.LeafIndex,
        SiblingHashes:  make([][]byte, len(proof.SiblingHashes)),
        Directions:     proof.Directions,
    }
    copy(tamperedProof.SiblingHashes, proof.SiblingHashes)
    if len(tamperedProof.SiblingHashes) > 0 {
        // Corrupt the first sibling hash
        tamperedProof.SiblingHashes[0] = []byte("corrupted!")
    }
    
    isValidTampered := VerifyProof(tamperedProof, root.Hash)
    fmt.Printf("Tampered proof valid: %v\n", isValidTampered)
    
    // Demonstrate proof size efficiency
    fmt.Println("\n=== Proof Size Efficiency ===")
    fmt.Printf("Number of chunks: %d\n", len(blocks))
    fmt.Printf("Proof size: %d hashes (each 32 bytes = %d bytes)\n", 
        len(proof.SiblingHashes), len(proof.SiblingHashes)*32)
    fmt.Printf("Total tree size: %d bytes (if you sent the whole thing)\n", 
        len(blocks)*len(blocks[0])+len(blocks)*32)
}