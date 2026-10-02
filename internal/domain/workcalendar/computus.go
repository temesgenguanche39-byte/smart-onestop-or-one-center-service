package workcalendar

import "time"

// OrthodoxEaster computes the date of Orthodox Easter (Fasika) for a given Gregorian year
// using the canonical Meeus/Jones/Butcher Julian Easter computus converted to Gregorian date.
func OrthodoxEaster(gregorianYear int) time.Time {
	y := gregorianYear
	a := y % 4
	b := y % 7
	c := y % 19
	d := (19*c + 15) % 30
	e := (2*a + 4*b - d + 34) % 7

	julianMonth := (d + e + 114) / 31
	julianDay := ((d + e + 114) % 31) + 1

	// Calculate Julian Day Number from the Julian calendar date (julianYear, julianMonth, julianDay)
	jy := y
	jm := julianMonth
	jd := julianDay

	if jm <= 2 {
		jy--
		jm += 12
	}

	// Meeus JDN for Julian Calendar
	jdn := int64(365.25*float64(jy+4716)) + int64(30.6001*float64(jm+1)) + int64(jd) - 1524

	// Convert JDN to Gregorian date
	gy, gm, gd := JDNToGregorian(jdn)
	loc := AddisAbabaLocation()
	return time.Date(gy, time.Month(gm), gd, 0, 0, 0, 0, loc)
}

// OrthodoxGoodFriday computes Good Friday (Siklet) for a given Gregorian year (2 days before Easter).
func OrthodoxGoodFriday(gregorianYear int) time.Time {
	easter := OrthodoxEaster(gregorianYear)
	return easter.AddDate(0, 0, -2)
}
