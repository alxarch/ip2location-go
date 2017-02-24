package ip2location

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
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

func NewDirDB(path string, mmap bool) (IP2LocationDB, error) {
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	dbs := MultiDB{}
	for _, entry := range entries {
		p := path + string(os.PathSeparator) + entry.Name()
		if entry.IsDir() {
			if db, err := NewDirDB(p, mmap); err != nil {
				return nil, err
			} else {
				dbs = append(dbs, db)
			}
		}
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".bin") {
			if db, err := NewFileDB(p, mmap); err != nil {
				return nil, err
			} else {
				dbs = append(dbs, db)
			}
		}
	}
	return dbs, nil
}
func NewFileDB(path string, mmap bool) (IP2LocationDB, error) {
	var err error
	s, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		return NewDirDB(path, mmap)
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
