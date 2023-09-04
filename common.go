package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hako/durafmt"
)

//#region Program functions

func uptime() time.Duration {
	return time.Since(timeLaunched) //.Truncate(time.Second)
}

func properExit() {
	// Not formatting string because I only want the exit message to be red.
	log.Println(color.HiRedString("[EXIT IN 15 SECONDS] Uptime was %s...", durafmt.Parse(time.Since(timeLaunched)).String()))
	log.Println(color.HiCyanString("--------------------------------------------------------------------------------"))
	time.Sleep(15 * time.Second)
	os.Exit(1)
}

//#endregion

//#region String functions

func containsAll(haystack string, needles []string) bool {
	for _, needle := range needles {
		if !strings.Contains(haystack, needle) {
			return false
		}
	}
	return true
}

func containsAny(haystack string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(haystack, needle) {
			return true
		}
	}
	return false
}

func wrapHyphens(i string, l int) string {
	n := i
	if len(n) < l {
		n = "- " + n + " -"
		for len(n) < l {
			n = "-" + n + "-"
		}
	}
	return n
}

func wrapHyphensW(i string) string {
	return wrapHyphens(i, 80)
}

func hexdec(s string) (string, error) {
	ret, err := strconv.ParseInt(strings.ReplaceAll(strings.ReplaceAll(s, "0x", ""), "#", ""), 16, 64)
	return fmt.Sprint(ret), err
}

func ssuff(i int) string {
	if i == 1 {
		return ""
	}
	return "s"
}

func disableLinks(s string) string {
	s = strings.ReplaceAll(s, "https://", "")
	s = strings.ReplaceAll(s, "http://", "")
	s = strings.ReplaceAll(s, "www.", "")
	return s
}

func shortenTime(input string) string {
	input = strings.ReplaceAll(input, " nanoseconds", "ns")
	input = strings.ReplaceAll(input, " nanosecond", "ns")
	input = strings.ReplaceAll(input, " microseconds", "μs")
	input = strings.ReplaceAll(input, " microsecond", "μs")
	input = strings.ReplaceAll(input, " milliseconds", "ms")
	input = strings.ReplaceAll(input, " millisecond", "ms")
	input = strings.ReplaceAll(input, " seconds", "s")
	input = strings.ReplaceAll(input, " second", "s")
	input = strings.ReplaceAll(input, " minutes", "m")
	input = strings.ReplaceAll(input, " minute", "m")
	input = strings.ReplaceAll(input, " hours", "h")
	input = strings.ReplaceAll(input, " hour", "h")
	input = strings.ReplaceAll(input, " days", "d")
	input = strings.ReplaceAll(input, " day", "d")
	input = strings.ReplaceAll(input, " weeks", "w")
	input = strings.ReplaceAll(input, " week", "w")
	input = strings.ReplaceAll(input, " months", "mo")
	input = strings.ReplaceAll(input, " month", "mo")
	return input
}

//#endregion
