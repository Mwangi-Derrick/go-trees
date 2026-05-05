package data



// Merkle node (hash pointer)
//root hash is sum of the children hashes
//if a child hash changes the parent hash also changes and so as the root hash
type MerkleNode struct {
    Hash  []byte //32 bytes from sha256       // hash of Left.Hash + Right.Hash + Data
    Left  *MerkleNode   // still an address, but...
    Right *MerkleNode
    // The trick: Left.Hash must equal the actual hash of Left's children
	Data []byte //only leaf nodes have data
}