// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package common

import (
    "testing"
    "fmt"
)


func TestEntryMarshal(t *testing.T) {
//type Entry struct {
//    Version     uint8  // 1
//    ChainID     Hash   // 32
//    ExIDSize    uint16 // 2l
//    PayloadSize uint16 // 2 Total of 37 bytes
//    ExtIDs      [][]byte
//    Data        []byte
//}
   
    e := new(Entry)
    e.Version = byte(255)
    e.ChainID.Bytes = make([]byte,32,32)
    for i:=0; i<32; i++ {
       e.ChainID.Bytes[i] = byte(i+1)
    }
    e.ExIDSize = uint16(3)
    e.PayloadSize = uint16(8)
    e.ExtIDs = make([][]byte,1,1)
    e.ExtIDs[0]= []byte{0,1,0xEE}
    e.Data = []byte{0xD1,0xD2,0xD3,0xD4,0xD5}
    
    binary,_ := e.MarshalBinary()
    for _,v := range binary {
      fmt.Printf("%2.2X ",0xFF&v)
    }
    fmt.Println()
}

