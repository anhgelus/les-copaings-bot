package exp

import (
	"fmt"
	"github.com/anhgelus/gokord"
	"math"
	"slices"
	"time"
)

func MessageXP(length uint, diversity uint) uint {
	return uint(math.Floor(
		0.025*math.Pow(float64(length), 1.25)*math.Sqrt(float64(diversity)) + 1,
	))
}

func CalcDiversity(msg string) uint {
	var chars []rune
	for _, c := range []rune(msg) {
		if !slices.Contains(chars, c) {
			chars = append(chars, c)
		}
	}
	return uint(len(chars))
}

func VocalXP(time uint) uint {
	return uint(math.Floor(
		0.01*math.Pow(float64(time), 1.3) + 1,
	))
}

// Level gives the level with the given XP.
// See LevelXP to get the XP required to get a level.
func Level(xp uint) uint {
	return uint(math.Floor(
		0.2 * math.Sqrt(float64(xp)),
	))
}

// LevelXP gives the XP required to get this level.
// See Level to get the level with the given XP.
func LevelXP(level uint) uint {
	return uint(math.Floor(
		25 * math.Pow(float64(level), 2),
	))
}

// TimeStampNDaysBefore returns the timestamp (year-month-day) n days before today
func TimeStampNDaysBefore(n uint) string {
	var y, d int
	var m time.Month
	if gokord.Debug {
		y, m, d = time.Unix(time.Now().Unix()-int64(24*60*60), 0).Date() // reduce time for debug
	} else {
		y, m, d = time.Unix(time.Now().Unix()-int64(n*24*60*60), 0).Date()
	}
	return fmt.Sprintf("%d-%d-%d", y, m, d)
}
