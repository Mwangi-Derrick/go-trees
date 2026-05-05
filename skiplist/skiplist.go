package main

import (
    "fmt"
    "math/rand"
    "sync"
    "time"
)

const MAX_LEVEL = 16      // Enough for 2^16 = 65536 elements
const PROBABILITY = 0.5   // 50% chance for each level

type SkipListNode struct {
    key     []byte
    value   []byte
    next    []*SkipListNode  // Next pointers at each level
}

type SkipList struct {
    head   *SkipListNode
    level  int               // Current max level
    mutex  sync.RWMutex      // For concurrent access
    length int
}

func NewSkipList() *SkipList {
    // Create head node with max levels
    head := &SkipListNode{
        next: make([]*SkipListNode, MAX_LEVEL),
    }
    
    return &SkipList{
        head:  head,
        level: 0,
    }
}

// randomLevel returns a random level (probabilistic)
func (sl *SkipList) randomLevel() int {
    level := 0
    for level < MAX_LEVEL-1 && rand.Float64() < PROBABILITY {
        level++
    }
    return level
}

// Insert adds a key-value pair
func (sl *SkipList) Insert(key, value []byte) {
    sl.mutex.Lock()
    defer sl.mutex.Unlock()
    
    // Track nodes we need to update at each level
    update := make([]*SkipListNode, MAX_LEVEL)
    
    // Start from head
    current := sl.head
    
    // Find position to insert at each level
    for i := sl.level; i >= 0; i-- {
        for current.next[i] != nil && compare(current.next[i].key, key) < 0 {
            current = current.next[i]
        }
        update[i] = current
    }
    
    // Move to the node at level 0 (where key would be)
    current = current.next[0]
    
    // If key exists, update value
    if current != nil && compare(current.key, key) == 0 {
        current.value = value
        return
    }
    
    // Generate random level for new node
    newLevel := sl.randomLevel()
    
    // If new level is higher than current max, update pointers
    if newLevel > sl.level {
        for i := sl.level + 1; i <= newLevel; i++ {
            update[i] = sl.head
        }
        sl.level = newLevel
    }
    
    // Create new node
    newNode := &SkipListNode{
        key:   key,
        value: value,
        next:  make([]*SkipListNode, newLevel+1),
    }
    
    // Insert the node at each level
    for i := 0; i <= newLevel; i++ {
        newNode.next[i] = update[i].next[i]
        update[i].next[i] = newNode
    }
    
    sl.length++
}

// Search finds a key and returns its value
func (sl *SkipList) Search(key []byte) ([]byte, bool) {
    sl.mutex.RLock()
    defer sl.mutex.RUnlock()
    
    current := sl.head
    
    // Start from highest level and work down
    for i := sl.level; i >= 0; i-- {
        for current.next[i] != nil && compare(current.next[i].key, key) < 0 {
            current = current.next[i]
        }
    }
    
    // Move to the potential node
    current = current.next[0]
    
    if current != nil && compare(current.key, key) == 0 {
        return current.value, true
    }
    
    return nil, false
}

// Delete removes a key
func (sl *SkipList) Delete(key []byte) bool {
    sl.mutex.Lock()
    defer sl.mutex.Unlock()
    
    update := make([]*SkipListNode, MAX_LEVEL)
    current := sl.head
    
    // Find node to delete at each level
    for i := sl.level; i >= 0; i-- {
        for current.next[i] != nil && compare(current.next[i].key, key) < 0 {
            current = current.next[i]
        }
        update[i] = current
    }
    
    current = current.next[0]
    
    // Key not found
    if current == nil || compare(current.key, key) != 0 {
        return false
    }
    
    // Remove node at each level
    for i := 0; i <= sl.level; i++ {
        if update[i].next[i] != current {
            break
        }
        update[i].next[i] = current.next[i]
    }
    
    // Reduce level if highest level is empty
    for sl.level > 0 && sl.head.next[sl.level] == nil {
        sl.level--
    }
    
    sl.length--
    return true
}

// Range returns all key-value pairs between start and end
func (sl *SkipList) Range(start, end []byte) [][2][]byte {
    sl.mutex.RLock()
    defer sl.mutex.RUnlock()
    
    result := make([][2][]byte, 0)
    
    // Find start position
    current := sl.head
    for i := sl.level; i >= 0; i-- {
        for current.next[i] != nil && compare(current.next[i].key, start) < 0 {
            current = current.next[i]
        }
    }
    
    // Collect until end
    current = current.next[0]
    for current != nil && compare(current.key, end) <= 0 {
        result = append(result, [2][]byte{current.key, current.value})
        current = current.next[0]
    }
    
    return result
}

// Length returns number of elements
func (sl *SkipList) Length() int {
    sl.mutex.RLock()
    defer sl.mutex.RUnlock()
    return sl.length
}

// Print shows the skip list structure
func (sl *SkipList) Print() {
    sl.mutex.RLock()
    defer sl.mutex.RUnlock()
    
    fmt.Println("=== Skip List ===")
    fmt.Printf("Level: %d, Length: %d\n", sl.level, sl.length)
    
    for i := sl.level; i >= 0; i-- {
        fmt.Printf("Level %d: ", i)
        current := sl.head.next[i]
        for current != nil {
            fmt.Printf("[%s=%s] ", current.key, current.value)
            current = current.next[i]
        }
        fmt.Println()
    }
}

// Helper: compare two byte slices
func compare(a, b []byte) int {
    if len(a) < len(b) {
        for i := 0; i < len(a); i++ {
            if a[i] != b[i] {
                return int(a[i]) - int(b[i])
            }
        }
        return -1
    } else {
        for i := 0; i < len(b); i++ {
            if a[i] != b[i] {
                return int(a[i]) - int(b[i])
            }
        }
        if len(a) == len(b) {
            return 0
        }
        return 1
    }
}

func main() {
    rand.Seed(time.Now().UnixNano())
    sl := NewSkipList()
    
    fmt.Println("=== Inserting 20 key-value pairs ===")
    for i := 0; i < 20; i++ {
        key := []byte(fmt.Sprintf("key_%d", i))
        value := []byte(fmt.Sprintf("value_%d", i))
        sl.Insert(key, value)
        fmt.Printf("Inserted %s\n", key)
    }
    
    sl.Print()
    
    fmt.Println("\n=== Searching ===")
    testKeys := []string{"key_5", "key_10", "key_99", "key_15"}
    for _, k := range testKeys {
        val, found := sl.Search([]byte(k))
        fmt.Printf("Search '%s': found=%v, value=%s\n", k, found, val)
    }
    
    fmt.Println("\n=== Range Query [key_5, key_15] ===")
    results := sl.Range([]byte("key_5"), []byte("key_15"))
    for _, r := range results {
        fmt.Printf("%s = %s\n", r[0], r[1])
    }
    
    fmt.Println("\n=== Deleting key_10 ===")
    deleted := sl.Delete([]byte("key_10"))
    fmt.Printf("Deleted: %v\n", deleted)
    
    _, found := sl.Search([]byte("key_10"))
    fmt.Printf("Search 'key_10' after delete: found=%v\n", found)
    
    sl.Print()
    
    fmt.Println("\n=== Concurrent Insert Demo ===")
    var wg sync.WaitGroup
    for i := 100; i < 200; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            key := []byte(fmt.Sprintf("concurrent_%d", i))
            value := []byte(fmt.Sprintf("val_%d", i))
            sl.Insert(key, value)
        }(i)
    }
    wg.Wait()
    
    fmt.Printf("Final length: %d\n", sl.Length())
}