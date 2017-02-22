package ip2location_test

import (
	"log"
	"os"
	"testing"

	ip2loc "github.com/alxarch/ip2location-go"
	orig "github.com/ip2location/ip2location-go"
)

var binfile string = "data/db.bin"
var dbfile *os.File

func init() {

	if p, ok := os.LookupEnv("IP2L_BINFILE"); ok {
		binfile = p
	}

	f, err := os.Open(binfile)
	if err != nil {
		log.Fatal(err)
	}
	dbfile = f
	orig.Open(binfile)

}

func Test_Meta(t *testing.T) {
	m := &ip2loc.DBMeta{}
	if err := m.Read(dbfile); err != nil {
		t.Error("Failed to init meta %s", err)
	}

	// log.Printf("%v", m)
}
func Test_NewDB(t *testing.T) {
	db, err := ip2loc.NewDB(dbfile)
	if err != nil {
		t.Error("Failed to init db %s", err)
	}
	if db == nil {
		t.Error("Failed to init db %s", err)

	}
	ip, ipt := ip2loc.ParseIP("127.0.0.1")
	if ipt != ip2loc.IPv4 {
		t.Error("Invalid ip type")
	}
	if ip.Int64() <= 0 {
		t.Error("Invalid ip number")
	}

}
