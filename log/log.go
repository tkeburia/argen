package log

import (
	"fmt"
	"github.com/spf13/viper"
)

func Verbose(message string, args ...interface{}) {
	if viper.GetBool("verbose") {
		fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(message, args...))
	}
}

func Info(message string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(message, args...))
}

func Error(message string, args ...interface{}) {
	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf(message, args...))
}

func ErrorS(message string) string {
	return fmt.Sprintf("\x1b[31;1m%s\x1b[0m\n", message)
}
