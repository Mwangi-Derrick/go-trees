package main

import (
    "bytes"
    "fmt"
)

// B-Tree of order 4 (max 4 children, min 2 except root)
// Each node holds up to 3 keys
const ORDER = 4
const MAX_KEYS = ORDER - 1  // 3
const MIN_KEYS = (ORDER / 2) - 1  // 1 (for non-root)

type BTreeNode struct {
    keys     [MAX_KEYS][]byte   // Sorted keys
    values   [MAX_KEYS][]byte   // Values (or disk offsets)
    children [ORDER]*BTreeNode  // Child pointers
    numKeys  int
    isLeaf   bool
}

type BTree struct {
    root *BTreeNode
}

func NewBTree() *BTree {
    return &BTree{
        root: &BTreeNode{isLeaf: true},
    }
}

// Search returns value for key
func (bt *BTree) Search(key []byte) ([]byte, bool) {
    return bt.root.search(key)
}

func (n *BTreeNode) search(key []byte) ([]byte, bool) {
    // Find the first key >= search key
    i := 0
    for i < n.numKeys && bytes.Compare(n.keys[i], key) < 0 {
        i++
    }
    
    // If found at this node
    if i < n.numKeys && bytes.Equal(n.keys[i], key) {
        return n.values[i], true
    }
    
    // If leaf, key doesn't exist
    if n.isLeaf {
        return nil, false
    }
    
    // Go to appropriate child
    return n.children[i].search(key)
}

// Insert adds a key-value pair
func (bt *BTree) Insert(key, value []byte) {
    root := bt.root
    if root.numKeys == MAX_KEYS {
        // Root is full → split and create new root
        newRoot := &BTreeNode{isLeaf: false}
        bt.root = newRoot
        newRoot.children[0] = root
        bt.splitChild(newRoot, 0)
        bt.insertNonFull(newRoot, key, value)
    } else {
        bt.insertNonFull(root, key, value)
    }
}

func (bt *BTree) insertNonFull(n *BTreeNode, key, value []byte) {
    i := n.numKeys - 1
    
    if n.isLeaf {
        // Find position and shift keys right
        for i >= 0 && bytes.Compare(n.keys[i], key) > 0 {
            n.keys[i+1] = n.keys[i]
            n.values[i+1] = n.values[i]
            i--
        }
        n.keys[i+1] = key
        n.values[i+1] = value
        n.numKeys++
    } else {
        // Find child to insert into
        for i >= 0 && bytes.Compare(n.keys[i], key) > 0 {
            i--
        }
        i++
        
        // If child is full, split it
        if n.children[i].numKeys == MAX_KEYS {
            bt.splitChild(n, i)
            // After split, decide which child to go to
            if bytes.Compare(n.keys[i], key) < 0 {
                i++
            }
        }
        bt.insertNonFull(n.children[i], key, value)
    }
}

func (bt *BTree) splitChild(parent *BTreeNode, childIdx int) {
    child := parent.children[childIdx]
    newChild := &BTreeNode{isLeaf: child.isLeaf}
    
    // Middle key moves up to parent
    mid := MAX_KEYS / 2  // 1 for ORDER=4
    
    // New child gets the right half of keys
    newChild.numKeys = MAX_KEYS - mid - 1
    for i := 0; i < newChild.numKeys; i++ {
        newChild.keys[i] = child.keys[mid+1+i]
        newChild.values[i] = child.values[mid+1+i]
    }
    
    // If not leaf, copy children too
    if !child.isLeaf {
        for i := 0; i <= newChild.numKeys; i++ {
            newChild.children[i] = child.children[mid+1+i]
        }
    }
    
    // Reduce child's key count
    child.numKeys = mid
    
    // Shift parent's keys right to make room
    for i := parent.numKeys; i > childIdx; i-- {
        parent.keys[i] = parent.keys[i-1]
        parent.values[i] = parent.values[i-1]
        parent.children[i+1] = parent.children[i]
    }
    
    // Insert middle key into parent
    parent.keys[childIdx] = child.keys[mid]
    parent.values[childIdx] = child.values[mid]
    parent.numKeys++
    parent.children[childIdx+1] = newChild
}

// Print the tree (for debugging)
func (bt *BTree) Print() {
    bt.printNode(bt.root, 0)
}

func (bt *BTree) printNode(n *BTreeNode, level int) {
    for i := 0; i < level; i++ {
        fmt.Print("  ")
    }
    fmt.Printf("Node (leaf=%v): ", n.isLeaf)
    for i := 0; i < n.numKeys; i++ {
        fmt.Printf("[%s=%s] ", n.keys[i], n.values[i])
    }
    fmt.Println()
    
    if !n.isLeaf {
        for i := 0; i <= n.numKeys; i++ {
            bt.printNode(n.children[i], level+1)
        }
    }
}

func main() {
    btree := NewBTree()
    
    // Insert some data
    testData := map[string]string{
        "cat":   "feline",
        "dog":   "canine",
        "bird":  "avian",
        "fish":  "aquatic",
        "lion":  "big cat",
        "mouse": "rodent",
        "snake": "reptile",
        "frog":  "amphibian",
        "bee":   "insect",
        "ant":   "insect",
    }
    
    for k, v := range testData {
        btree.Insert([]byte(k), []byte(v))
        fmt.Printf("Inserted: %s -> %s\n", k, v)
    }
    
    fmt.Println("\n=== B-Tree Structure ===")
    btree.Print()
    
    fmt.Println("\n=== Searches ===")
    searches := []string{"dog", "bird", "elephant", "cat"}
    for _, s := range searches {
        val, found := btree.Search([]byte(s))
        fmt.Printf("Search '%s': found=%v, value=%s\n", s, found, val)
    }
}