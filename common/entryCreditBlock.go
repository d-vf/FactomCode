package common

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sync"
)

type CChain struct {
	ChainID *Hash
	Name    [][]byte

	NextBlock       *CBlock
	NextBlockHeight int
	BlockMutex      sync.Mutex
}

type CBlock struct {
	//Marshalized
	Header    *CBlockHeader
	CBEntries []CBEntry //Interface

	//Not Marshalized
	CBHash     *Hash
	MerkleRoot *Hash
	Chain      *CChain
}


//Entry Credit Block Header
type CBlockHeader struct {
    ChainID    *Hash
    BodyHash   *Hash
    PrevKeyMR  *Hash
    PrevHash   *Hash
    DBHeight    int
    SegmentsMR *Hash
    BalanceMR  *Hash
    EntryCount  int
    BodySize    int
}

type CBInfo struct {
	CBHash     *Hash
	FBHash     *Hash
	FBBlockNum uint64
	ChainID    *Hash
}

func CreateCBlock(chain *CChain, prev *CBlock, cap uint) (b *CBlock, err error) {
	if prev == nil && chain.NextBlockHeight != 0 {
		return nil, errors.New("Previous block cannot be nil")
	} else if prev != nil && chain.NextBlockHeight == 0 {
		return nil, errors.New("Origin block cannot have a parent block")
	}

	b = new(CBlock)

	b.Header = new(CBlockHeader)
	b.Header.ChainID = chain.ChainID

	if prev == nil {
		b.Header.PrevKeyMR = NewHash()
		b.Header.PrevHash = NewHash()
		b.Header.BodyHash = NewHash()
	} else {
		if prev.MerkleRoot == nil {
			prev.BuildMerkleRoot()
		}
		b.Header.PrevKeyMR = prev.MerkleRoot

		if prev.CBHash == nil {
			prev.BuildCBHash()
		}
		b.Header.PrevHash = prev.CBHash
	}

	b.Header.DBHeight = chain.NextBlockHeight
	b.Header.SegmentsMR = NewHash()
	b.Header.BalanceMR = NewHash()
	b.Chain = chain
	b.CBEntries = make([]CBEntry, 0, cap)

	return b, err
}

func (b *CBlock) BuildMerkleRoot() (err error) {

	// Create the Entry Block Key Merkle Root from the hash of Header and the Body Merkle Root
	hashes := make([]*Hash, 0, 2)
	binaryEBHeader, _ := b.Header.MarshalBinary()
	hashes = append(hashes, Sha(binaryEBHeader))
	hashes = append(hashes, b.Header.BodyHash)
	merkle := BuildMerkleTreeStore(hashes)
	b.MerkleRoot = merkle[len(merkle)-1] // MerkleRoot is not marshalized in Entry Block

	return
}

func (b *CBlock) BuildCBHash() (err error) {

	binaryEB, _ := b.MarshalBinary()
	b.CBHash = Sha(binaryEB)

	return
}

func (b *CBlock) BuildCBBodyHash() (bodyHash *Hash, err error) {
	var buf bytes.Buffer
	for i := 0; i < len(b.CBEntries); i++ {
		data, _ := b.CBEntries[i].MarshalBinary()
		buf.Write(data)
	}
	bodyHash = Sha(buf.Bytes())

	return bodyHash, nil
}

func (b *CBlock) AddCBEntry(e CBEntry) (err error) {
	b.CBEntries = append(b.CBEntries, e)
	return
}

func (b *CBlock) AddEndOfMinuteMarker(eomType byte) (err error) {

	eOMEntry := &EndOfMinuteEntry{
		entryType: TYPE_MINUTE_NUMBER,
		EOM_Type:  eomType}

	b.AddCBEntry(eOMEntry)

	return
}

func (b *CBlock) AddServerIndexEntry(serverIndex byte) (err error) {

	cbEntry := &ServerIndexEntry{
		entryType:   TYPE_SERVER_INDEX,
		ServerIndex: serverIndex}

	b.AddCBEntry(cbEntry)

	return
}

func (b *CBlock) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	data, _ = b.Header.MarshalBinary()
	buf.Write(data)

	for i := 0; i < b.Header.EntryCount; i++ {
		data, _ := b.CBEntries[i].MarshalBinary()
		buf.Write(data)
	}
	return buf.Bytes(), err
}

func (b *CBlock) MarshalledSize() int {
	var size int = 0

	size += b.Header.MarshalledSize()

	for _, entry := range b.CBEntries {
		size += entry.MarshalledSize()
	}

	return size
}

func (b *CBlock) UnmarshalBinary(data []byte) (err error) {
	h := new(CBlockHeader)
	h.UnmarshalBinary(data)
	b.Header = h

	data = data[h.MarshalledSize():]

	b.CBEntries = make([]CBEntry, b.Header.EntryCount)
	for i := 0; i < b.Header.EntryCount; i++ {
		if data[0] == TYPE_BUY {
			b.CBEntries[i] = new(BuyCBEntry)
		} else if data[0] == TYPE_PAY_CHAIN {
			b.CBEntries[i] = new(PayChainCBEntry)
		} else if data[0] == TYPE_PAY_ENTRY {
			b.CBEntries[i] = new(PayEntryCBEntry)
		} else if data[0] == TYPE_SERVER_INDEX {
			b.CBEntries[i] = new(ServerIndexEntry)
		} else if data[0] == TYPE_MINUTE_NUMBER {
			b.CBEntries[i] = new(EndOfMinuteEntry)
		}
		err = b.CBEntries[i].UnmarshalBinary(data)
		if err != nil {
			return
		}
		data = data[b.CBEntries[i].MarshalledSize():]
	}

	return nil
}



func (b *CBlockHeader) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.Write(b.ChainID.Bytes)
	buf.Write(b.BodyHash.Bytes)
	buf.Write(b.PrevKeyMR.Bytes)
	buf.Write(b.PrevHash.Bytes)

	binary.Write(&buf, binary.BigEndian, b.DBHeight)

	buf.Write(b.SegmentsMR.Bytes)
	buf.Write(b.BalanceMR.Bytes)

	binary.Write(&buf, binary.BigEndian, b.EntryCount)
	binary.Write(&buf, binary.BigEndian, b.BodySize)

	return buf.Bytes(), err
}

func (b *CBlockHeader) MarshalledSize() int {
	var size int = 0

	size += HASH_LENGTH // b.ChainID
	size += HASH_LENGTH // b.BodyHash
	size += HASH_LENGTH // b.PrevKeyMR
	size += HASH_LENGTH // b.PrevHash
	size += 4 // DB Height
	size += HASH_LENGTH // b.SegmentsMR
	size += HASH_LENGTH // b.BalanceMR
	size += 4 // Entry count
	size += 4 // Body Size

	return size
}

func (b *CBlockHeader) UnmarshalBinary(data []byte) (err error) {

	b.ChainID, data = UnmarshalHash(data)

	b.BodyHash,  data = UnmarshalHash(data)
	b.PrevKeyMR, data = UnmarshalHash(data)
	b.PrevHash,  data = UnmarshalHash(data)
	
	b.DBHeight, data = int(binary.BigEndian.Uint32(data[0:4])), data[4:]

	b.SegmentsMR, data = UnmarshalHash(data)
	b.BalanceMR,  data = UnmarshalHash(data)
	
	b.EntryCount, data = int(binary.BigEndian.Uint32(data[0:4])), data[4:]
	b.BodySize,   data = int(binary.BigEndian.Uint32(data[0:4])), data[4:]

	return nil
}

type CBEntry interface {
	Type() byte
	PublicKey() *Hash
	Credits() int
	MarshalBinary() ([]byte, error)
	MarshalledSize() int
	UnmarshalBinary(data []byte) (err error)
}

type BuyCBEntry struct {
	entryType    byte
	publicKey    *Hash
	credits      int
	CBEntry      //interface
	FactomTxHash *Hash
}

type PayEntryCBEntry struct {
	entryType byte
	publicKey *Hash
	credits   int
	CBEntry   //interface
	EntryHash *Hash
	TimeStamp int64
	Sig       []byte
}

type PayChainCBEntry struct {
	entryType        byte
	publicKey        *Hash
	credits          int
	CBEntry          //interface
	EntryHash        *Hash
	ChainIDHash      *Hash
	EntryChainIDHash *Hash //Hash(EntryHash+ChainIDHash)
	Sig              []byte
}

type ServerIndexEntry struct {
	CBEntry     //interface
	entryType   byte
	ServerIndex byte
}

type EndOfMinuteEntry struct {
	CBEntry   //interface
	entryType byte
	EOM_Type  byte
}

type ECBalance struct {
	PublicKey *Hash
	Credits   int
}

func NewPayEntryCBEntry(pubKey *Hash, entryHash *Hash, credits int,
	timeStamp int64, sig []byte) *PayEntryCBEntry {
	e := &PayEntryCBEntry{}
	e.publicKey = pubKey
	e.entryType = TYPE_PAY_ENTRY
	e.credits = credits
	e.EntryHash = entryHash
	e.TimeStamp = timeStamp
	e.Sig = sig

	return e
}

func NewPayChainCBEntry(pubKey *Hash, entryHash *Hash, credits int,
	chainIDHash *Hash, entryChainIDHash *Hash, sig []byte) *PayChainCBEntry {
	e := &PayChainCBEntry{}
	e.publicKey = pubKey
	e.entryType = TYPE_PAY_CHAIN
	e.credits = credits
	e.EntryHash = entryHash
	e.ChainIDHash = chainIDHash
	e.EntryChainIDHash = entryChainIDHash
	e.Sig = sig

	return e
}

func NewBuyCBEntry(pubKey *Hash, factoidTxHash *Hash,
	credits int) *BuyCBEntry {
	e := &BuyCBEntry{}
	e.publicKey = pubKey
	e.entryType = TYPE_BUY
	e.FactomTxHash = factoidTxHash
	e.credits = credits

	return e
}

func (e *BuyCBEntry) Type() byte {
	return e.entryType
}

func (e *BuyCBEntry) PublicKey() *Hash {
	return e.publicKey
}

func (e *BuyCBEntry) Credits() int {
	return e.credits
}

func (e *BuyCBEntry) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.Write([]byte{e.entryType})

	data, err = e.publicKey.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)
	binary.Write(&buf, binary.BigEndian, e.Credits())

	data, err = e.FactomTxHash.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	return buf.Bytes(), nil
}

func (e *BuyCBEntry) MarshalledSize() int {
	var size int = 0
	size += 1                               // Type (byte)
	size += e.publicKey.MarshalledSize()    // PublicKey
	size += 4                               // Credits (int32)
	size += e.FactomTxHash.MarshalledSize() // Factoid Trans Hash

	return size
}

func (e *BuyCBEntry) UnmarshalBinary(data []byte) (err error) {
	e.entryType, data = data[0], data[1:]
	e.publicKey = new(Hash)

	e.publicKey, data = UnmarshalHash(data)

	buf, data := bytes.NewBuffer(data[:4]), data[4:]
	binary.Read(buf, binary.BigEndian, &e.credits)

	e.FactomTxHash = new(Hash)
	e.FactomTxHash, data = UnmarshalHash(data)
	
	return nil
}

func (e *PayEntryCBEntry) Type() byte {
	return e.entryType
}

func (e *PayEntryCBEntry) PublicKey() *Hash {
	return e.publicKey
}

func (e *PayEntryCBEntry) Credits() int {
	return e.credits
}

func (e *PayEntryCBEntry) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.Write([]byte{e.entryType})

	data, err = e.publicKey.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, e.credits)

	data, err = e.EntryHash.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, e.TimeStamp)

	count := len(e.Sig)
	binary.Write(&buf, binary.BigEndian, uint32(count))
	_, err = buf.Write(e.Sig)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (e *PayEntryCBEntry) MarshalledSize() int {
	var size int = 0
	size += 1                            // Type (byte)
	size += e.publicKey.MarshalledSize() // PublicKey
	size += 4                            // Credits (int32)
	size += e.EntryHash.MarshalledSize() // Entry Hash
	size += 8                            //	TimeStamp int64
	size += 4                            // len(e.Sig)
	size += len(e.Sig)                   // sig

	return size
}

func (e *PayEntryCBEntry) UnmarshalBinary(data []byte) (err error) {
	e.entryType, data = data[0], data[1:]

	e.publicKey, data = UnmarshalHash(data)

	buf, data := bytes.NewBuffer(data[:4]), data[4:]
	binary.Read(buf, binary.BigEndian, &e.credits)

	e.EntryHash = new(Hash)
	e.EntryHash, data = UnmarshalHash(data)

	buf = bytes.NewBuffer(data[:8])
	binary.Read(buf, binary.BigEndian, &e.TimeStamp)
	data = data[8:]

	length := binary.BigEndian.Uint32(data[0:4])
	data = data[4:]
	e.Sig = data[:length]

	return nil
}

func (e *PayChainCBEntry) Type() byte {
	return e.entryType
}

func (e *PayChainCBEntry) PublicKey() *Hash {
	return e.publicKey
}

func (e *PayChainCBEntry) Credits() int {
	return e.credits
}

func (e *PayChainCBEntry) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.Write([]byte{e.entryType})

	data, err = e.publicKey.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, e.credits)

	data, err = e.EntryHash.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	data, err = e.ChainIDHash.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	data, err = e.EntryChainIDHash.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	count := len(e.Sig)
	binary.Write(&buf, binary.BigEndian, uint32(count))

	_, err = buf.Write(e.Sig)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (e *PayChainCBEntry) MarshalledSize() int {
	var size int = 0
	size += 1                                   // Type (byte)
	size += e.publicKey.MarshalledSize()        // PublicKey
	size += 4                                   // Credits (int32)
	size += e.EntryHash.MarshalledSize()        // Entry Hash
	size += e.ChainIDHash.MarshalledSize()      // ChainID Hash
	size += e.EntryChainIDHash.MarshalledSize() // EntryChainID Hash
	size += 4                                   // len(e.Sig)
	size += len(e.Sig)                          // sig

	return size
}

func (e *PayChainCBEntry) UnmarshalBinary(data []byte) (err error) {
	e.entryType, data = data[0], data[1:]

	e.publicKey, data = UnmarshalHash(data)

	buf, data := bytes.NewBuffer(data[:4]), data[4:]
	binary.Read(buf, binary.BigEndian, &e.credits)

	e.EntryHash,        data = UnmarshalHash(data)
	e.ChainIDHash,      data = UnmarshalHash(data)
	e.EntryChainIDHash, data = UnmarshalHash(data)

	length := binary.BigEndian.Uint32(data[0:4])
	data = data[4:]
	e.Sig = data[:length]

	return nil
}

// ServerIndexEntry
func (e *ServerIndexEntry) Type() byte {
	return e.entryType
}

func (e *ServerIndexEntry) PublicKey() *Hash {
	return nil
}

func (e *ServerIndexEntry) Credits() int {
	return 0
}

func (e *ServerIndexEntry) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.Write([]byte{e.entryType})

	buf.Write([]byte{e.ServerIndex})

	return buf.Bytes(), nil
}

func (e *ServerIndexEntry) MarshalledSize() int {
	var size int = 0
	size += 1 // Type (byte)
	size += 1 // ServerIndex (byte)

	return size
}

func (e *ServerIndexEntry) UnmarshalBinary(data []byte) (err error) {
	e.entryType, data = data[0], data[1:]
	e.ServerIndex, data = data[0], data[1:]

	return nil
}
 
// EndOfMinuteEntry
func (e *EndOfMinuteEntry) Type() byte {
	return e.entryType
}

func (e *EndOfMinuteEntry) PublicKey() *Hash {
	return nil
}

func (e *EndOfMinuteEntry) Credits() int {
	return 0
}

func (e *EndOfMinuteEntry) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.Write([]byte{e.entryType})

	buf.Write([]byte{e.EOM_Type})

	return buf.Bytes(), nil
}

func (e *EndOfMinuteEntry) MarshalledSize() int {
	var size int = 0
	size += 1 // Type (byte)
	size += 1 // EOM_Type (byte)

	return size
}

func (e *EndOfMinuteEntry) UnmarshalBinary(data []byte) (err error) {
	e.entryType, data = data[0], data[1:]
	e.EOM_Type, data = data[0], data[1:]

	return nil
}
