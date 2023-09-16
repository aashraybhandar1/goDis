package log

import (
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

// Given a file with index data. That is the file which maps offset to the position in the log file
func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}
	fi, err := os.Stat(f.Name())

	if err != nil {
		return nil, err
	}

	// This indicates teh current size of the file where indexes are stored
	idx.size = uint64(fi.Size())

	//This method is technically used to increase the file size to the MaxIndexBytes
	if err = os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}

	//Then the file is memory mapped since writing to disk always is expensive
	if idx.mmap, err = gommap.Map(idx.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED); err != nil {
		return nil, err
	}
	return idx, nil
}

func (i *index) Close() error {
	//No sure what this does but some sync related to the mmap file? Im guessing waits for all writes or whatever
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}
	//Ensures stuff from memory map is flushed to the file. Which is the persistent storage in this case
	if err := i.file.Sync(); err != nil {
		return err
	}
	//Truncate, in this case shrinks the file to the actual size. Im guessing coz if you call newIndex with this file the size should be accurate
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}
	return i.file.Close()
}

// in is the relative index we want to read from the start of the file
func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}
	if in == -1 {
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}
	pos = uint64(out) * entWidth
	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}
	//Reads the first 4 bytes for offset
	out = enc.Uint32(i.mmap[pos : pos+offWidth])
	// Reads next 8 bytes for the position
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
	return out, pos, nil
}

func (i *index) Write(off uint32, pos uint64) error {
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF
	}
	//Similarly like previous method appends stuff at the end of the file. For given offset what position the data(log) is present at
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
	i.size += uint64(entWidth)
	return nil
}

func (i *index) Name() string {
	return i.file.Name()
}
