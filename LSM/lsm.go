package main

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "math/rand"
    "os"
    "sort"
    "sync"
    "time"
)

// ============================================
// Level 0: Skip List (MemTable)
// ============================================

type SkipListNode struct {
    key   []byte
    value []byte
    next  []*SkipListNode
}

type SkipList struct {
    head  *SkipListNode
    level int
    mu    sync.RWMutex
    size  int
}

func NewSkipList() *SkipList {
    head := &SkipListNode{next: make([]*SkipListNode, 16)}
    return &SkipList{head: head, level: 0}
}

func (sl *SkipList) randomLevel() int {
    level := 0
    for level < 15 && rand.Float64() < 0.5 {
        level++
    }
    return level
}

func (sl *SkipList) Put(key, value []byte) {
    sl.mu.Lock()
    defer sl.mu.Unlock()
    
    update := make([]*SkipListNode, 16)
    current := sl.head
    
    for i := sl.level; i >= 0; i-- {
        for current.next[i] != nil && bytes.Compare(current.next[i].key, key) < 0 {
            current = current.next[i]
        }
        update[i] = current
    }
    
    current = current.next[0]
    if current != nil && bytes.Equal(current.key, key) {
        current.value = value
        return
    }
    
    newLevel := sl.randomLevel()
    if newLevel > sl.level {
        for i := sl.level + 1; i <= newLevel; i++ {
            update[i] = sl.head
        }
        sl.level = newLevel
    }
    
    newNode := &SkipListNode{
        key:   key,
        value: value,
        next:  make([]*SkipListNode, newLevel+1),
    }
    
    for i := 0; i <= newLevel; i++ {
        newNode.next[i] = update[i].next[i]
        update[i].next[i] = newNode
    }
    sl.size++
}

func (sl *SkipList) Get(key []byte) ([]byte, bool) {
    sl.mu.RLock()
    defer sl.mu.RUnlock()
    
    current := sl.head
    for i := sl.level; i >= 0; i-- {
        for current.next[i] != nil && bytes.Compare(current.next[i].key, key) < 0 {
            current = current.next[i]
        }
    }
    
    current = current.next[0]
    if current != nil && bytes.Equal(current.key, key) {
        return current.value, true
    }
    return nil, false
}

func (sl *SkipList) Collect() [][2][]byte {
    sl.mu.RLock()
    defer sl.mu.RUnlock()
    
    result := make([][2][]byte, 0, sl.size)
    current := sl.head.next[0]
    for current != nil {
        result = append(result, [2][]byte{current.key, current.value})
        current = current.next[0]
    }
    return result
}

func (sl *SkipList) Size() int {
    sl.mu.RLock()
    defer sl.mu.RUnlock()
    return sl.size
}

// ============================================
// Level 1: Write-Ahead Log (WAL)
// ============================================

type WAL struct {
    file *os.File
    mu   sync.Mutex
}

func NewWAL(path string) (*WAL, error) {
    f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
    if err != nil {
        return nil, err
    }
    return &WAL{file: f}, nil
}

func (w *WAL) Append(key, value []byte) error {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    // Write: [4-byte key length][key][4-byte value length][value]
    keyLen := uint32(len(key))
    valLen := uint32(len(value))
    
    binary.Write(w.file, binary.LittleEndian, keyLen)
    w.file.Write(key)
    binary.Write(w.file, binary.LittleEndian, valLen)
    w.file.Write(value)
    
    return w.file.Sync() // Force to disk
}

func (w *WAL) Recover() ([][2][]byte, error) {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    w.file.Seek(0, 0)
    
    entries := make([][2][]byte, 0)
    for {
        var keyLen, valLen uint32
        
        err := binary.Read(w.file, binary.LittleEndian, &keyLen)
        if err != nil {
            break
        }
        
        key := make([]byte, keyLen)
        w.file.Read(key)
        
        binary.Read(w.file, binary.LittleEndian, &valLen)
        value := make([]byte, valLen)
        w.file.Read(value)
        
        entries = append(entries, [2][]byte{key, value})
    }
    
    return entries, nil
}

func (w *WAL) Close() error {
    return w.file.Close()
}

// ============================================
// Level 2: SSTable (Sorted String Table)
// ============================================

type SSTable struct {
    path   string
    keys   [][]byte
    values [][]byte
}

func WriteSSTable(path string, entries [][2][]byte) (*SSTable, error) {
    // Sort by key
    sort.Slice(entries, func(i, j int) bool {
        return bytes.Compare(entries[i][0], entries[j][0]) < 0
    })
    
    // Remove duplicates (keep latest)
    unique := make([][2][]byte, 0, len(entries))
    for i := 0; i < len(entries); i++ {
        if i > 0 && bytes.Equal(entries[i][0], entries[i-1][0]) {
            continue
        }
        unique = append(unique, entries[i])
    }
    
    // Write to file (simplified — in real DBs, this would have blocks and index)
    f, err := os.Create(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    
    keys := make([][]byte, len(unique))
    values := make([][]byte, len(unique))
    
    for i, entry := range unique {
        keys[i] = entry[0]
        values[i] = entry[1]
        
        // Write key
        binary.Write(f, binary.LittleEndian, uint32(len(entry[0])))
        f.Write(entry[0])
        
        // Write value
        binary.Write(f, binary.LittleEndian, uint32(len(entry[1])))
        f.Write(entry[1])
    }
    
    return &SSTable{
        path:   path,
        keys:   keys,
        values: values,
    }, nil
}

func (s *SSTable) Get(key []byte) ([]byte, bool) {
    // Binary search
    idx := sort.Search(len(s.keys), func(i int) bool {
        return bytes.Compare(s.keys[i], key) >= 0
    })
    
    if idx < len(s.keys) && bytes.Equal(s.keys[idx], key) {
        return s.values[idx], true
    }
    return nil, false
}

// ============================================
// Level 3: Compaction
// ============================================

func Compact(sstables []*SSTable, outputPath string) (*SSTable, error) {
    // Collect all entries
    allEntries := make([][2][]byte, 0)
    for _, sst := range sstables {
        for i := 0; i < len(sst.keys); i++ {
            allEntries = append(allEntries, [2][]byte{sst.keys[i], sst.values[i]})
        }
    }
    
    // Remove duplicates (keep latest — later in slice)
    seen := make(map[string]bool)
    merged := make([][2][]byte, 0, len(allEntries))
    for i := len(allEntries) - 1; i >= 0; i-- {
        keyStr := string(allEntries[i][0])
        if !seen[keyStr] {
            seen[keyStr] = true
            merged = append([][2][]byte{allEntries[i]}, merged...)
        }
    }
    
    return WriteSSTable(outputPath, merged)
}

// ============================================
// Level 4: Complete LSM Engine
// ============================================

type LSMTree struct {
    memtable  *SkipList
    wal       *WAL
    sstables  []*SSTable
    maxSize   int // Max memtable entries before flush
    flushCh   chan bool
    stopCh    chan bool
    wg        sync.WaitGroup
    mu        sync.RWMutex
}

func NewLSMTree(dataDir string, maxMemSize int) (*LSMTree, error) {
    os.MkdirAll(dataDir, 0755)
    
    wal, err := NewWAL(dataDir + "/wal.log")
    if err != nil {
        return nil, err
    }
    
    lsm := &LSMTree{
        memtable: NewSkipList(),
        wal:      wal,
        sstables: make([]*SSTable, 0),
        maxSize:  maxMemSize,
        flushCh:  make(chan bool, 1),
        stopCh:   make(chan bool),
    }
    
    // Recover from WAL
    entries, err := wal.Recover()
    if err == nil {
        for _, entry := range entries {
            lsm.memtable.Put(entry[0], entry[1])
        }
    }
    
    // Load existing SSTables
    files, _ := os.ReadDir(dataDir)
    for _, file := range files {
        // Simplified loading
        _ = file
    }
    
    // Start background flusher
    lsm.wg.Add(1)
    go lsm.backgroundFlusher()
    
    return lsm, nil
}

func (lsm *LSMTree) Put(key, value []byte) error {
    // 1. Write to WAL (durability)
    if err := lsm.wal.Append(key, value); err != nil {
        return err
    }
    
    // 2. Write to MemTable
    lsm.mu.Lock()
    lsm.memtable.Put(key, value)
    shouldFlush := lsm.memtable.Size() >= lsm.maxSize
    lsm.mu.Unlock()
    
    // 3. Trigger flush if needed
    if shouldFlush {
        select {
        case lsm.flushCh <- true:
        default:
        }
    }
    
    return nil
}

func (lsm *LSMTree) Get(key []byte) ([]byte, bool) {
    lsm.mu.RLock()
    defer lsm.mu.RUnlock()
    
    // 1. Check MemTable
    if value, ok := lsm.memtable.Get(key); ok {
        return value, true
    }
    
    // 2. Check SSTables (newest first)
    for i := len(lsm.sstables) - 1; i >= 0; i-- {
        if value, ok := lsm.sstables[i].Get(key); ok {
            return value, true
        }
    }
    
    return nil, false
}

func (lsm *LSMTree) Flush() error {
    lsm.mu.Lock()
    
    // Get current memtable
    currentMem := lsm.memtable
    entries := currentMem.Collect()
    
    // Create new memtable
    lsm.memtable = NewSkipList()
    lsm.mu.Unlock()
    
    if len(entries) == 0 {
        return nil
    }
    
    // Write to SSTable
    sst, err := WriteSSTable(fmt.Sprintf("sst_%d.db", time.Now().UnixNano()), entries)
    if err != nil {
        return err
    }
    
    lsm.mu.Lock()
    lsm.sstables = append(lsm.sstables, sst)
    lsm.mu.Unlock()
    
    // Rotate WAL (create new)
    lsm.wal.Close()
    newWAL, _ := NewWAL("wal.log.new")
    lsm.wal = newWAL
    
    return nil
}

func (lsm *LSMTree) backgroundFlusher() {
    defer lsm.wg.Done()
    
    for {
        select {
        case <-lsm.flushCh:
            lsm.Flush()
        case <-lsm.stopCh:
            return
        }
    }
}

func (lsm *LSMTree) Close() error {
    close(lsm.stopCh)
    lsm.wg.Wait()
    lsm.Flush()
    return lsm.wal.Close()
}

func (lsm *LSMTree) Stats() string {
    lsm.mu.RLock()
    defer lsm.mu.RUnlock()
    
    totalKeys := 0
    for _, sst := range lsm.sstables {
        totalKeys += len(sst.keys)
    }
    
    return fmt.Sprintf(
        "MemTable: %d keys\nSSTables: %d\nTotal keys on disk: %d",
        lsm.memtable.Size(), len(lsm.sstables), totalKeys,
    )
}

// ============================================
// Demonstration
// ============================================

func main() {
    rand.NewSource(time.Now().UnixNano())
    
    fmt.Println("=== LSM Tree Demonstration ===")
    
    // Create LSM tree
    lsm, err := NewLSMTree("./lsm_data", 100) // Flush after 100 entries
    if err != nil {
        panic(err)
    }
    defer lsm.Close()
    
    // Phase 1: Insert 500 entries
    fmt.Println("Phase 1: Inserting 500 key-value pairs")
    for i := 0; i < 500; i++ {
        key := []byte(fmt.Sprintf("key_%04d", i))
        value := []byte(fmt.Sprintf("value_%d", i))
        lsm.Put(key, value)
        
        if (i+1)%100 == 0 {
            fmt.Printf("  Inserted %d entries\n", i+1)
        }
    }
    
    fmt.Println("\nPhase 2: Engine statistics")
    fmt.Println(lsm.Stats())
    
    fmt.Println("\nPhase 3: Reading some values")
    testKeys := []string{"key_0000", "key_0250", "key_0499", "key_9999"}
    for _, k := range testKeys {
        val, found := lsm.Get([]byte(k))
        if found {
            fmt.Printf("  %s -> %s\n", k, val)
        } else {
            fmt.Printf("  %s -> NOT FOUND\n", k)
        }
    }
    
    fmt.Println("\nPhase 4: Inserting more (forces flush)")
    for i := 500; i < 550; i++ {
        key := []byte(fmt.Sprintf("key_%04d", i))
        value := []byte(fmt.Sprintf("value_%d", i))
        lsm.Put(key, value)
    }
    fmt.Println(lsm.Stats())
    
    fmt.Println("\nPhase 5: Updating existing keys")
    lsm.Put([]byte("key_0100"), []byte("updated_value"))
    val, found := lsm.Get([]byte("key_0100"))
    fmt.Printf("  key_0100 after update: %s (found: %v)\n", val, found)
    
    fmt.Println("\nPhase 6: After multiple flushes")
    for i := 0; i < 3; i++ {
        lsm.Flush()
    }
    fmt.Println(lsm.Stats())
    
    fmt.Println("\n LSM Tree demonstration complete")
    fmt.Println("\nKey concepts demonstrated:")
    fmt.Println("  • WAL provides crash durability")
    fmt.Println("  • MemTable gives fast writes")
    fmt.Println("  • SSTables persist data to disk")
    fmt.Println("  • Background flush prevents memory blowup")
}