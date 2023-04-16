package main

import (
	"log"
	"os"
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
