package leveldat

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"io"
	"os"
)

// LevelDat implements the encoding and decoding of level.dat files. An empty
// LevelDat is a valid value and may be used to Marshal and Write to a writer or
// file afterward.
type LevelDat struct {
	hdr  header
	data []byte
}

// header holds the header for a level.dat file.
type header struct {
	StorageVersion int32
	FileLength     int32
}

// ReadFile reads a level.dat at a path and returns it.
func ReadFile(name string) (*LevelDat, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("level.dat: open file: %w", err)
	}
	defer f.Close()
	return Read(bufio.NewReader(f))
}

// Read reads a level.dat from r and returns it.
func Read(r io.Reader) (*LevelDat, error) {
	var ldat LevelDat
	if err := binary.Read(r, binary.LittleEndian, &ldat.hdr); err != nil {
		return nil, fmt.Errorf("level.dat: read header: %w", err)
	}
	ldat.data = make([]byte, ldat.hdr.FileLength)
	if n, err := r.Read(ldat.data); err != nil || int32(n) != ldat.hdr.FileLength {
		return nil, fmt.Errorf("level.dat: read data: %w", err)
	}
	return &ldat, nil
}

// Unmarshal decodes the level.dat properties from ld into dst. Unmarshal
// returns an error if dst was unable to store all properties found in the
// level.dat.
func (ld *LevelDat) Unmarshal(dst any) error {
	if err := nbt.UnmarshalEncoding(ld.data, dst, nbt.LittleEndian); err != nil {
		return fmt.Errorf("level.dat: decode nbt: %w", err)
	}
	return nil
}

// Ver returns the version of the level.dat decoded, or 0 if ld is the empty
// value.
func (ld *LevelDat) Ver() int {
	return int(ld.hdr.StorageVersion)
}

// Marshal encodes src and stores it in the level.dat. src should be either a
// struct or a map of fields. Marshal updates the storage version to the latest.
func (ld *LevelDat) Marshal(src any) error {
	var err error
	ld.data, err = nbt.MarshalEncoding(src, nbt.LittleEndian)
	if err != nil {
		return fmt.Errorf("level.dat: encode nbt: %w", err)
	}
	ld.hdr = header{
		StorageVersion: Version,
		FileLength:     int32(len(ld.data)),
	}
	return nil
}

// Write writes ld to w.
func (ld *LevelDat) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, ld.hdr); err != nil {
		return fmt.Errorf("level.dat: write header: %w", err)
	}
	if _, err := w.Write(ld.data); err != nil {
		return fmt.Errorf("level.dat: write data: %w", err)
	}
	return nil
}

// WriteFile writes ld to a file at name.
func (ld *LevelDat) WriteFile(name string) error {
	f, err := os.OpenFile(name, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("level.dat: open file: %w", err)
	}
	w := bufio.NewWriter(f)
	defer func() {
		_ = w.Flush()
		_ = f.Close()
	}()
	return ld.Write(w)
}
