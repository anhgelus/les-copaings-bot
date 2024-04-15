package xp

import (
	"github.com/anhgelus/gokord"
	"math"
)

func XPMessage(length uint, diversity uint) uint {
	return uint(math.Floor(
		0.025*math.Pow(float64(length), 1.25)*math.Sqrt(float64(diversity)) + 1,
	))
}

func XPVocal(time uint) uint {
	return uint(math.Floor(
		0.01*math.Pow(float64(time), 1.3) + 1,
	))
}

func Level(xp uint) uint {
	return uint(math.Floor(
		0.2 * math.Sqrt(float64(xp)),
	))
}

func XPForLevel(level uint) uint {
	return uint(math.Floor(
		math.Pow(float64(5*level), 2),
	))
}

func Lose(time uint, xp uint) uint {
	if gokord.Debug {
		return uint(math.Floor(
			math.Pow(float64(time), 3) * math.Pow(10, -2+math.Log(float64(time))) * math.Floor(float64(xp/500)+1),
		)) // a little bit faster to lose xp
	}
	return uint(math.Floor(
		math.Pow(float64(time), 2) * math.Pow(10, -2+math.Log(float64(time/85))) * math.Floor(float64(xp/500)+1),
	))
}
