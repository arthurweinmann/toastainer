package utils

import (
	"strconv"
	"time"
)

func GetMonthYear(n time.Time) int {
	return CombineMonthYear(int(n.Month()), n.Year())
}

func CombineMonthYear(month, year int) int {
	t, _ := strconv.Atoi(strconv.Itoa(month) + strconv.Itoa(year))
	return t
}
