package ip2location

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"strconv"
	"time"
)

type DB struct {
	r       io.ReaderAt
	meta    meta
	offsets map[QueryMode]int64
}

type meta struct {
	dbtype    uint8
	column    uint8
	date      time.Time
	ipv4count uint32
	ipv4addr  uint32
	ipv6count uint32
	ipv6addr  uint32
	ipv4index uint32
	ipv6index uint32
	ipv4col   uint32
	ipv6col   uint32
	ok        bool
}

func (m meta) Date() time.Time {
	return m.date
}

type Record struct {
	CountryCode        string
	CountryName        string
	Region             string
	City               string
	ISP                string
	Latitude           float32
	Longitude          float32
	Domain             string
	Zipcode            string
	Timezone           string
	Netspeed           string
	Iddcode            string
	Areacode           string
	Weatherstationcode string
	Weatherstationname string
	Mcc                string
	Mnc                string
	Mobilebrand        string
	Elevation          float64
	Usagetype          string
}

type dbOffsetMap [25]uint8

var offsetMaps = map[QueryMode]dbOffsetMap{
	countryshort:       dbOffsetMap{0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
	countrylong:        dbOffsetMap{0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
	region:             dbOffsetMap{0, 0, 0, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
	city:               dbOffsetMap{0, 0, 0, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4},
	isp:                dbOffsetMap{0, 0, 3, 0, 5, 0, 7, 5, 7, 0, 8, 0, 9, 0, 9, 0, 9, 0, 9, 7, 9, 0, 9, 7, 9},
	latitude:           dbOffsetMap{0, 0, 0, 0, 0, 5, 5, 0, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5},
	longitude:          dbOffsetMap{0, 0, 0, 0, 0, 6, 6, 0, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6},
	domain:             dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 6, 8, 0, 9, 0, 10, 0, 10, 0, 10, 0, 10, 8, 10, 0, 10, 8, 10},
	zipcode:            dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 7, 7, 7, 7, 0, 7, 7, 7, 0, 7, 0, 7, 7, 7, 0, 7},
	timezone:           dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 8, 7, 8, 8, 8, 7, 8, 0, 8, 8, 8, 0, 8},
	netspeed:           dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 11, 0, 11, 8, 11, 0, 11, 0, 11, 0, 11},
	iddcode:            dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 12, 0, 12, 0, 12, 9, 12, 0, 12},
	areacode:           dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 13, 0, 13, 0, 13, 10, 13, 0, 13},
	weatherstationcode: dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 14, 0, 14, 0, 14, 0, 14},
	weatherstationname: dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 15, 0, 15, 0, 15, 0, 15},
	mcc:                dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 16, 0, 16, 9, 16},
	mnc:                dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 17, 0, 17, 10, 17},
	mobilebrand:        dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 18, 0, 18, 11, 18},
	elevation:          dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 19, 0, 19},
	usagetype:          dbOffsetMap{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 12, 20},
}

var (
	MissingFileError    = errors.New("Invalid database file.")
	NotSupportedError   = errors.New("This parameter is unavailable for selected data file. Please upgrade the data file.")
	InvalidAddressError = errors.New("Invalid IP address.")
)

const api_version string = "8.0.3"

// get api version
func ApiVersion() string {
	return api_version
}

var (
	max_ipv4_range = big.NewInt(4294967295)
	max_ipv6_range = big.NewInt(0)
)

func init() {
	max_ipv6_range.SetString("340282366920938463463374607431768211455", 10)
}

type QueryMode uint32

const (
	countryshort       QueryMode = 0x00001
	countrylong        QueryMode = 0x00002
	region             QueryMode = 0x00004
	city               QueryMode = 0x00008
	isp                QueryMode = 0x00010
	latitude           QueryMode = 0x00020
	longitude          QueryMode = 0x00040
	domain             QueryMode = 0x00080
	zipcode            QueryMode = 0x00100
	timezone           QueryMode = 0x00200
	netspeed           QueryMode = 0x00400
	iddcode            QueryMode = 0x00800
	areacode           QueryMode = 0x01000
	weatherstationcode QueryMode = 0x02000
	weatherstationname QueryMode = 0x04000
	mcc                QueryMode = 0x08000
	mnc                QueryMode = 0x10000
	mobilebrand        QueryMode = 0x20000
	elevation          QueryMode = 0x40000
	usagetype          QueryMode = 0x80000
	all                QueryMode = countryshort | countrylong | region | city | isp | latitude | longitude | domain | zipcode | timezone | netspeed | iddcode | areacode | weatherstationcode | weatherstationname | mcc | mnc | mobilebrand | elevation | usagetype
)

type IPType uint32

const (
	IPv4 IPType = 4
	IPv6 IPType = 6
)

// get IP type and calculate IP number; calculates index too if exists
func (db *DB) CheckIP(ip string) (iptype IPType, ipnum *big.Int, ipindex uint32) {
	iptype = 0
	ipnum = big.NewInt(0)
	tmp := big.NewInt(0)
	ipindex = 0
	if ipaddress := net.ParseIP(ip); ipaddress != nil {
		if v4 := ipaddress.To4(); v4 != nil {
			iptype = IPv4
			ipnum.SetBytes(v4)
			if db.meta.ipv4index > 0 {
				tmp.Rsh(ipnum, 16)
				tmp.Lsh(tmp, 3)
				ipindex = uint32(tmp.Add(tmp, big.NewInt(int64(db.meta.ipv4index))).Uint64())
			}
		} else if v6 := ipaddress.To16(); v6 != nil {
			iptype = IPv6
			ipnum.SetBytes(v6)
			if db.meta.ipv6index > 0 {
				tmp.Rsh(ipnum, 112)
				tmp.Lsh(tmp, 3)
				ipindex = uint32(tmp.Add(tmp, big.NewInt(int64(db.meta.ipv6index))).Uint64())
			}
		}
	}
	return
}

// read byte
func rUint8(r io.ReaderAt, pos int64) (uint8, error) {
	data := make([]byte, 1)
	if _, err := r.ReadAt(data, pos-1); err != nil {
		return 0, err
	}
	return data[0], nil
}
func rPos(r io.ReaderAt, pos int64) (int64, error) {
	if p, err := rUint32(r, pos); err != nil {
		return 0, err
	} else {
		return int64(p), nil
	}
}

// read unsigned 32-bit integer
func rUint32(r io.ReaderAt, pos int64) (uint32, error) {
	data := make([]byte, 4)
	if _, err := r.ReadAt(data, pos-1); err != nil {
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
func rUint128(r io.ReaderAt, pos int64) (*big.Int, error) {
	data := make([]byte, 16)
	if _, err := r.ReadAt(data, pos-1); err != nil {
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
func rStr(r io.ReaderAt, pos int64) (string, error) {
	var s string
	lenbyte := make([]byte, 1)
	if _, err := r.ReadAt(lenbyte, pos); err != nil {
		return s, err
	}
	strlen := lenbyte[0]
	data := make([]byte, strlen)
	if _, err := r.ReadAt(data, pos+1); err != nil {
		return s, err
	}
	return string(data[:strlen]), nil
}

// read float
func rFloat(r io.ReaderAt, pos int64) (float32, error) {
	var f float32
	data := make([]byte, 4)
	if _, err := r.ReadAt(data, pos-1); err != nil {
		return 0.0, err
	}
	buf := bytes.NewReader(data)
	if err := binary.Read(buf, binary.LittleEndian, &f); err != nil {
		return .0, err
	}
	return f, nil
}

// initialize the component with the database path
func OpenDB(dbpath string) (*DB, error) {
	data, err := ioutil.ReadFile(dbpath)
	if err != nil {
		return nil, err
	}
	return NewDB(bytes.NewReader(data))
}

func NewDB(r io.ReaderAt) (*DB, error) {

	db := &DB{r: r}
	var err error
	if db.meta.dbtype, err = rUint8(db.r, 1); err != nil {
		return nil, err
	}
	if db.meta.column, err = rUint8(db.r, 2); err != nil {
		return nil, err
	}
	var y, m, d uint8
	if y, err = rUint8(db.r, 3); err != nil {
		return nil, err
	}
	if m, err = rUint8(db.r, 4); err != nil {
		return nil, err
	}
	if d, err = rUint8(db.r, 5); err != nil {
		return nil, err
	}

	db.meta.date = time.Date(int(y), time.Month(int(m)), int(d), 0, 0, 0, 0, nil)

	if db.meta.ipv4count, err = rUint32(db.r, 6); err != nil {
		return nil, err
	}
	if db.meta.ipv4addr, err = rUint32(db.r, 10); err != nil {
		return nil, err
	}
	if db.meta.ipv6count, err = rUint32(db.r, 14); err != nil {
		return nil, err
	}
	if db.meta.ipv6addr, err = rUint32(db.r, 18); err != nil {
		return nil, err
	}
	if db.meta.ipv4index, err = rUint32(db.r, 22); err != nil {
		return nil, err
	}
	if db.meta.ipv6index, err = rUint32(db.r, 26); err != nil {
		return nil, err
	}
	db.meta.ipv4col = uint32(db.meta.column << 2)              // 4 bytes each column
	db.meta.ipv6col = uint32(16 + ((db.meta.column - 1) << 2)) // 4 bytes each column, except IPFrom column which is 16 bytes

	dbt := db.meta.dbtype

	db.offsets = make(map[QueryMode]int64)
	for m, om := range offsetMaps {
		if pos := om[dbt]; pos != 0 {
			// since both IPv4 and IPv6 use 4 bytes for the below columns, can just do it once here
			db.offsets[m] = int64(uint32(pos-1) << 2)
		}
	}

	db.meta.ok = true
	return db, nil
}

// get all fields
func (db *DB) All(ipaddress string) (Record, error) {
	return db.Query(ipaddress, all)
}

// get country code
func (db *DB) CountryCode(ipaddress string) (Record, error) {
	return db.Query(ipaddress, countryshort)
}

// get country name
func (db *DB) CountryName(ipaddress string) (Record, error) {
	return db.Query(ipaddress, countrylong)
}

// get region
func (db *DB) Region(ipaddress string) (Record, error) {
	return db.Query(ipaddress, region)
}

// get city
func (db *DB) City(ipaddress string) (Record, error) {
	return db.Query(ipaddress, city)
}

// get isp
func (db *DB) ISP(ipaddress string) (Record, error) {
	return db.Query(ipaddress, isp)
}

// get latitude
func (db *DB) Latitude(ipaddress string) (Record, error) {
	return db.Query(ipaddress, latitude)
}

// get longitude
func (db *DB) Longitude(ipaddress string) (Record, error) {
	return db.Query(ipaddress, longitude)
}

// get domain
func (db *DB) Domain(ipaddress string) (Record, error) {
	return db.Query(ipaddress, domain)
}

// get zip code
func (db *DB) ZipCode(ipaddress string) (Record, error) {
	return db.Query(ipaddress, zipcode)
}

// get time zone
func (db *DB) Timezone(ipaddress string) (Record, error) {
	return db.Query(ipaddress, timezone)
}

// get net speed
func (db *DB) Netspeed(ipaddress string) (Record, error) {
	return db.Query(ipaddress, netspeed)
}

// get idd code
func (db *DB) IDDCode(ipaddress string) (Record, error) {
	return db.Query(ipaddress, iddcode)
}

// get area code
func (db *DB) AreaCode(ipaddress string) (Record, error) {
	return db.Query(ipaddress, areacode)
}

// get weather station code
func (db *DB) WeatherStationCode(ipaddress string) (Record, error) {
	return db.Query(ipaddress, weatherstationcode)
}

// get weather station name
func (db *DB) WeatherStationName(ipaddress string) (Record, error) {
	return db.Query(ipaddress, weatherstationname)
}

// get mobile country code
func (db *DB) MCC(ipaddress string) (Record, error) {
	return db.Query(ipaddress, mcc)
}

// get mobile network code
func (db *DB) MNC(ipaddress string) (Record, error) {
	return db.Query(ipaddress, mnc)
}

// get mobile carrier brand
func (db *DB) Mobilebrand(ipaddress string) (Record, error) {
	return db.Query(ipaddress, mobilebrand)
}

// get elevation
func (db *DB) Elevation(ipaddress string) (Record, error) {
	return db.Query(ipaddress, elevation)
}

// get usage type
func (db *DB) UsageType(ipaddress string) (Record, error) {
	return db.Query(ipaddress, usagetype)
}

var bigOne = big.NewInt(1)

// main Query
func (db *DB) Query(ipaddress string, mode QueryMode) (x Record, err error) {
	// read metadata
	if !db.meta.ok {
		return x, MissingFileError
	}

	// check IP type and return IP number & index (if exists)
	iptype, ipno, ipindex := db.CheckIP(ipaddress)

	if iptype == 0 {
		return x, InvalidAddressError
	}

	var colsize, baseaddr, low, high, mid, rowoffset uint32
	var ipfrom, ipto, maxip *big.Int

	switch iptype {
	case IPv4:
		baseaddr = db.meta.ipv4addr
		high = db.meta.ipv4count
		maxip = max_ipv4_range
		colsize = db.meta.ipv4col
	case IPv6:
		baseaddr = db.meta.ipv6addr
		high = db.meta.ipv6count
		maxip = max_ipv6_range
		colsize = db.meta.ipv6col
	default:
		return x, InvalidAddressError
	}

	// reading index
	if ipindex > 0 {
		if low, err = rUint32(db.r, int64(ipindex)); err != nil {
			return
		}
		if high, err = rUint32(db.r, int64(ipindex+4)); err != nil {
			return
		}
	}

	if ipno.Cmp(maxip) >= 0 {
		ipno = ipno.Sub(ipno, bigOne)
	}

	var pos int64
	for low <= high {
		mid = ((low + high) >> 1)
		o1 := int64(baseaddr + (mid * colsize))
		o2 := int64(rowoffset + colsize)

		switch iptype {
		case IPv4:
			if pos, err = rPos(db.r, o1); err != nil {
				return
			}
			ipfrom = big.NewInt(pos)
			if pos, err = rPos(db.r, o2); err != nil {
				return
			}
			ipto = big.NewInt(pos)
		case IPv6:
			if ipfrom, err = rUint128(db.r, o1); err != nil {
				return
			}
			if ipto, err = rUint128(db.r, o2); err != nil {
				return
			}
		default:
			return x, InvalidAddressError
		}

		inrange := ipno.Cmp(ipfrom) >= 0 && ipno.Cmp(ipto) < 0
		if !inrange {
			if ipno.Cmp(ipfrom) < 0 {
				high = mid - 1
			} else {
				low = mid + 1
			}
			continue
		}

		if iptype == IPv6 {
			o1 = o1 + 12 // coz below is assuming all columns are 4 bytes, so got 12 left to go to make 16 bytes total
		}
		for m, mo := range db.offsets {
			if mode&m != 1 {
				continue
			}
			if pos, err = rPos(db.r, o1+mo); err != nil {
				return
			}
			pos += o1

			switch m {
			case countrylong:
				x.CountryName, err = rStr(db.r, pos+3)
			case countryshort:
				x.CountryCode, err = rStr(db.r, pos)
			case region:
				x.Region, err = rStr(db.r, pos)
			case city:
				x.City, err = rStr(db.r, pos)
			case isp:
				x.ISP, err = rStr(db.r, pos)
			case latitude:
				x.Latitude, err = rFloat(db.r, pos)
			case longitude:
				x.Longitude, err = rFloat(db.r, pos)
			case domain:
				x.Domain, err = rStr(db.r, pos)
			case zipcode:
				x.Zipcode, err = rStr(db.r, pos)
			case timezone:
				x.Timezone, err = rStr(db.r, pos)
			case netspeed:
				x.Netspeed, err = rStr(db.r, pos)
			case iddcode:
				x.Iddcode, err = rStr(db.r, pos)
			case areacode:
				x.Areacode, err = rStr(db.r, pos)
			case weatherstationcode:
				x.Weatherstationcode, err = rStr(db.r, pos)
			case weatherstationname:
				x.Weatherstationname, err = rStr(db.r, pos)
			case mcc:
				x.Mcc, err = rStr(db.r, pos)
			case mnc:
				x.Mnc, err = rStr(db.r, pos)
			case mobilebrand:
				x.Mobilebrand, err = rStr(db.r, pos)
			case usagetype:
				x.Usagetype, err = rStr(db.r, pos)
			case elevation:
				var s string
				if s, err = rStr(db.r, pos); err == nil {
					x.Elevation, err = strconv.ParseFloat(s, 32)
				}
			}
			if err != nil {
				return
			}
		}
		return x, nil
	}
	return x, NotSupportedError
}

// for debugging purposes
func (x Record) Print() {
	fmt.Printf("country_short: %s\n", x.CountryCode)
	fmt.Printf("country_long: %s\n", x.CountryName)
	fmt.Printf("region: %s\n", x.Region)
	fmt.Printf("city: %s\n", x.City)
	fmt.Printf("isp: %s\n", x.ISP)
	fmt.Printf("latitude: %f\n", x.Latitude)
	fmt.Printf("longitude: %f\n", x.Longitude)
	fmt.Printf("domain: %s\n", x.Domain)
	fmt.Printf("zipcode: %s\n", x.Zipcode)
	fmt.Printf("timezone: %s\n", x.Timezone)
	fmt.Printf("netspeed: %s\n", x.Netspeed)
	fmt.Printf("iddcode: %s\n", x.Iddcode)
	fmt.Printf("areacode: %s\n", x.Areacode)
	fmt.Printf("weatherstationcode: %s\n", x.Weatherstationcode)
	fmt.Printf("weatherstationname: %s\n", x.Weatherstationname)
	fmt.Printf("mcc: %s\ncheckip", x.Mcc)
	fmt.Printf("mnc: %s\n", x.Mnc)
	fmt.Printf("mobilebrand: %s\n", x.Mobilebrand)
	fmt.Printf("elevation: %f\n", x.Elevation)
	fmt.Printf("usagetype: %s\n", x.Usagetype)
}
