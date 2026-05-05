package data



// Merkle node (hash pointer)
//root hash is sum of the children hashes
//if a child hash changes the parent hash also changes and so as the root hash
type MerkleNode struct {
    Hash   []byte      // hash of Left.Hash + Right.Hash + Data
    Left   *MerkleNode
    Right  *MerkleNode
    Parent *MerkleNode
    Data   []byte      // only leaf nodes have data
}