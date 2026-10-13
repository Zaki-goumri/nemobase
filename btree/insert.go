package btree

import (
	"bytes"
	"encoding/binary"
)

func nodeLookupLE(node BNode, key []byte) uint16 {
	nkeys := node.nkeys()
	found := uint16(0)
	for i := uint16(1); i < nkeys; i++ {
		cmp := bytes.Compare(node.getKey(i), key)
		if cmp <= 0 {
			found = i
		}
		if cmp >= 0 {
			break
		}
	}
	return found
}

func nodeAppendKV(newNode BNode, idx uint16, ptr uint64, key, val []byte) {
	newNode.setPtr(idx, ptr)
	pos := newNode.kvPos(idx)
	binary.LittleEndian.PutUint16(newNode[pos+0:], uint16(len(key)))
	binary.LittleEndian.PutUint16(newNode[pos+2:], uint16(len(val)))
	copy(newNode[pos+4:], key)
	copy(newNode[pos+4+uint16(len(key)):], val)
	newNode.setOffset(idx+1, newNode.getOffset(idx)+4+uint16(len(key))+uint16(len(val)))
}

func nodeAppendRange(newNode, oldNode BNode, dstNew, srcOld, n uint16) {
	for i := uint16(0); i < n; i++ {
		dst, src := dstNew+i, srcOld+i
		nodeAppendKV(newNode, dst, oldNode.getPtr(src), oldNode.getKey(src), oldNode.getVal(src))
	}
}

func leafInsert(newNode, oldNode BNode, idx uint16, key, val []byte) {
	newNode.setHeader(BNODE_LEAF, oldNode.nkeys()+1)
	nodeAppendRange(newNode, oldNode, 0, 0, idx)
	nodeAppendKV(newNode, idx, 0, key, val)
	nodeAppendRange(newNode, oldNode, idx+1, idx, oldNode.nkeys()-idx)
}

func leafUpdate(newNode, oldNode BNode, idx uint16, key, val []byte) {
	newNode.setHeader(BNODE_LEAF, oldNode.nkeys())
	nodeAppendRange(newNode, oldNode, 0, 0, idx)
	nodeAppendKV(newNode, idx, 0, key, val)
	nodeAppendRange(newNode, oldNode, idx+1, idx, oldNode.nkeys()-idx-1)

}

func nodeInsert(tree *BTree, newNode, node BNode, idx uint16, key, val []byte) {
	kptr := node.getPtr(idx)
	knode := tree.get(kptr)
	tree.del(kptr)

	knode = treeInsert(tree, knode, key, val)

	nsplit, split := nodeSplit3(knode)
	nodeReplaceKidN(tree, newNode, node, idx, split[:nsplit]...)
}

func treeInsert(tree *BTree, node BNode, key, val []byte) BNode {
	newNode := make(BNode, 2*PAGE_SIZE)
	idx := nodeLookupLE(node, key)
	switch node.btype() {
	case BNODE_NODE:
		nodeInsert(tree, newNode, node, idx, key, val)
	case BNODE_LEAF:
		if idx < node.nkeys() && bytes.Equal(node.getKey(idx), key) {
			leafUpdate(newNode, node, idx, key, val)
		} else {
			leafInsert(newNode, node, idx+1, key, val)
		}
	}
	return newNode
}

func nodeSplit2(left, right, old BNode) {
	nleft := old.nkeys() / 2
	leftBytes := func() uint16 {
		return HEADER + 8*nleft + 2*nleft + old.getOffset(nleft)
	}
	for leftBytes() > PAGE_SIZE {
		nleft--
	}

	rightBytes := func() uint16 {
		return old.nbytes() - leftBytes() + HEADER
	}
	for rightBytes() > PAGE_SIZE {
		nleft++
	}

	nright := old.nkeys() - nleft

	left.setHeader(old.btype(), nleft)
	right.setHeader(old.btype(), nright)
	nodeAppendRange(left, old, 0, 0, nleft)
	nodeAppendRange(right, old, 0, nleft, nright)
}

func nodeSplit3(old BNode) (uint16, [3]BNode) {
	if old.nbytes() <= PAGE_SIZE {
		return 1, [3]BNode{old}
	}
	left := make(BNode, 2*PAGE_SIZE)
	right := make(BNode, PAGE_SIZE)
	nodeSplit2(left, right, old)
	if left.btype() <= PAGE_SIZE {
		return 2, [3]BNode{left, right}
	}
	leftLeft := make(BNode, 2*PAGE_SIZE)
	middle := make(BNode, PAGE_SIZE)
	nodeSplit2(leftLeft, middle, left)
	return 3, [3]BNode{leftLeft, middle, right}
}

func nodeReplaceKidN(tree *BTree, newNode, node BNode, idx uint16, kids ...BNode) {}
func (tree *BTree) Insert(key, val []byte)                                        {}
