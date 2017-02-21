package ip2location

import (
	"bytes"
	"encoding/binary"
	"io"
	"math/big"
	"sync"
)

var bpool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 16)
	},
}

func blank(size int) []byte {
	if b, ok := bpool.Get().([]byte); ok && size <= cap(b) {
		return b[:size]
	}
	return make([]byte, size)
}

func release(b []byte) {
	bpool.Put(b[:cap(b)])
}

// read byte
func readUint8(r io.ReaderAt, pos int64) (uint8, error) {
	data := blank(1)
	defer release(data)
	if _, err := r.ReadAt(data, pos-1); err != nil {
		return 0, err
	}
	return data[0], nil
}

// read unsigned 32-bit integer
func readUint32(r io.ReaderAt, pos uint32) (uint32, error) {
	data := blank(4)
	defer release(data)
	if _, err := r.ReadAt(data, int64(pos-1)); err != nil {
		return 0, err
	}
	var retval uint32
	buf := bytes.NewReader(data)
	if err := binary.Read(buf, binary.LittleEndian, &retval); err != nil {
		return 0, err
	}
	return retval, nil
}

// read unsigned 128-bit integer
func readUint128(r io.ReaderAt, pos uint32) (*big.Int, error) {
	data := blank(16)
	defer release(data)
	if _, err := r.ReadAt(data, int64(pos-1)); err != nil {
		return nil, err
	}
	retval := big.NewInt(0)

	// little endian to big endian
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
	retval.SetBytes(data)
	return retval, nil
}

// read string
func readString(r io.ReaderAt, pos uint32) (string, error) {
	var s string
	lenbyte := blank(1)
	defer release(lenbyte)
	if _, err := r.ReadAt(lenbyte, int64(pos)); err != nil {
		return s, err
	}
	strlen := lenbyte[0]
	data := make([]byte, strlen)
	if _, err := r.ReadAt(data, int64(pos)+1); err != nil {
		return s, err
	}
	return string(data[:strlen]), nil
}

// read float
func rFloat(r io.ReaderAt, pos uint32) (float32, error) {
	var f float32
	data := make([]byte, 4)
	if _, err := r.ReadAt(data, int64(pos)-1); err != nil {
		return 0.0, err
	}
	buf := bytes.NewReader(data)
	if err := binary.Read(buf, binary.LittleEndian, &f); err != nil {
		return .0, err
	}
	return f, nil
}
