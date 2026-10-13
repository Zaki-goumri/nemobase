package btree

import "encoding/binary"

const HEADER = 4
const PAGE_SIZE = 4096
const (
	BNODE_NODE uint16 = 1
	BNODE_LEAF uint16 = 2
)

/*
byte:    0        4              4+8n           4+8n+2n                end

	|--------|--------------|---------------|------------------------|
	| header |   pointers   |    offsets     |      key-values        |
	|  4B    |   n * 8B     |    n * 2B      |     (variable size)    |
	|--------|--------------|---------------|------------------------|
	 type|nkeys   child page#s   entry end       klen|vlen|key|val ...
	  2B  2B      (leaf: unused)  positions        repeated per entry
*/
type BNode []byte

type BTree struct {
	root uint64
	get  func(uint64) BNode
	new  func(BNode) uint64
	del  func(uint64)
}

func (node BNode) btype() uint16 {
	return binary.LittleEndian.Uint16(node[0:2])
}

func (node BNode) nkeys() uint16 {
	return binary.LittleEndian.Uint16(node[2:4])
}

func (node BNode) getPtr(idx uint16) uint64 {
	pos := HEADER + 8*idx
	return binary.LittleEndian.Uint64(node[pos:])
}
func (node BNode) setPtr(idx uint16, val uint64) {
	pos := HEADER + 8*idx
	binary.LittleEndian.PutUint64(node[pos:], val)
}

func (node BNode) setHeader(btype uint16, nkeys uint16) {
	binary.LittleEndian.PutUint16(node[0:2], btype)
	binary.LittleEndian.PutUint16(node[2:4], nkeys)
}

func (node BNode) getOffset(idx uint16) uint16 {
	if idx == 0 {
		return 0
	}
	pos := HEADER + 8*node.nkeys() + 2*(idx-1)
	return binary.LittleEndian.Uint16(node[pos:])
}

func (node BNode) setOffset(idx uint16, offset uint16) {
	pos := HEADER + 8*node.nkeys() + 2*(idx-1)
	binary.LittleEndian.PutUint16(node[pos:], offset)
}

func (node BNode) kvPos(idx uint16) uint16 {
	return HEADER + 8*node.nkeys() + 2*node.nkeys() + node.getOffset(idx)
}

func (node BNode) getKey(idx uint16) []byte {
	pos := node.kvPos(idx)
	klen := binary.LittleEndian.Uint16(node[pos:])
	return node[pos+4:][:klen]
}

func (node BNode) getVal(idx uint16) []byte {
	pos := node.kvPos(idx)
	klen := binary.LittleEndian.Uint16(node[pos+0:])
	vlen := binary.LittleEndian.Uint16(node[pos+2:])
	return node[pos+4+klen:][:vlen]
}

func (node BNode) nbytes() uint16 {
	return node.kvPos(node.nkeys())
}
