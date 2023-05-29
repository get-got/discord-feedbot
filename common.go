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
