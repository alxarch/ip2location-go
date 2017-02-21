package ip2location

import "errors"

type IPQueryProvider interface {
	Query(string, *Record, QueryMode) error
}

type request struct {
	Mode   QueryMode
	IP     string
	Record *Record
}

type SafeDB struct {
	db IPQueryProvider

	done      chan struct{}
	requests  chan request
	responses chan error
}

func NewSafeDB(db IPQueryProvider) *SafeDB {
	if nil == db {
		return nil
	}
	sd := &SafeDB{
		db:        db,
		requests:  make(chan request, 1000),
		responses: make(chan error, 1000),
		done:      make(chan struct{}),
	}
	go func() {
		defer close(sd.requests)
		defer close(sd.responses)
		for {
			select {
			case <-sd.done:
				return
			case req := <-sd.requests:
				sd.responses <- sd.db.Query(req.IP, req.Record, req.Mode)
			}
		}
	}()
	return sd

}

func (d *SafeDB) Close() {
	if nil != d.done {
		close(d.done)
	}
}

var NotRunningError = errors.New("DB service not running")

func (d *SafeDB) Query(ip string, r *Record, q QueryMode) error {
	if nil == d.requests {
		return NotRunningError
	}
	d.requests <- request{q, ip, r}
	return <-d.responses
}
