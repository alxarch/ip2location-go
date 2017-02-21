package ip2location

import "fmt"

type Record struct {
	CountryCode        string
	CountryName        string
	Region             string
	City               string
	ISP                string
	Latitude           float32
	Longitude          float32
	Domain             string
	ZipCode            string
	Timezone           string
	NetSpeed           string
	IDDCode            string
	Areacode           string
	WeatherStationCode string
	WeatherStationName string
	MCC                string
	MNC                string
	MobileBrand        string
	Elevation          float64
	UsageType          string
}

// func (x Record) String() string {
// 	return fmt.Sprintf("%v", []string{
// 		"CountryCode", x.CountryCode,
// 		"CountryName", x.CountryName,
// 		"Region", x.Region,
// 		"City", x.City,
// 		"ISP", x.ISP,
// 		"Latitude", fmt.Sprint(x.Latitude),
// 		"Longitude", fmt.Sprint(x.Longitude),
// 		"Domain", x.Domain,
// 		"ZipCode", x.ZipCode,
// 		"Timezone", x.Timezone,
// 		"NetSpeed", x.NetSpeed,
// 		"IDDCode", x.IDDCode,
// 		"Areacode", x.Areacode,
// 		"WeatherStationCode", x.WeatherStationCode,
// 		"WeatherStationName", x.WeatherStationName,
// 		"MCC", x.MCC,
// 		"MNC", x.MNC,
// 		"MobileBrand", x.MobileBrand,
// 		"Elevation", fmt.Sprint(x.Elevation),
// 		"UsageType", x.UsageType,
// 	})
// }

// for debugging purposes
func (x Record) Print() {
	fmt.Printf("country_short: %s\n", x.CountryCode)
	fmt.Printf("country_long: %s\n", x.CountryName)
	fmt.Printf("QueryRegion: %s\n", x.Region)
	fmt.Printf("QueryCity: %s\n", x.City)
	fmt.Printf("QueryISP: %s\n", x.ISP)
	fmt.Printf("QueryLatitude: %f\n", x.Latitude)
	fmt.Printf("QueryLongitude: %f\n", x.Longitude)
	fmt.Printf("QueryDomain: %s\n", x.Domain)
	fmt.Printf("QueryZipCodeQueryTimeZone %s\n", x.ZipCode)
	fmt.Printf("timezone: %s\n", x.Timezone)
	fmt.Printf("QueryNetSpeed: %s\n", x.NetSpeed)
	fmt.Printf("QueryIDDCode: %s\n", x.IDDCode)
	fmt.Printf("QueryAreaCode: %s\n", x.Areacode)
	fmt.Printf("QueryWeatherStationCode: %s\n", x.WeatherStationCode)
	fmt.Printf("QueryWeatherStationName: %s\n", x.WeatherStationName)
	fmt.Printf("QueryMCC: %s\ncheckip", x.MCC)
	fmt.Printf("QueryMNC: %s\n", x.MNC)
	fmt.Printf("QueryMobileBrand: %s\n", x.MobileBrand)
	fmt.Printf("QueryElevation: %f\n", x.Elevation)
	fmt.Printf("QueryUsageType: %s\n", x.UsageType)
}
