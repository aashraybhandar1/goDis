package log

import (
	"os"

	"github.com/tysonmote/gommap"
)

var (
	offWidth  uint64 = 4
	posWidth  uint64 = 8
	entWiddth        = offWidth + posWidth
)

type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}
