package util

import (
	"github.com/FactomProject/FactomCode/common"
)

//------------------------------------------------
// DBlock array sorting implementation - accending
type ByDBlockIDAccending []common.DirectoryBlock

func (f ByDBlockIDAccending) Len() int {
	return len(f)
}
func (f ByDBlockIDAccending) Less(i, j int) bool {
	return f[i].Header.BlockHeight < f[j].Header.BlockHeight
}
func (f ByDBlockIDAccending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

//------------------------------------------------
// CBlock array sorting implementation - accending
type ByCBlockIDAccending []common.CBlock

func (f ByCBlockIDAccending) Len() int {
	return len(f)
}
func (f ByCBlockIDAccending) Less(i, j int) bool {
	return f[i].Header.DBHeight < f[j].Header.DBHeight
}
func (f ByCBlockIDAccending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

//------------------------------------------------
// EBlock array sorting implementation - accending
type ByEBlockIDAccending []common.EBlock

func (f ByEBlockIDAccending) Len() int {
	return len(f)
}
func (f ByEBlockIDAccending) Less(i, j int) bool {
	return f[i].Header.EBHeight < f[j].Header.EBHeight
}
func (f ByEBlockIDAccending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
