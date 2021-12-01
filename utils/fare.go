package utils

import (
	"math"
)

func Fare(lat float64, long float64, dstlat float64, dstlng float64) [2]float64 {
	R := 6371e3
	l1 := lat * (math.Pi / 180)
	l2 := dstlat * (math.Pi / 180)
	diff1 := (dstlat - lat) * (math.Pi / 180)
	diff2 := (dstlng - long) * (math.Pi / 180)
	a := (math.Sin(diff1/2) * math.Sin(diff1/2)) + (math.Cos(l1) * math.Cos(l2) * math.Sin(diff2/2) * math.Sin(diff2/2))
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := R * c / 1000
	var arr [2]float64
	arr[0] = d
	basefare := d * 5
	GST := 0.050 * basefare
	arr[1] = basefare + GST + 20
	return arr
}
