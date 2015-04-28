// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	
)

type Hash struct {
	Bytes []byte `json:"bytes"`
}

//Fixed sixe hash used for map, where byte slice won't work
type HashF [HASH_LENGTH]byte

func (h HashF) Hash() Hash {
	return Hash{Bytes: h[:]}
}

func (h *HashF) From(hash *Hash) {
	copy(h[:], hash.Bytes)
}

func NewHash() *Hash {
	h := new(Hash)
	h.Bytes = make([]byte, HASH_LENGTH)
	return h
}

func CreateHash(entities ...BinaryMarshallable) (h *Hash, err error) {
	sha := sha256.New()
	h = new(Hash)
	for _, entity := range entities {
		data, err := entity.MarshalBinary()
		if err != nil {
			return nil, err
		}
		sha.Write(data)
	}
	h.Bytes = sha.Sum(nil)
	return
}

func (h *Hash) MarshalBinary() (bytes [] byte, err error) {
    bytes = make([]byte,HASH_LENGTH,HASH_LENGTH)
    copy(bytes,h.Bytes)
    return bytes, nil
}

func (h *Hash) UnMarshalBinary(data []byte) (hash *Hash, err error) {
    hash = NewHash()
    copy(hash.Bytes,data[:HASH_LENGTH])
    return hash, nil
}
   
func (h *Hash) MarshalledSize() int {
    return HASH_LENGTH
} 

// Unmarshals the Hash, and returns the incremented pointer
func UnmarshalHash(data []byte) (newHash *Hash, newData []byte) {
    newHash = NewHash()
    copy(newHash.Bytes,data[:HASH_LENGTH])
    newData = data[HASH_LENGTH:]
    return newHash,newData
}



// NewShaHash returns a new ShaHash from a byte slice.  An error is returned if
// the number of bytes passed in is not HASH_LENGTH.
func NewShaHash(newHash []byte) (*Hash, error) {
/*********** This makes no sense TODO *****************	
    var sh Hash
	err := sh.SetBytes(newHash)
	if err != nil {
		return nil, err
	}
	return &sh, err
**/
    return Sha(newHash), nil
    
}

func Sha(p []byte) (h *Hash) {
	sha := sha256.New()
	sha.Write(p)

	h = new(Hash)
	h.Bytes = sha.Sum(nil)
	return h
}

func (h *Hash) String() string {
	return hex.EncodeToString(h.Bytes)
}

func (h *Hash) ByteString() string {
	return string(h.Bytes)
}

func HexToHash(hexStr string) (h *Hash, err error) {
	h = new(Hash)
	h.Bytes, err = hex.DecodeString(hexStr)
	return h, err
}


// Compare two Hashes.  If either or both Hash is nil, then they
// are not considered equal. (Not sure about that decision...)
func (a *Hash) IsSameAs(b *Hash) bool {
	if a == nil || b == nil {
		return false
	}

	if bytes.Compare(a.Bytes, b.Bytes) == 0 {
		return true
	}

	return false
}
