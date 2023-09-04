package main

import (
	"fmt"

	"github.com/fatih/color"
)

type taskFlag struct {
	Label string
	Color func(string, ...interface{}) string
}

// Task Flag Types
var (
	lStandard taskFlag = taskFlag{"", color.HiWhiteString}
	lVerbose  taskFlag = taskFlag{"VERBOSE", color.CyanString}
	lDebug    taskFlag = taskFlag{"DEBUG", color.HiCyanString}
	lDebug2   taskFlag = taskFlag{"DEBUG2", color.HiBlueString}
	lInfo     taskFlag = taskFlag{"INFO", color.BlueString}
	lTip      taskFlag = taskFlag{"TIP", color.HiMagentaString}
	lSuccess  taskFlag = taskFlag{"SUCCESS", color.HiGreenString}
	lWarning  taskFlag = taskFlag{"WARNING", color.HiYellowString}
	lError    taskFlag = taskFlag{"ERROR", color.HiRedString}
	lErrorInf taskFlag = taskFlag{"ERROR DETAILS", color.RedString}
)

// Log Instructions
type logInstructions struct {
	Location string
	Task     string
	Flag     *taskFlag
	Inline   bool
	Color    func(string, ...interface{}) string
}

func LogWrapper(location string, task string, flag *taskFlag, inline bool,
	colorFunc func(string, ...interface{}) string,
	body string, bodyParams ...interface{}) string {

	//TODO: Alternative handling of log content for non-colored output. -- adminChannel.LogProgram, etc etc

	linePrefix := location
	if task != "" {
		linePrefix = fmt.Sprintf("%s >> %s", location, task)
	}
	if flag != nil {
		linePrefix += flag.Color(" < %s >", flag.Label)
	}

	if !inline {
		linePrefix += "\n\n\t"
	} else {
		linePrefix += " - "
	}

	lineSuffix := ""
	if !inline {
		lineSuffix += "\n"
	}

	formatted := fmt.Sprintf(body, bodyParams...)

	return color.WhiteString(linePrefix) + colorFunc(formatted) + lineSuffix
}

func (l logInstructions) SetFlag(flag *taskFlag) logInstructions {
	l.Flag = flag
	l.Color = flag.Color
	return l
}

func (l logInstructions) ClearFlag() logInstructions {
	l.Flag = nil
	return l
}

func (l logInstructions) SetTask(task string) logInstructions {
	l.Task = task
	return l
}

func (l logInstructions) ClearTask() logInstructions {
	l.Task = ""
	return l
}

func (l logInstructions) Clear() logInstructions {
	l.ClearFlag()
	l.ClearTask()
	l.Inline = false
	return l
}

// follow config entirely
func (l logInstructions) Log(body string, bodyParams ...interface{}) string {
	return LogWrapper(l.Location, l.Task, l.Flag, l.Inline, l.Color, body, bodyParams...)
}

// for overriding color
func (l logInstructions) LogC(colorFunc func(string, ...interface{}) string, body string, bodyParams ...interface{}) string {
	return LogWrapper(l.Location, l.Task, l.Flag, l.Inline, colorFunc, body, bodyParams...)
}

// for overriding inline
func (l logInstructions) LogI(inline bool, body string, bodyParams ...interface{}) string {
	return LogWrapper(l.Location, l.Task, l.Flag, inline, l.Color, body, bodyParams...)
}

// for overriding color and inline
func (l logInstructions) LogCI(colorFunc func(string, ...interface{}) string, inline bool, body string, bodyParams ...interface{}) string {
	return LogWrapper(l.Location, l.Task, l.Flag, inline, colorFunc, body, bodyParams...)
}
