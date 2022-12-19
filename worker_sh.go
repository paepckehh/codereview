package codereview

import (
	"bytes"
	"os"
	"strings"

	shfmtsyntax "mvdan.cc/sh/syntax"
)

// parseSH and compact an posix complient shell script
func (c *Config) parseSH(filename string) Result {
	var compact []byte
	var lineTotal, lineSavings, byteSavings int
	defer func() {
		if r := recover(); r != nil {
			out("[codereview] [external ast parser crash/panic] [skip] [" + filename + "]")
		}
	}()
	f, err := os.ReadFile(filename)
	if err != nil {
		errOut("[error] [unable to read file] [" + filename + "]")
		return Result{}
	}
	lineTotal, fileSize := bytes.Count(f, _lf), len(f)
	if fileSize < 1 {
		return c.finalizeWorker(0, 0, 0, filename, []byte{})
	}
	parser := shfmtsyntax.NewParser()
	printer := shfmtsyntax.NewPrinter(shfmtsyntax.Minify(false))
	code, err := parser.Parse(bytes.NewReader(f), filename)
	if err != nil {
		if !strings.Contains(string(f[:32]), "tcl") && c.Verbose {
			errOut("[unable to parse] [sh] [" + filename + "] [" + err.Error() + "]")
		}
		return Result{}
	}
	buffer := bytes.Buffer{}
	printer.Print(&buffer, code)
	co := []byte("#!/bin/sh\n")
	co = append(co, buffer.Bytes()...)
	compact, _ = removeEmptyLines(co)
	lineSavings = lineTotal - bytes.Count(compact, _lf)
	byteSavings = fileSize - len(compact)
	return c.finalizeWorker(lineTotal, lineSavings, byteSavings, filename, compact)
}

// workerSH
func (c *Config) workerSH() {
	for i := 0; i < worker; i++ {
		go func() {
			lineTotal, lineSaved, byteSaved := 0, 0, 0
			for filename := range channel_sh {
				r := c.parseSH(filename)
				lineTotal += r.TotalLines
				lineSaved += r.LineSavings
				byteSaved += r.ByteSavings
			}
			channel_statFile <- statFile{lineTotal: lineTotal, lineSavings: lineSaved, byteSavings: byteSaved}
			bg.Done()
		}()
	}
}
