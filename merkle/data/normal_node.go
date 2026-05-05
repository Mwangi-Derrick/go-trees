package data

// Normal node
type NormalNode struct {
    Data  string
    Left  *NormalNode   // just an address
    Right *NormalNode
}