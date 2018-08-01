package cli

import (
	"fmt"
	"os"
)

// fatalPrintln prints the given string (similar to fmt.Println) and exits the process with a non zero code.
func fatalPrintln(printable interface{}) {
	fmt.Println(printable)
	os.Exit(1)
}

// fatalPrintf prints the given string (similar to fmt.Printf) and exits the process with a non zero code.
func fatalPrintf(format string, args ...interface{}) {
	fmt.Printf(format, args)
	os.Exit(1)
}
