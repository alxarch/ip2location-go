package ip2location

import (
	"sync"
)

// PoolDB allows for db pooling
type PoolDB struct {
	pool    *sync.Pool
	Factory func() (IP2LocationDB, error)
}

func (p *PoolDB) Query(ip string, r *Record, m QueryMode) error {
	if p.pool == nil {
		p.pool = &sync.Pool{
			New: func() interface{} {
				if db, err := p.Factory(); err == nil {
					return db
				} else {
					return &errorDB{err}
				}
			},
		}
	}
	db := p.pool.Get().(IP2LocationDB)
	defer p.pool.Put(db)
	return db.Query(ip, r, m)
}
func (p *PoolDB) Close() {

}

type errorDB struct {
	Error error
}

func (e *errorDB) Query(string, *Record, QueryMode) error {
	return e.Error
}
func (e errorDB) Close() {}
