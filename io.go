package codereview

import (
	"os"
)

//
// Display IO
//

func out(msg string) {
	os.Stdout.Write([]byte(msg + "\n"))
}

//
// Error Display IO
//

var silent bool

func errOut(msg string) {
	if !silent {
		out("[error] " + msg)
	}
}

func errExit(msg string) {
	errOut(msg)
	os.Exit(1)
}

//
// File IO
//

func shebang(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		return ""
	}
	shebang := make([]byte, 9)
	n, err := f.Read(shebang)
	if err != nil || n < 9 {
		return ""
	}
	return string(shebang)
}
