package main

import (
	"math"
	"regexp"
	"strconv"
	"time"
)

type CaseInsensitiveReplacer struct {
	toReplace   *regexp.Regexp
	replaceWith string
}

func NewCaseInsensitiveReplacer(toReplace, with string) *CaseInsensitiveReplacer {
	return &CaseInsensitiveReplacer{
		toReplace:   regexp.MustCompile("(?i)" + toReplace),
		replaceWith: with,
	}
}
func (cir *CaseInsensitiveReplacer) Replace(str string) string {
	return cir.toReplace.ReplaceAllString(str, cir.replaceWith)
}

func zeroPad(str string) (result string) {
	if len(str) < 2 {
		result = "0" + str
	} else {
		result = str
	}
	return
}

func secondsToHuman(input float64) (result string) {
	hours := math.Floor(float64(input) / 60 / 60)
	seconds := int(input) % (60 * 60)
	minutes := math.Floor(float64(seconds) / 60)
	seconds = int(input) % 60

	if hours > 0 {
		result = strconv.Itoa(int(hours)) + ":" + zeroPad(strconv.Itoa(int(minutes))) + ":" + zeroPad(strconv.Itoa(int(seconds)))
	} else {
		result = zeroPad(strconv.Itoa(int(minutes))) + ":" + zeroPad(strconv.Itoa(int(seconds)))
	}

	return
}

func roundTime(d, r time.Duration) time.Duration {
	if r <= 0 {
		return d
	}
	neg := d < 0
	if neg {
		d = -d
	}
	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}
	if neg {
		return -d
	}
	return d
}
