package datetime

import (
	"fmt"
	"time"
)

func ExampleGetNowDate() {
	result := GetNowDate()

	expected := time.Now().Format("2006-01-02")

	fmt.Println(result == expected)

	// Output:
	// true
}

func ExampleGetNowTime() {
	result := GetNowTime()

	expected := time.Now().Format("15:04:05")

	fmt.Println(result == expected)

	// Output:
	// true
}

func ExampleGetNowDateTime() {
	result := GetNowDateTime()

	expected := time.Now().Format("2006-01-02 15:04:05")

	fmt.Println(result == expected)

	// Output:
	// true
}

func ExampleFormatTimeToStr() {
	datetime, _ := time.Parse("2006-01-02 15:04:05", "2021-01-02 16:04:08")

	result1 := FormatTimeToStr(datetime, "yyyy-mm-dd hh:mm:ss")
	result2 := FormatTimeToStr(datetime, "yyyy-mm-dd")
	result3 := FormatTimeToStr(datetime, "dd-mm-yy hh:mm:ss")
	result4 := FormatTimeToStr(datetime, "yyyy-mm-dd hh")

	fmt.Println(result1)
	fmt.Println(result2)
	fmt.Println(result3)
	fmt.Println(result4)

	// Output:
	// 2021-01-02 16:04:08
	// 2021-01-02
	// 02-01-21 16:04:08
	// 2021-01-02 16
}

func ExampleAddDay() {
	date, _ := time.Parse("2006-01-02 15:04:05", "2021-01-01 00:00:00")

	after1Day := AddDay(date, 1)
	before1Day := AddDay(date, -1)

	fmt.Println(after1Day.Format("2006-01-02 15:04:05"))
	fmt.Println(before1Day.Format("2006-01-02 15:04:05"))

	// Output:
	// 2021-01-02 00:00:00
	// 2020-12-31 00:00:00
}

func ExampleAddWeek() {
	date, _ := time.Parse("2006-01-02", "2021-01-01")

	after2Weeks := AddWeek(date, 2)
	before2Weeks := AddWeek(date, -2)

	fmt.Println(after2Weeks.Format("2006-01-02"))
	fmt.Println(before2Weeks.Format("2006-01-02"))

	// Output:
	// 2021-01-15
	// 2020-12-18
}

func ExampleAddMonth() {
	date, _ := time.Parse("2006-01-02", "2021-01-01")

	after2Months := AddMonth(date, 2)
	before2Months := AddMonth(date, -2)

	fmt.Println(after2Months.Format("2006-01-02"))
	fmt.Println(before2Months.Format("2006-01-02"))

	// Output:
	// 2021-03-01
	// 2020-11-01
}

func ExampleAddHour() {
	date, _ := time.Parse("2006-01-02 15:04:05", "2021-01-01 00:00:00")

	after2Hours := AddHour(date, 2)
	before2Hours := AddHour(date, -2)

	fmt.Println(after2Hours.Format("2006-01-02 15:04:05"))
	fmt.Println(before2Hours.Format("2006-01-02 15:04:05"))

	// Output:
	// 2021-01-01 02:00:00
	// 2020-12-31 22:00:00
}

func ExampleAddMinute() {
	date, _ := time.Parse("2006-01-02 15:04:05", "2021-01-01 00:00:00")

	after2Minutes := AddMinute(date, 2)
	before2Minutes := AddMinute(date, -2)

	fmt.Println(after2Minutes.Format("2006-01-02 15:04:05"))
	fmt.Println(before2Minutes.Format("2006-01-02 15:04:05"))

	// Output:
	// 2021-01-01 00:02:00
	// 2020-12-31 23:58:00
}

func ExampleAddDuration() {
	date, _ := time.Parse("2006-01-02 15:04:05", "2021-01-01 00:00:00")

	after2Minutes := AddDuration(date, 2*time.Minute)
	before2Minutes := AddDuration(date, -2*time.Minute)

	fmt.Println(after2Minutes.Format("2006-01-02 15:04:05"))
	fmt.Println(before2Minutes.Format("2006-01-02 15:04:05"))

	// Output:
	// 2021-01-01 00:02:00
	// 2020-12-31 23:58:00
}

func ExampleAddYear() {
	date, _ := time.Parse("2006-01-02", "2021-01-01")

	after2Years := AddYear(date, 2)
	before2Years := AddYear(date, -2)

	fmt.Println(after2Years.Format("2006-01-02"))
	fmt.Println(before2Years.Format("2006-01-02"))

	// Output:
	// 2023-01-01
	// 2019-01-01
}
func ExampleBetweenWithUnit() {
	start := time.Date(2026, 3, 4, 10, 0, 0, 0, time.UTC)

	value1, unit1 := BetweenWithUnit(start, start.Add(10*time.Second))
	value2, unit2 := BetweenWithUnit(start, start.Add(61*time.Second))
	value3, unit3 := BetweenWithUnit(start, start.Add(25*time.Hour))

	fmt.Println(value1, unit1)
	fmt.Println(value2, unit2)
	fmt.Println(value3, unit3)

	// Output:
	// 10 second
	// 2 minute
	// 2 day
}

func ExampleFormatStrToTime() {
	result1, _ := FormatStrToTime("2021-01-02 16:04:08", "yyyy-mm-dd hh:mm:ss")
	result2, _ := FormatStrToTime("2021-01-02", "yyyy-mm-dd")
	result3, _ := FormatStrToTime("02-01-21 16:04:08", "dd-mm-yy hh:mm:ss")

	fmt.Println(result1)
	fmt.Println(result2)
	fmt.Println(result3)

	// Output:
	// 2021-01-02 16:04:08 +0000 UTC
	// 2021-01-02 00:00:00 +0000 UTC
	// 2021-01-02 16:04:08 +0000 UTC
}
