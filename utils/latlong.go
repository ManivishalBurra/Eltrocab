package utils

import "math/rand"

func Generatelatlong() [2]float64 {
	lat := -90 + rand.Float64()*(180)
	long := -180 + rand.Float64()*(360)
	var arr [2]float64
	arr[0] = lat
	arr[1] = long
	return arr
}
