package ip2location

import (
	"bytes"
	"io"
	"os"
	"syscall"
)

type FileDB struct {
	f  *os.File
	db *DB
}

func (fd *FileDB) Query(ip string, r *Record, mode QueryMode) error {
	return fd.db.Query(ip, r, mode)
}
func (fdb *FileDB) Close() {
	if nil != fdb.f {
		fdb.f.Close()
	}
	if nil != fdb.db {
		fdb.db.Close()
	}

}

func NewFileDB(path string, mmap bool) (IP2LocationDB, error) {
	var err error
	s, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	db := &FileDB{}
	var r io.ReaderAt
	if mmap {
		var fd int
		if fd, err = syscall.Open(path, syscall.O_RDONLY, 0); err != nil {
			return nil, err
		}
		var data []byte
		if data, err = syscall.Mmap(fd, 0, int(s.Size()), syscall.PROT_READ, syscall.MAP_SHARED); err != nil {
			return nil, err
		}
		r = bytes.NewReader(data)
	} else {
		if db.f, err = os.Open(path); err != nil {
			return nil, err
		}
		r = db.f
	}
	if r != nil {
		if db.db, err = NewDB(r); err != nil {
			return nil, err
		}
	}

	return db, nil

}
