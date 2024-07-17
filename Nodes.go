package main

import (
	"bytes"
	"encoding/binary"
)

const (
	//First 2 bits
	LEAF   = 2
	BRANCH = 1
	//HEADER is 4 bits long. first 2 are the type and the second 2 are the number of keys the node has
	HEADER       = 4
	PAGE_SIZE    = 4096
	MAX_KEY_SIZE = 1000
	MAX_VAL_SIZE = 3000
	OFFSET_SIZE  = 2

	IndexERRMSG = "Index is bigger than expected in the function: "
)

/*
	Node byte struct

header

	node[0:2]type
	node[2:4]keys

pointers

	number/index of keys * 8

offsets

	number/index of keys * 2

actual kv

	key len 2
	val len 2
	key
	value
*/
type Node []byte

func (n Node) getType() uint16 {
	return binary.LittleEndian.Uint16(n[0:2])
}

func (n Node) getnKeys() uint16 {
	return binary.LittleEndian.Uint16(n[2:4])
}

func (n Node) setHeader(nType, nKeys uint16) {
	binary.LittleEndian.PutUint16(n[0:2], nType)
	binary.LittleEndian.PutUint16(n[2:4], nKeys)
}

func (n Node) getChildPtr(i uint16) uint64 {
	if i > n.getnKeys() {
		panic(IndexERRMSG + "getChildPtr")
	}
	pos := HEADER + 8*i
	return binary.LittleEndian.Uint64(n[pos:])
}

func (n Node) setChildPtr(i uint16, val uint64) {
	if i > n.getnKeys() {
		panic(IndexERRMSG + "setChildPter")
	}
	pos := HEADER + 8*i
	binary.LittleEndian.PutUint64(n[pos:], val)
}

func getOffsetPos(n Node, i uint16) uint16 {
	if i > n.getnKeys() {
		panic(IndexERRMSG + "getOffsetPos")
	}
	return HEADER + 8*n.getnKeys() + OFFSET_SIZE*(i-1)
}

func (n Node) getOffset(i uint16) uint16 {
	if i == 0 {
		return 0
	}
	return binary.LittleEndian.Uint16(n[getOffsetPos(n, i):])
}

func (n Node) setOffset(i, offset uint16) {
	if i > n.getnKeys() {

		panic(IndexERRMSG + "setOffset")
	}
	binary.LittleEndian.PutUint16(n[getOffsetPos(n, i):], offset)
}

func (n Node) getKVPos(i uint16) uint16 {
	if i > n.getnKeys() {
		panic(IndexERRMSG + "getKVPos")
	}
	return HEADER + 8*n.getnKeys() + 2*n.getnKeys() + n.getOffset(i)
}

func (n Node) getKey(i uint16) []byte {
	if i > n.getnKeys() {
		panic(IndexERRMSG + "getKey")
	}
	klen := binary.LittleEndian.Uint16(n[n.getKVPos(i):])
	return n[n.getKVPos(i)+4:][:klen]

}

func (n Node) getVal(i uint16) []byte {
	if i > n.getnKeys() {
		panic(IndexERRMSG + "getVal")
	}
	klen := binary.LittleEndian.Uint16(n[n.getKVPos(i):])
	vlen := binary.LittleEndian.Uint16(n[n.getKVPos(i)+2:])
	return n[n.getKVPos(i)+4+klen:][:vlen]
}

func (n Node) nBytes() uint16 {
	return n.getKVPos(n.getnKeys())
}

func nodeLookupLE(n Node, key []byte) uint16 {
	found := uint16(0)
	for i := uint16(1); i < n.getnKeys(); i++ {
		cur := bytes.Compare(n.getKey(i), key)
		if cur <= 0 {
			found = i
		}
		if cur > 0 {
			break
		}
	}
	return found
}

func insertLeaf(new, src Node, i uint16, key, val []byte) {
	new.setHeader(LEAF, src.getnKeys()+1)
	copyIndexRange(new, src, 0, 0, i)
	updateIndexKV(new, i, 0, key, val)
	copyIndexRange(new, src, 0, 0, i)
}

func copyIndexRange(new, src Node, newIndex, srcIndex, nK uint16) {
	for i := srcIndex; i < nK; i++ {
		pos := src.getKVPos(i)
		klen := binary.LittleEndian.Uint16(src[pos:])
		copy(new[pos:], src[pos:])
		copy(new[pos+2:], src[pos+2:])
		copy(new[pos+4:], src[pos+4:])
		copy(new[pos+4+klen:], src[pos+4+klen:])

	}
}

func updateIndexKV(new Node, i uint16, ptr uint64, key, val []byte) {
	new.setChildPtr(i, ptr)
	pos := new.getKVPos(i)
	binary.LittleEndian.PutUint16(new[pos:], uint16(len(key)))
	binary.LittleEndian.PutUint16(new[pos+2:], uint16(len(val)))
	copy(new[pos+4:], key)
	copy(new[pos+4+uint16(len(key)):], val)
}
