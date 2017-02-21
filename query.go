package ip2location

type QueryMode uint32

const (
	QueryCountryCode        QueryMode = 0x00001
	QueryCountryName        QueryMode = 0x00002
	QueryRegion             QueryMode = 0x00004
	QueryCity               QueryMode = 0x00008
	QueryISP                QueryMode = 0x00010
	QueryLatitude           QueryMode = 0x00020
	QueryLongitude          QueryMode = 0x00040
	QueryDomain             QueryMode = 0x00080
	QueryZipCode            QueryMode = 0x00100
	QueryTimeZone           QueryMode = 0x00200
	QueryNetSpeed           QueryMode = 0x00400
	QueryIDDCode            QueryMode = 0x00800
	QueryAreaCode           QueryMode = 0x01000
	QueryWeatherStationCode QueryMode = 0x02000
	QueryWeatherStationName QueryMode = 0x04000
	QueryMCC                QueryMode = 0x08000
	QueryMNC                QueryMode = 0x10000
	QueryMobileBrand        QueryMode = 0x20000
	QueryElevation          QueryMode = 0x40000
	QueryUsageType          QueryMode = 0x80000
	QueryAll                QueryMode = QueryCountryCode | QueryCountryName | QueryRegion | QueryCity | QueryISP | QueryLatitude | QueryLongitude | QueryDomain | QueryZipCode | QueryTimeZone | QueryNetSpeed | QueryIDDCode | QueryAreaCode | QueryWeatherStationCode | QueryWeatherStationName | QueryMCC | QueryMNC | QueryMobileBrand | QueryElevation | QueryUsageType
)
