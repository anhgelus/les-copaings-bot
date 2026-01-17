package exp

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/anhgelus/gokord"
)

const DebugFactor = 30

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

type LevelScale struct{}

func (LevelScale) Normalize(min, max, x float64) float64 {
	if min < 0 || max < 0 || x < 0 {
		panic("Values must be positive or null for a level scale.")
	}
	levelMin := LevelExact(min)
	return (LevelExact(x) - levelMin) / (LevelExact(max) - levelMin)
}

// Level gives the level with the given XP.
// See LevelXP to get the XP required to get a level.
func Level(xp uint) uint {
	return uint(math.Floor(LevelExact(float64(xp))))
}

// LevelExact gives the exact level with the given XP.
// See Level to get the floored level.
func LevelExact(xp float64) float64 {
	return 0.2 * math.Sqrt(xp)
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
	var unix time.Time
	if gokord.Debug {
		unix = time.Unix(time.Now().Unix()-int64(DebugFactor*n), 0) // reduce time for debug
	} else {
		unix = time.Unix(time.Now().Unix()-int64(n*24*60*60), 0)
	}
	unix = unix.UTC()
	return fmt.Sprintf("%d-%d-%d %d:%d:%d UTC", unix.Year(), unix.Month(), unix.Day(), unix.Hour(), unix.Minute(), unix.Second())
}

func TrimMessage(s string) string {
	not := regexp.MustCompile("[^a-zA-Z0-9éèêàùûç,;:!.?]")
	ping := regexp.MustCompile("<(@&?|#)[0-9]{18}>")
	link := regexp.MustCompile("https?://[a-zA-Z0-9.]+[.][a-z]+.*")

	s = ping.ReplaceAllLiteralString(s, "")
	s = link.ReplaceAllLiteralString(s, "")
	s = not.ReplaceAllLiteralString(s, "")

	return strings.Trim(s, " ")
}
