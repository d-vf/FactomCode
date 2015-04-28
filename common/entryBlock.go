// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package common

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	//	"time"
)

const (
	eBlockHeaderLen = 80
	Separator       = "/"
)

// Used for communication via the APIs
type EChain struct {

	ChainID *Hash
	Name       [][]byte
	FirstEntry *Entry

	//Not Marshalized
	//Blocks       []*EBlock
	NextBlock       *EBlock
	NextBlockHeight uint32
	BlockMutex      sync.Mutex
}

type EBlock struct {
	//Marshalized
	Header    *EBlockHeader
	EBEntries []*EBEntry

	//Not Marshalized
	EBHash     *Hash
	MerkleRoot *Hash
	Chain      *EChain
	IsSealed   bool
}

type EBInfo struct {
	EBHash     *Hash
	MerkleRoot *Hash
	DBHash     *Hash
	DBBlockNum uint64
	ChainID    *Hash
}

type EBlockHeader struct {
	Version    byte
	ChainID    *Hash
	BodyMR     *Hash
	PrevKeyMR  *Hash
	PrevHash   *Hash
	EBHeight   uint32
	DBHeight   uint32
	StartTime  uint64
	EntryCount uint32
}

type EBEntry struct {
	EntryHash *Hash
}

func NewEBEntry(h *Hash) *EBEntry {
	return &EBEntry{
		EntryHash: h}
}

func (e *EBEntry) EncodableFields() map[string]reflect.Value {
	fields := map[string]reflect.Value{
		`EntryHash`: reflect.ValueOf(e.EntryHash),
	}
	return fields
}

func (e *EBEntry) MarshalBinary() ([]byte, error) {
	return e.EntryHash.Bytes, nil
}

func (e *EBEntry) UnmarshalBinary(data []byte) (err error) {
	return nil
}

func (e *EBEntry) ShaHash() *Hash {
	return Sha(e.EntryHash.Bytes)
}

func (b *EBlockHeader) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.WriteByte(b.Version)	
	buf.Write(b.ChainID.Bytes)

	buf.Write(b.BodyMR.Bytes)
	buf.Write(b.PrevKeyMR.Bytes)
	buf.Write(b.PrevHash.Bytes)

	binary.Write(&buf, binary.BigEndian, b.EBHeight)
	binary.Write(&buf, binary.BigEndian, b.DBHeight)
	binary.Write(&buf, binary.BigEndian, b.EntryCount)

	return buf.Bytes(), err
}

func (b *EBlockHeader) MarshalledSize() int {
	var size int = 0

	size += 1
	size += 4
	size += HASH_LENGTH // b.ChainID.MarshalledSize()
	size += HASH_LENGTH // b.BodyMR.MarshalledSize()
	size += HASH_LENGTH // b.PrevKeyMR.MarshalledSize()
	size += HASH_LENGTH // b.PrevHash.MarshalledSize()
	size += 4
	size += 4
	size += 4

	return size
}

func (b *EBlockHeader) UnmarshalBinary(data []byte) (err error) {
fmt.Println("Here Unmarshal EBblock header")
	b.Version, data = data[0], data[1:]

	b.ChainID,data = UnmarshalHash(data)
	b.BodyMR,data = UnmarshalHash(data)
	b.PrevKeyMR,data = UnmarshalHash(data)
	b.PrevHash,data = UnmarshalHash(data)
    
	b.EBHeight,   data = binary.BigEndian.Uint32(data[0:4]), data[4:]
    b.DBHeight,   data = binary.BigEndian.Uint32(data[0:4]), data[4:]
	b.EntryCount, data = binary.BigEndian.Uint32(data[0:4]), data[4:]

	return nil
}

func CreateBlock(chain *EChain, prev *EBlock, capacity uint) (b *EBlock, err error) {
	if prev == nil && chain.NextBlockHeight != 0 {
		return nil, errors.New("Previous block cannot be nil")
	} else if prev != nil && chain.NextBlockHeight == 0 {
		return nil, errors.New("Origin block cannot have a parent block")
	}

	b = new(EBlock)

	b.Header = new(EBlockHeader)
	b.Header.Version = VERSION_0
	b.Header.ChainID = chain.ChainID
	b.Header.EBHeight = chain.NextBlockHeight
	if prev == nil {
		b.Header.PrevHash = NewHash()
		b.Header.PrevKeyMR = NewHash()
	} else {
		b.Header.PrevHash = prev.EBHash
		if prev.MerkleRoot == nil {
			prev.BuildMerkleRoot()
		}
		b.Header.PrevKeyMR = prev.MerkleRoot
	}

	b.Chain = chain

	b.EBEntries = make([]*EBEntry, 0, capacity)

	b.IsSealed = false

	return b, err
}

func (b *EBlock) AddEBEntry(e *Entry) (err error) {
	h, err := CreateHash(e)
	if err != nil {
		return
	}

	ebEntry := NewEBEntry(h)

	b.EBEntries = append(b.EBEntries, ebEntry)

	return
}

func (b *EBlock) AddEndOfMinuteMarker(eomType byte) (err error) {
	bytes := make([]byte, 32)
	bytes[31] = eomType

	h, _ := NewShaHash(bytes)

	ebEntry := NewEBEntry(h)

	b.EBEntries = append(b.EBEntries, ebEntry)

	return
}

func (block *EBlock) BuildMerkleRoot() (err error) {
	// Create the Entry Block Boday Merkle Root from EB Entries
	hashes := make([]*Hash, 0, len(block.EBEntries))
	for _, entry := range block.EBEntries {
		hashes = append(hashes, entry.EntryHash)
	}
	merkle := BuildMerkleTreeStore(hashes)

	// Create the Entry Block Key Merkle Root from the hash of Header and the Body Merkle Root
	hashes = make([]*Hash, 0, 2)
	binaryEBHeader, _ := block.Header.MarshalBinary()
	hashes = append(hashes, Sha(binaryEBHeader))
	hashes = append(hashes, block.Header.BodyMR)
	merkle = BuildMerkleTreeStore(hashes)
	block.MerkleRoot = merkle[len(merkle)-1] // MerkleRoot is not marshalized in Entry Block

	return
}

func (e *EBlock) EncodableFields() map[string]reflect.Value {
	fields := map[string]reflect.Value{
		`Header`:    reflect.ValueOf(e.Header),
		`EBEntries`: reflect.ValueOf(e.EBEntries),
	}
	return fields
}

func (b *EBlock) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	data, err = b.Header.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	//binary.Write(&buf, binary.BigEndian, b.BlockID)
	//data, _ = b.PreviousHash.MarshalBinary()
	//buf.Write(data)

	count := uint64(len(b.EBEntries))
	binary.Write(&buf, binary.BigEndian, count)
	for i := uint64(0); i < count; i = i + 1 {
		data, err := b.EBEntries[i].MarshalBinary()
		if err != nil {
			return nil, err
		}

		buf.Write(data)
	}

	return buf.Bytes(), err
}

func (b *EBlock) MarshalledSize() (size int) {
	size += b.Header.MarshalledSize()
	size += 4 // len(Entries) uint64
	size += len(b.EBEntries)*HASH_LENGTH

	return size
}

func (b *EBlock) UnmarshalBinary(data []byte) (err error) {

    b.Header = new(EBlockHeader)
	b.Header.UnmarshalBinary(data)

	data = data[b.Header.MarshalledSize():]

	count, data := binary.BigEndian.Uint32(data[0:4]), data[4:]

	b.EBEntries = make([]*EBEntry, count)
	for i := uint32(0); i < count; i = i + 1 {
		b.EBEntries[i] = new(EBEntry)
		b.EBEntries[i].EntryHash, data = UnmarshalHash(data)
	}

	return nil
}



// To decode the binary name to a string to enable internal path search in db
// The algorithm is PathString = Hex(Name[0]) + ":" + Hex(Name[0]) + ":" + ... + Hex(Name[n])
func DecodeStringToChainName(pathstr string) (name [][]byte) {
	strArray := strings.Split(pathstr, Separator)
	bArray := make([][]byte, 0, 32)
	for _, str := range strArray {
		bytes, _ := DecodeBinary(&str)
		bArray = append(bArray, bytes)
	}
	return bArray
}
