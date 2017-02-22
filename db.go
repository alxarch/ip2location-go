package ip2location

import (
	"errors"
	"io"
	"math/big"
	"net"
	"os"
	"strconv"
)

type DBType uint8

const (
	DB1 DBType = iota + 1
	DB2
	DB3
	DB4
	DB5
	DB6
	DB7
	DB8
	DB9
	DB10
	DB11
	DB12
	DB13
	DB14
	DB15
	DB16
	DB17
	DB18
	DB19
	DB20
	DB21
	DB22
	DB23
	DB24
	maxdb
)

type DB struct {
	r       io.ReaderAt
	meta    DBMeta
	offsets map[QueryMode]uint32
	mode    QueryMode
}

type dbOffsetMap [25]uint8

var offsetMaps = map[QueryMode]dbOffsetMap{
	QueryCountryCode:        dbOffsetMap{0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
	QueryCountryName:        dbOffsetMap{0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
	QueryRegion:             dbOffsetMap{0, 0, 0, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
	QueryCity:               dbOffsetMap{0, 0, 0, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4},
	QueryISP:                dbOffsetMap{0, 0, 3, 0, 5, 0, 7, 5, 7, 0, 8, 0, 9, 0, 9, 0, 9, 0, 9, 7, 9, 0, 9, 7, 9},
	QueryLatitude:           dbOffsetMap{0, 0, 0, 0, 0, 5, 5, 0, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5},
	QueryLongitude:          dbOffsetMap{0, 0, 0, 0, 0, 6, 6, 0, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6},
	QueryDomain:             dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 6, 8, 0, 9, 0, 10, 0, 10, 0, 10, 0, 10, 8, 10, 0, 10, 8, 10},
	QueryZipCode:            dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 7, 7, 7, 7, 0, 7, 7, 7, 0, 7, 0, 7, 7, 7, 0, 7},
	QueryTimeZone:           dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 8, 7, 8, 8, 8, 7, 8, 0, 8, 8, 8, 0, 8},
	QueryNetSpeed:           dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 11, 0, 11, 8, 11, 0, 11, 0, 11, 0, 11},
	QueryIDDCode:            dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 12, 0, 12, 0, 12, 9, 12, 0, 12},
	QueryAreaCode:           dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 13, 0, 13, 0, 13, 10, 13, 0, 13},
	QueryWeatherStationCode: dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 14, 0, 14, 0, 14, 0, 14},
	QueryWeatherStationName: dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 15, 0, 15, 0, 15, 0, 15},
	QueryMCC:                dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 16, 0, 16, 9, 16},
	QueryMNC:                dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 17, 0, 17, 10, 17},
	QueryMobileBrand:        dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 18, 0, 18, 11, 18},
	QueryElevation:          dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 19, 0, 19},
	QueryUsageType:          dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 12, 20},
}

var (
	MissingFileError            = errors.New("Invalid database file.")
	NotSupportedError           = errors.New("This parameter is unavailable for selected data file. Please upgrade the data file.")
	InvalidAddressError         = errors.New("Invalid IP address.")
	UnsupportedAddressTypeError = errors.New("Unsupported IP address type.")
	NoMatchError                = errors.New("No matching IP range found.")
)

func NewDB(r io.ReaderAt) (db *DB, err error) {
	db = &DB{r: r}
	if err = db.meta.Read(r); err != nil {
		return
	}
	dbt := db.meta.dbtype

	db.offsets = make(map[QueryMode]uint32)
	for m, ofm := range offsetMaps {
		if pos := ofm[dbt]; pos != 0 {
			// since both IPv4 and IPv6 use 4 bytes for the below columns, can just do it once here
			db.offsets[m] = uint32(pos-1) << 2
			db.mode |= m
		}
	}

	return db, nil
}

func (db *DB) Close() {}

func (db *DB) Index(ip *big.Int, t IPType) uint32 {
	switch t {
	case IPv4:
		tmp := big.NewInt(0)
		tmp.Rsh(ip, 16)
		tmp.Lsh(tmp, 3)
		return uint32(tmp.Add(tmp, db.meta.ipv4bigidx).Uint64())
	case IPv6:
		tmp := big.NewInt(0)
		tmp.Rsh(ip, 112)
		tmp.Lsh(tmp, 3)
		return uint32(tmp.Add(tmp, db.meta.ipv6bigidx).Uint64())
	}
	return 0
}

// get IP type and calculate IP number; calculates index too if exists
func (db *DB) CheckIP(ips string) (ip *big.Int, ipt IPType, index uint32) {
	ip = big.NewInt(0)
	if a := net.ParseIP(ips); a != nil {
		if v4 := a.To4(); v4 != nil {
			ipt = IPv4
			ip.SetBytes(v4)
		} else if v6 := a.To16(); v6 != nil {
			ipt = IPv6
			ip.SetBytes(v6)
		}
		index = db.Index(ip, ipt)
	}
	return
}

// main Query
func (db *DB) Query(ipaddress string, x *Record, mode QueryMode) (err error) {
	// check IP type and return IP number & index (if exists)
	ip, t, index := db.CheckIP(ipaddress)
	return db.query(ip, t, index, x, mode)
}
func (db *DB) query(ip *big.Int, ipt IPType, index uint32, x *Record, mode QueryMode) (err error) {
	if mode&db.mode == 0 {
		return NotSupportedError
	}
	if !db.meta.Has(ipt) {
		return UnsupportedAddressTypeError
	}
	base, high, colsize, maxip := db.meta.Indexes(ipt)
	var low, mid uint32
	var ipfrom, ipto *big.Int

	// reading index
	if index > 0 {
		if low, err = readUint32(db.r, uint32(index)); err != nil {
			return
		}
		if high, err = readUint32(db.r, index+4); err != nil {
			return
		}
	}

	if ip.Cmp(maxip) >= 0 {
		ip = ip.Sub(ip, bigOne)
	}

	var pos uint32
	cs := uint32(colsize)
	for low <= high {
		mid = ((low + high) >> 1) // (low + high) / 2
		o1 := uint32(base + (mid * colsize))
		o2 := o1 + cs

		switch ipt {
		case IPv4:
			var ipn uint32
			if ipn, err = readUint32(db.r, o1); err != nil {
				return
			} else {
				ipfrom = big.NewInt(int64(ipn))
			}
			if ipn, err = readUint32(db.r, o2); err != nil {
				return
			} else {
				ipto = big.NewInt(int64(ipn))
			}
		case IPv6:
			if ipfrom, err = readUint128(db.r, o1); err != nil {
				return
			}
			if ipto, err = readUint128(db.r, o2); err != nil {
				return
			}
		default:
			return InvalidAddressError
		}

		inrange := ip.Cmp(ipfrom) >= 0 && ip.Cmp(ipto) < 0
		if !inrange {
			if ip.Cmp(ipfrom) < 0 {
				high = mid - 1
			} else {
				low = mid + 1
			}
			continue
		}

		if ipt == IPv6 {
			o1 += 12 // coz below is assuming all columns are 4 bytes, so got 12 left to go to make 16 bytes total
		}

		for m, mo := range db.offsets {
			if mode&m == 0 {
				// Query is not intereseted in mode
				continue
			}
			if pos, err = readUint32(db.r, o1+mo); err != nil {
				return
			}

			switch m {
			case QueryCountryName:
				x.CountryName, err = readString(db.r, pos+3)
			case QueryCountryCode:
				x.CountryCode, err = readString(db.r, pos)
			case QueryRegion:
				x.Region, err = readString(db.r, pos)
			case QueryCity:
				x.City, err = readString(db.r, pos)
			case QueryISP:
				x.ISP, err = readString(db.r, pos)
			case QueryLatitude:
				x.Latitude, err = rFloat(db.r, pos)
			case QueryLongitude:
				x.Longitude, err = rFloat(db.r, pos)
			case QueryDomain:
				x.Domain, err = readString(db.r, pos)
			case QueryZipCode:
				x.ZipCode, err = readString(db.r, pos)
			case QueryTimeZone:
				x.Timezone, err = readString(db.r, pos)
			case QueryNetSpeed:
				x.NetSpeed, err = readString(db.r, pos)
			case QueryIDDCode:
				x.IDDCode, err = readString(db.r, pos)
			case QueryAreaCode:
				x.Areacode, err = readString(db.r, pos)
			case QueryWeatherStationCode:
				x.WeatherStationCode, err = readString(db.r, pos)
			case QueryWeatherStationName:
				x.WeatherStationName, err = readString(db.r, pos)
			case QueryMCC:
				x.MCC, err = readString(db.r, pos)
			case QueryMNC:
				x.MNC, err = readString(db.r, pos)
			case QueryMobileBrand:
				x.MobileBrand, err = readString(db.r, pos)
			case QueryUsageType:
				x.UsageType, err = readString(db.r, pos)
			case QueryElevation:
				var s string
				if s, err = readString(db.r, pos); err == nil {
					x.Elevation, err = strconv.ParseFloat(s, 32)
				}
			}
			if err != nil {
				return
			}
		}
		return nil
	}
	return NoMatchError
}

type IP2LocationDB interface {
	Query(string, *Record, QueryMode) error
	Close()
}

type MultiDB []IP2LocationDB

func (md MultiDB) Close() {
	for _, db := range md {
		db.Close()
	}
}

func (md MultiDB) Query(ip string, r *Record, mode QueryMode) error {
	matches := 0
	var lasterr error
	for _, db := range md {
		if err := db.Query(ip, r, mode); err != nil {
			switch err {
			case NotSupportedError, UnsupportedAddressTypeError, NoMatchError:
				lasterr = err
			default:
				return err
			}
		} else {
			matches++

		}
	}
	if matches == 0 {
		return lasterr
	}
	return nil
}

type FileDB struct {
	f  *os.File
	db *DB
}

func NewFileDB(path string) (IP2LocationDB, error) {
	if f, err := os.Open(path); err != nil {
		return nil, err
	} else {
		return &DB{r: f}, nil
	}

}

func (fdb *FileDB) Close() {
	if fdb.f != nil {
		fdb.f.Close()
	}
}
