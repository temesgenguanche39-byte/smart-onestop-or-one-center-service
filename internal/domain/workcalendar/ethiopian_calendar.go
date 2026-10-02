package workcalendar

import (
	"fmt"
	"time"
)

// EthiopianDate represents a date in the Ethiopian Calendar (Ge'ez Calendar).
// It consists of 12 months of 30 days each, plus a 13th month (Pagume) of 5 days (6 in leap years).
type EthiopianDate struct {
	Year  int `json:"year"`
	Month int `json:"month"` // 1 to 13
	Day   int `json:"day"`   // 1 to 30 (1 to 5 or 6 for Pagume)
}

// Ethiopian JDN Epoch constant:
// Meskerem 1, 1 EE corresponds to Julian Day Number 1,724,221 (August 29, 8 CE Julian).
const ethiopicJDNEpoch int64 = 1724221

// Month names in English and Amharic
var ethiopianMonthNamesEN = [...]string{
	"",
	"Meskerem",
	"Tikimt",
	"Hidar",
	"Tahsas",
	"Tir",
	"Yekatit",
	"Megabit",
	"Miazia",
	"Ginbot",
	"Sene",
	"Hamle",
	"Nehase",
	"Pagume",
}

var ethiopianMonthNamesAM = [...]string{
	"",
	"መስከረም",
	"ጥቅምት",
	"ኅዳር",
	"ታኅሣሥ",
	"ጥር",
	"የካቲት",
	"መጋቢት",
	"ሚያዝያ",
	"ግንቦት",
	"ሰኔ",
	"ሐምሌ",
	"ነሐሴ",
	"ጳጉሜ",
}

// IsEthiopianLeapYear returns true if the given Ethiopian year is a leap year.
// In the Ethiopian calendar, a leap year occurs when year % 4 == 3.
func IsEthiopianLeapYear(year int) bool {
	return (year % 4) == 3
}

// DaysInEthiopianMonth returns the number of days in a given Ethiopian year and month.
func DaysInEthiopianMonth(year, month int) int {
	if month < 1 || month > 13 {
		return 0
	}
	if month <= 12 {
		return 30
	}
	// Month 13 (Pagume)
	if IsEthiopianLeapYear(year) {
		return 6
	}
	return 5
}

// EthiopianMonthName returns the name of the Ethiopian month in the specified language ("am" or "en").
func EthiopianMonthName(month int, lang string) string {
	if month < 1 || month > 13 {
		return ""
	}
	if lang == "am" {
		return ethiopianMonthNamesAM[month]
	}
	return ethiopianMonthNamesEN[month]
}

// EthiopianToJDN converts an Ethiopian calendar date to Julian Day Number.
func EthiopianToJDN(year, month, day int) int64 {
	return ethiopicJDNEpoch + int64(365*(year-1)) + int64(year/4) + int64((month-1)*30) + int64(day-1)
}

// JDNToEthiopian converts a Julian Day Number to an Ethiopian calendar date.
func JDNToEthiopian(jdn int64) (year, month, day int) {
	offset := jdn - ethiopicJDNEpoch

	// Approximate year from 1461-day 4-year cycle (365*4 + 1)
	y := int((4*offset + 1463) / 1461)
	priorDays := int64(365*(y-1) + y/4)
	dayInYear := int(offset - priorDays)

	if dayInYear < 0 {
		y--
		priorDays = int64(365*(y-1) + y/4)
		dayInYear = int(offset - priorDays)
	}

	month = (dayInYear / 30) + 1
	day = (dayInYear % 30) + 1
	year = y
	return year, month, day
}

// GregorianToJDN converts a Gregorian calendar date to Julian Day Number.
func GregorianToJDN(year, month, day int) int64 {
	a := (14 - month) / 12
	y := int64(year + 4800 - a)
	m := int64(month + 12*a - 3)
	return int64(day) + (153*m+2)/5 + 365*y + y/4 - y/100 + y/400 - 32045
}

// JDNToGregorian converts a Julian Day Number to a Gregorian calendar date.
func JDNToGregorian(jdn int64) (year, month, day int) {
	a := jdn + 32044
	b := (4*a + 3) / 146097
	c := a - (146097*b)/4
	d := (4*c + 3) / 1461
	e := c - (1461*d)/4
	m := (5*e + 2) / 153
	day = int(e - (153*m+2)/5 + 1)
	month = int(m + 3 - 12*(m/10))
	year = int(100*b + d - 4800 + m/10)
	return year, month, day
}

// ToEthiopianDate converts a standard Go time.Time to an EthiopianDate.
func ToEthiopianDate(t time.Time) EthiopianDate {
	loc := AddisAbabaLocation()
	tInLoc := t.In(loc)
	jdn := GregorianToJDN(tInLoc.Year(), int(tInLoc.Month()), tInLoc.Day())
	y, m, d := JDNToEthiopian(jdn)
	return EthiopianDate{Year: y, Month: m, Day: d}
}

// ToGregorianDate converts an EthiopianDate to a Go time.Time in the Africa/Addis_Ababa timezone at 00:00:00.
func ToGregorianDate(ed EthiopianDate) time.Time {
	jdn := EthiopianToJDN(ed.Year, ed.Month, ed.Day)
	gy, gm, gd := JDNToGregorian(jdn)
	loc := AddisAbabaLocation()
	return time.Date(gy, time.Month(gm), gd, 0, 0, 0, 0, loc)
}

// FormatEthiopian formats an EthiopianDate as a string in English or Amharic.
// Example: "መስከረም 17, 2017" or "Meskerem 17, 2017".
func FormatEthiopian(ed EthiopianDate, lang string) string {
	mName := EthiopianMonthName(ed.Month, lang)
	return fmt.Sprintf("%s %d, %d", mName, ed.Day, ed.Year)
}

// ToGeezNumeral converts an integer (1 to 9999) to Ge'ez numerals.
func ToGeezNumeral(num int) string {
	if num <= 0 || num > 9999 {
		return fmt.Sprintf("%d", num)
	}

	geezDigits := []string{"", "፩", "፪", "፫", "፬", "፭", "፮", "፯", "፰", "፱"}
	geezTens := []string{"", "፲", "፳", "፴", "፵", "፶", "፷", "፸", "፹", "፺"}
	geezHundreds := "፻"
	geezTenThousands := "፼"

	_ = geezTenThousands

	result := ""

	// Thousands and hundreds
	hundredsPart := num / 100
	remainder := num % 100

	if hundredsPart > 0 {
		if hundredsPart > 1 {
			result += toGeezUnder100(hundredsPart, geezDigits, geezTens)
		}
		result += geezHundreds
	}

	if remainder > 0 {
		result += toGeezUnder100(remainder, geezDigits, geezTens)
	}

	return result
}

func toGeezUnder100(n int, digits, tens []string) string {
	res := ""
	t := n / 10
	d := n % 10
	if t > 0 {
		res += tens[t]
	}
	if d > 0 {
		res += digits[d]
	}
	return res
}
