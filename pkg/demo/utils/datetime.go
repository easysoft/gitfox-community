package main

import (
	"fmt"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

func main() {
	ExampleGetNowDate()
	ExampleGetNowTime()
	ExampleGetNowDateTime()
	ExampleFormatTimeToStr()
	ExampleAddDay()
	ExampleAddWeek()
	ExampleAddMonth()
	ExampleAddHour()
}

func ExampleGetNowDate() {
	result := datetime.GetNowDate()

	expected := time.Now().Format("2006-01-02")

	fmt.Println("ExampleGetNowDate")
	fmt.Println(result == expected)

	// Output:
	// true
}

func ExampleGetNowTime() {
	result := datetime.GetNowTime()

	expected := time.Now().Format("15:04:05")

	fmt.Println("ExampleGetNowTime")
	fmt.Println(result == expected)

	// Output:
	// true
}

func ExampleGetNowDateTime() {
	result := datetime.GetNowDateTime()

	expected := time.Now().Format("2006-01-02 15:04:05")

	fmt.Println("ExampleGetNowDateTime")
	fmt.Println(result == expected)

	// Output:
	// true
}

func ExampleFormatTimeToStr() {
	t, _ := time.Parse("2006-01-02 15:04:05", "2021-01-02 16:04:08")

	result1 := datetime.FormatTimeToStr(t, "yyyy-mm-dd hh:mm:ss")
	result2 := datetime.FormatTimeToStr(t, "yyyy-mm-dd")
	result3 := datetime.FormatTimeToStr(t, "dd-mm-yy hh:mm:ss")
	result4 := datetime.FormatTimeToStr(t, "yyyy-mm-dd hh")

	fmt.Println("ExampleFormatTimeToStr")
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

	after1Day := datetime.AddDay(date, 1)
	before1Day := datetime.AddDay(date, -1)

	fmt.Println("ExampleAddDay")
	fmt.Println(after1Day.Format("2006-01-02 15:04:05"))
	fmt.Println(before1Day.Format("2006-01-02 15:04:05"))

	// Output:
	// 2021-01-02 00:00:00
	// 2020-12-31 00:00:00
}

func ExampleAddWeek() {
	date, _ := time.Parse("2006-01-02", "2021-01-01")

	after2Weeks := datetime.AddWeek(date, 2)
	before2Weeks := datetime.AddWeek(date, -2)

	fmt.Println("ExampleAddWeek")
	fmt.Println(after2Weeks.Format("2006-01-02"))
	fmt.Println(before2Weeks.Format("2006-01-02"))

	// Output:
	// 2021-01-15
	// 2020-12-18
}

func ExampleAddMonth() {
	date, _ := time.Parse("2006-01-02", "2021-01-01")

	after2Months := datetime.AddMonth(date, 2)
	before2Months := datetime.AddMonth(date, -2)

	fmt.Println("ExampleAddMonth")
	fmt.Println(after2Months.Format("2006-01-02"))
	fmt.Println(before2Months.Format("2006-01-02"))

	// Output:
	// 2021-03-01
	// 2020-11-01
}

func ExampleAddHour() {
	date, _ := time.Parse("2006-01-02 15:04:05", "2021-01-01 00:00:00")

	after2Hours := datetime.AddHour(date, 2)
	before2Hours := datetime.AddHour(date, -2)

	fmt.Println("ExampleAddHour")
	fmt.Println(after2Hours.Format("2006-01-02 15:04:05"))
	fmt.Println(before2Hours.Format("2006-01-02 15:04:05"))

	// Output:
	// 2021-01-01 02:00:00
	// 2020-12-31 22:00:00
}

func ExampleAddMinute() {
	date, _ := time.Parse("2006-01-02 15:04:05", "2021-01-01 00:00:00")

	after2Minutes := datetime.AddMinute(date, 2)
	before2Minutes := datetime.AddMinute(date, -2)

	fmt.Println("ExampleAddMinute")
	fmt.Println(after2Minutes.Format("2006-01-02 15:04:05"))
	fmt.Println(before2Minutes.Format("2006-01-02 15:04:05"))

	// Output:
	// 2021-01-01 00:02:00
	// 2020-12-31 23:58:00
}

func ExampleAddDuration() {
	date, _ := time.Parse("2006-01-02 15:04:05", "2021-01-01 00:00:00")

	after2Minutes := datetime.AddDuration(date, 2*time.Minute)
	before2Minutes := datetime.AddDuration(date, -2*time.Minute)

	fmt.Println("ExampleAddDuration")
	fmt.Println(after2Minutes.Format("2006-01-02 15:04:05"))
	fmt.Println(before2Minutes.Format("2006-01-02 15:04:05"))

	// Output:
	// 2021-01-01 00:02:00
	// 2020-12-31 23:58:00
}

func ExampleAddYear() {
	date, _ := time.Parse("2006-01-02", "2021-01-01")

	after2Years := datetime.AddYear(date, 2)
	before2Years := datetime.AddYear(date, -2)

	fmt.Println("ExampleAddYear")
	fmt.Println(after2Years.Format("2006-01-02"))
	fmt.Println(before2Years.Format("2006-01-02"))

	// Output:
	// 2023-01-01
	// 2019-01-01
}
func ExampleFormatStrToTime() {
	result1, _ := datetime.FormatStrToTime("2021-01-02 16:04:08", "yyyy-mm-dd hh:mm:ss")
	result2, _ := datetime.FormatStrToTime("2021-01-02", "yyyy-mm-dd")
	result3, _ := datetime.FormatStrToTime("02-01-21 16:04:08", "dd-mm-yy hh:mm:ss")

	fmt.Println("ExampleFormatStrToTime")
	fmt.Println(result1)
	fmt.Println(result2)
	fmt.Println(result3)

	// Output:
	// 2021-01-02 16:04:08 +0000 UTC
	// 2021-01-02 00:00:00 +0000 UTC
	// 2021-01-02 16:04:08 +0000 UTC
}
