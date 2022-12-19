// package main ...
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"paepcke.de/codereview"
)

// main ...
func main() {
	t0 := time.Now()
	c := codereview.NewConfig()
	l := len(os.Args)
	switch {
	case l > 1:
		opt, t, p, x, s, b := "", 0, 0, 0, 0, 0
		for i := 1; i < l; i++ {
			o := os.Args[i]
			switch {
			case o[0] == '-':
				switch {
				case o == "--inplace" || o == "-w":
					c.Inplace = true
					opt += "[--inplace] "
				case o == "--exclude" || o == "-e":
					i++
					switch {
					case i < l:
						c.Exclude = append(c.Exclude, os.Args[i])
						opt += "[--exclude " + os.Args[i] + "] "
					default:
						errExit("exclude switch value missing")
					}
				case o == "--help" || o == "-h":
					out(_syntax)
					os.Exit(0)
				case o == "--hash" || o == "-H":
					c.HashOut = true
				case o == "--verbose" || o == "-v":
					c.Verbose = true
					opt += "[--verbose] "
				case o == "--silent" || o == "-q":
					c.Silent = true
					opt += "[--silent] "
				case o == "--debug" || o == "-d":
					c.Debug = true
					opt += "[--debug] "
				case o == "--disable-c":
					c.LangC = false
					opt += "[--disable-c] "
				case o == "--disable-go":
					c.LangGO = false
					opt += "[--disable-go] "
				case o == "--disable-sh":
					c.LangSH = false
					opt += "[--disable-sh] "
				case o == "--disable-asm":
					c.LangASM = false
					opt += "[--disable-asm] "
				case o == "--disable-make":
					c.LangMAKE = false
					opt += "[--disable-make] "
				case o == "--disable-hidden-files":
					c.SkipHidden = true
					opt += "[--disable-hidden-files] "
				default:
					errExit("unkown commandline switch [" + o + "]")
				}
			case o == ".", o == "*":
				if c.Path != "" {
					errExit("more than one [file|directory] path specified")
				}
				var err error
				if c.Path, err = os.Getwd(); err != nil {
					errExit("invalid current directory [.] path")
				}
			case isFile(o):
				if c.Path != "" {
					errExit(" more than one [file|directory] path specified")
				}
				c.Path = o
				c.SingleFile = true
			case isDir(o):
				if c.Path != "" {
					errExit("more than one [file|directory] path specified")
				}
				c.Path = o
			default:
				errExit("invalid path [" + o + "]")
			}
		}
		r := codereview.Result{}
		switch {
		case c.SingleFile:
			var code []byte
			t, p = 1, 1
			r = c.ParseFileSheBang()
			if !r.Found || r.TotalLines == 0 {
				errExit(string(code))
			}
			x, s, b, code = r.TotalLines, r.LineSavings, r.ByteSavings, r.Result
		default:
			if !c.Inplace {
				c.HashOut = true
			}
			out("CODEREVIEW [start] [" + c.Path + "] " + opt)
			tt := c.WalkDir()
			t, p, x, s, b = tt.TotalFiles, tt.ProcessedFiles, tt.TotalLines, tt.TotalSavingsLines, tt.TotalSavingsBytes

		}
		report(c, t, p, x, s, b, t0)
	default:
		out(_syntax)
	}
}

const (
	_syntax string = "syntax: codereview [options] <file|directory>\n\n--inplace [-w]\n\t\twrite changes direct back to source files\n\t\twithout this option changes outputs goes to stdout\n\n--hash [-H]\n\t\tshow hashsum only\n\n--exclude [-e]\n\t\texclude all directories matching any of the keywords\n\t\tthis option can be specified several times\n\n--disable-c\n--disable-go\n--disable-sh\n--disable-asm\n--disable-make\n--disable-hidden-files\n\t\tdisable [files|language support] in recursive directory walk\n\n--verbose [-v]\n--silent [-q]\n--debug [-d]\n--help [-h]\n"
)

// report ...
func report(c *codereview.Config, t, p, x, s, b int, t0 time.Time) {
	if c.SingleFile && c.HashOut {
		return
	}
	r := "] [removeable: "
	if c.Inplace {
		r = "] [lines removed: "
	}
	out("CODEREVIEW [_done] [" + time.Since(t0).String() + "]")
	out("CODEREVIEW [stats] [total files: " + strconv.Itoa(t) + "] [processed: " + strconv.Itoa(p) + "]")
	out("CODEREVIEW [stats] [total loc: " + strconv.Itoa(x) + r + strconv.Itoa(s) + "] [savings: " + hruIEC(uint64(b), "bytes") + "]")
}

//
// LITTLE GENERIC HELPER SECTION
//

// const ...
const (
	_modeDir uint32 = 1 << (32 - 1 - 0)
)

// out ...
func out(msg string) {
	os.Stdout.Write([]byte(msg + "\n"))
}

// errExit ...
func errExit(msg string) {
	out("[error] " + msg)
	os.Exit(1)
}

// isDir ...
func isDir(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return uint32(fi.Mode())&_modeDir != 0
}

// isFile ...
func isFile(filename string) bool {
	fi, err := os.Lstat(filename)
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular()
}

// hruIEC converts value to hru IEC 60027 units
func hruIEC(i uint64, u string) string {
	return hru(i, 1024, u)
}

// hru [human readable units] backend
func hru(i, unit uint64, u string) string {
	if i < unit {
		return fmt.Sprintf("%d %s", i, u)
	}
	div, exp := unit, 0
	for n := i / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	switch u {
	case "":
		return fmt.Sprintf("%.3f %c", float64(i)/float64(div), "kMGTPE"[exp])
	case "bit":
		return fmt.Sprintf("%.0f %c%s", float64(i)/float64(div), "kMGTPE"[exp], u)
	case "bytes", "bytes/sec":
		return fmt.Sprintf("%.1f %c%s", float64(i)/float64(div), "kMGTPE"[exp], u)
	}
	return fmt.Sprintf("%.3f %c%s", float64(i)/float64(div), "kMGTPE"[exp], u)
}
