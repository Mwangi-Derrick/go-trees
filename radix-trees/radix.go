package main

import (
    "fmt"
    "strings"
)

// Radix tree node (compressed trie)
type RadixNode struct {
    prefix    string                 // The path segment stored here
    children  map[byte]*RadixNode    // Indexed by first byte of child prefix
    value     []byte                 // Value (if this is a terminal node)
    isTerminal bool                  // Marks end of a full path
}

type RadixTree struct {
    root *RadixNode
}

func NewRadixTree() *RadixTree {
    return &RadixTree{
        root: &RadixNode{
            children: make(map[byte]*RadixNode),
        },
    }
}

// Insert a key-value pair (key is a path like "/home/derrick/file.txt")
func (r *RadixTree) Insert(key string, value []byte) {
    r.root.insert(key, value)
}

func (n *RadixNode) insert(key string, value []byte) {
    if key == "" {
        n.isTerminal = true
        n.value = value
        return
    }
    
    firstByte := key[0]
    child, exists := n.children[firstByte]
    
    if !exists {
        // No child with this prefix → create leaf node
        newNode := &RadixNode{
            prefix:     key,
            children:   make(map[byte]*RadixNode),
            isTerminal: true,
            value:      value,
        }
        n.children[firstByte] = newNode
        return
    }
    
    // Find common prefix between key and child.prefix
    common := commonPrefix(key, child.prefix)
    
    if common == child.prefix {
        // Current child's prefix is fully contained in key
        // Recurse deeper
        child.insert(key[len(common):], value)
        return
    }
    
    // Split the child node
    // Example: child has "hello", key is "help"
    // common = "hel", child suffix = "lo", key suffix = "p"
    
    // Create a new intermediate node
    splitNode := &RadixNode{
        prefix:   common,
        children: make(map[byte]*RadixNode),
    }
    
    // Original child becomes a child of split node
    childSuffix := child.prefix[len(common):]
    child.prefix = childSuffix
    splitNode.children[childSuffix[0]] = child
    
    // Insert the new key as another child
    keySuffix := key[len(common):]
    newNode := &RadixNode{
        prefix:     keySuffix,
        children:   make(map[byte]*RadixNode),
        isTerminal: true,
        value:      value,
    }
    splitNode.children[keySuffix[0]] = newNode
    
    // Replace the original child with split node
    n.children[firstByte] = splitNode
}

func commonPrefix(a, b string) string {
    minLen := len(a)
    if len(b) < minLen {
        minLen = len(b)
    }
    
    for i := 0; i < minLen; i++ {
        if a[i] != b[i] {
            return a[:i]
        }
    }
    return a[:minLen]
}

// Search returns value for a key
func (r *RadixTree) Search(key string) ([]byte, bool) {
    return r.root.search(key)
}

func (n *RadixNode) search(key string) ([]byte, bool) {
    if key == "" && n.isTerminal {
        return n.value, true
    }
    
    if key == "" {
        return nil, false
    }
    
    firstByte := key[0]
    child, exists := n.children[firstByte]
    if !exists {
        return nil, false
    }
    
    // Check if key starts with child's prefix
    if strings.HasPrefix(key, child.prefix) {
        return child.search(key[len(child.prefix):])
    }
    
    return nil, false
}

// CollectAll returns all key-value pairs under a prefix
func (r *RadixTree) CollectAll(prefix string) map[string][]byte {
    result := make(map[string][]byte)
    
    // Navigate to node at prefix
    currentNode := r.root
    remaining := prefix
    
    for remaining != "" {
        firstByte := remaining[0]
        child, exists := currentNode.children[firstByte]
        if !exists {
            return result
        }
        
        if strings.HasPrefix(remaining, child.prefix) {
            remaining = remaining[len(child.prefix):]
            currentNode = child
        } else if strings.HasPrefix(child.prefix, remaining) {
            // We're at a node with prefix containing our search path
            remaining = ""
            currentNode = child
            break
        } else {
            return result
        }
    }
    
    // Collect all from this node down
    var collect func(*RadixNode, string)
    collect = func(node *RadixNode, path string) {
        fullPath := path + node.prefix
        if node.isTerminal {
            result[fullPath] = node.value
        }
        for _, child := range node.children {
            collect(child, fullPath)
        }
    }
    
    collect(currentNode, prefix[:len(prefix)-len(remaining)])
    return result
}

// Delete removes a key
func (r *RadixTree) Delete(key string) {
    r.root.delete(key)
}

func (n *RadixNode) delete(key string) bool {
    if key == "" {
        if n.isTerminal {
            n.isTerminal = false
            n.value = nil
            return len(n.children) == 0  // True if this node can be removed
        }
        return false
    }
    
    firstByte := key[0]
    child, exists := n.children[firstByte]
    if !exists {
        return false
    }
    
    if strings.HasPrefix(key, child.prefix) {
        if child.delete(key[len(child.prefix):]) {
            // Child has no more data → remove it
            delete(n.children, firstByte)
        }
    }
    
    return !n.isTerminal && len(n.children) == 0
}

func (r *RadixTree) Print() {
    r.printNode(r.root, 0, "")
}

func (r *RadixTree) printNode(n *RadixNode, level int, prefix string) {
    indent := strings.Repeat("  ", level)
    marker := ""
    if n.isTerminal {
        marker = fmt.Sprintf(" = %s", n.value)
    }
    fmt.Printf("%s%s%s%s\n", indent, prefix, n.prefix, marker)
    
    for _, child := range n.children {
        r.printNode(child, level+1, "")
    }
}

func main() {
    tree := NewRadixTree()
    
    // Insert file paths (like a file system)
    files := map[string][]byte{
        "/home/derrick/docs/readme.txt": []byte("Readme content"),
        "/home/derrick/docs/notes.txt":  []byte("Notes content"),
        "/home/derrick/photos/1.jpg":    []byte("Image data"),
        "/home/alice/thesis.pdf":        []byte("PDF content"),
        "/home/alice/notes.txt":         []byte("Alice's notes"),
        "/var/log/nginx/access.log":     []byte("Log line 1\nLog line 2"),
    }
    
    for path, content := range files {
        tree.Insert(path, content)
    }
    
    fmt.Println("=== Radix Tree Structure ===")
    tree.Print()
    
    fmt.Println("\n=== Search for specific file ===")
    val, found := tree.Search("/home/derrick/docs/readme.txt")
    fmt.Printf("/home/derrick/docs/readme.txt: found=%v, value='%s'\n", found, val)
    
    fmt.Println("\n=== Collect all under /home/derrick/ ===")
    all := tree.CollectAll("/home/derrick/")
    for path, content := range all {
        fmt.Printf("%s: '%s'\n", path, content)
    }
    
    fmt.Println("\n=== Delete /home/derrick/docs/notes.txt ===")
    tree.Delete("/home/derrick/docs/notes.txt")
    tree.Print()
}