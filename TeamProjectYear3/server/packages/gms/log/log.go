package log

import (
	"fmt"
	"time"
)

// const CLR_0B = "\x1b[30;1m"
const CLR_RB = "\x1b[31;1m"
const CLR_GB = "\x1b[32;1m"
const CLR_YB = "\x1b[33;1m"
const CLR_BB = "\x1b[34;1m"
const CLR_MB = "\x1b[35;1m"

// const CLR_CB = "\x1b[36;1m"
// const CLR_WB = "\x1b[37;1m"
const CLR_N = "\x1b[0m"

const dateFormatString = "15:04:05"

func Debug(message string) {
	fmt.Printf("%s%s%s%s%s\n", time.Now().Format(dateFormatString), CLR_BB, " Debug: ", CLR_N, message)
}

func Info(message string) {
	fmt.Printf("%s%s%s%s%s\n", time.Now().Format(dateFormatString), CLR_GB, " Info: ", CLR_N, message)
}

func Silly(message string) {
	fmt.Printf("%s%s%s%s%s\n", time.Now().Format(dateFormatString), CLR_MB, " Silly: ", CLR_N, message)
}

func Warning(message string) {
	fmt.Printf("%s%s%s%s%s\n", time.Now().Format(dateFormatString), CLR_YB, " Warning: ", CLR_N, message)
}

func Error(message string) {
	fmt.Printf("%s%s%s%s%s\n", time.Now().Format(dateFormatString), CLR_RB, " Error: ", CLR_N, message)
}
