package personalDB

type BTree struct {
	root uint64
	get  func(uint64) []byte
}
