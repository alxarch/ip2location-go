package ip2location

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
