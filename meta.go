package ip2location

import (
	"io"
	"math/big"
	"time"
)

var (
	max_ipv4_range = big.NewInt(4294967295)
	max_ipv6_range = big.NewInt(0)
)

func init() {
	max_ipv6_range.SetString("340282366920938463463374607431768211455", 10)
}

type DBMeta struct {
	dbtype      DBType
	colsize     uint8
	date        time.Time
	ipv4count   uint32
	ipv4addr    uint32
	ipv6count   uint32
	ipv6addr    uint32
	ipv4index   uint32
	ipv6index   uint32
	ipv4colsize uint32
	ipv6colsize uint32
	ipv4bigidx  *big.Int
	ipv6bigidx  *big.Int
}

func (m *DBMeta) Type() DBType {
	return m.dbtype
}
func (m *DBMeta) HasIndex(t IPType) bool {
	switch t {
	case IPv4:
		return m.ipv4index > 0
	case IPv6:
		return m.ipv4index > 0
	default:
		return false
	}
}
func (m *DBMeta) Indexes(t IPType) (start, end, colsize uint32, max *big.Int) {
	switch t {
	case IPv4:
		start = m.ipv4addr
		end = m.ipv4count
		max = max_ipv4_range
		colsize = m.ipv4colsize
	case IPv6:
		start = m.ipv6addr
		end = m.ipv6count
		max = max_ipv6_range
		colsize = m.ipv6colsize
	}
	return
}
func (m *DBMeta) Has(t IPType) bool {
	switch t {
	case IPv4:
		return m.ipv4count > 0
	case IPv6:
		return m.ipv6count > 0
	default:
		return false
	}

}
func (m *DBMeta) Date() time.Time {
	return m.date
}

func readDbType(r io.ReaderAt) (DBType, error) {
	t, err := readUint8(r, 1)
	return DBType(t), err
}
func readDbDate(r io.ReaderAt) (t time.Time, err error) {
	var y, m, d uint8
	if y, err = readUint8(r, 3); err != nil {
		return
	}
	if m, err = readUint8(r, 4); err != nil {
		return
	}
	if d, err = readUint8(r, 5); err != nil {
		return
	}
	t = time.Date(int(y)+2000, time.Month(int(m)), int(d), 0, 0, 0, 0, time.UTC)
	return
}
func (m *DBMeta) Read(r io.ReaderAt) (err error) {
	if m.dbtype, err = readDbType(r); err != nil {
		return
	}
	if m.date, err = readDbDate(r); err != nil {
		return
	}

	if m.colsize, err = readUint8(r, 2); err != nil {
	}
	if m.ipv4count, err = readUint32(r, 6); err != nil {
		return
	}
	if m.ipv4addr, err = readUint32(r, 10); err != nil {
		return
	}
	if m.ipv6count, err = readUint32(r, 14); err != nil {
		return
	}
	if m.ipv6addr, err = readUint32(r, 18); err != nil {
		return
	}
	if m.ipv4index, err = readUint32(r, 22); err != nil {
		return
	}
	if m.ipv6index, err = readUint32(r, 26); err != nil {
		return
	}
	m.ipv4bigidx = big.NewInt(int64(m.ipv4index))
	m.ipv6bigidx = big.NewInt(int64(m.ipv6index))
	m.ipv4colsize = uint32(m.colsize * 4)               // 4 bytes each column
	m.ipv6colsize = uint32(16 + ((m.colsize - 1) << 2)) // 4 bytes each column, except IPFrom column which is 16 bytes

	return nil
}
