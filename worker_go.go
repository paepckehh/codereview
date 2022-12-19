package codereview

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"

	gofumpt "mvdan.cc/gofumpt/format"
)

const (
	_goLangVersion       = "1.19"
	_goCompilerDirective = "//go:"
)

// parseGO and compact an golang file
func (c *Config) parseGO(filename string) Result {
	var compact []byte
	var lineTotal, lineSavings, byteSavings, compactSize int
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
	f = nil // do not try to parse broken, invalid & empty files, do not rely on fs.stat.size
	if fileSize < 1 {
		return c.finalizeWorker(0, 0, 0, filename, []byte{})
	}
	// read file & parse via golang ast
	t := token.NewFileSet()
	p, err := parser.ParseFile(t, filename, nil, parser.ParseComments)
	if err != nil && c.Verbose {
		errOut("[go] [unable to parse] [" + filename + "] [" + err.Error() + "]")
		return Result{}
	}
	// first round: pre-process and fix all compiler contrains, print
	o, n := bytes.Buffer{}, printer.Config{Mode: printer.RawFormat}
	if err = n.Fprint(&o, t, p); err != nil {
		errOut("[go] [unable to process] [" + filename + "]")
		return Result{}
	}
	// re-parse buffer, because go/printer/gobuild.go fixGoBuildlines() is not exposed
	tt := token.NewFileSet()
	pp, err := parser.ParseFile(tt, filename, o.Bytes(), parser.ParseComments)
	if err != nil && c.Verbose {
		errOut("[go] [unable to parse] [" + filename + "] [" + err.Error() + "]")
		return Result{}
	}
	// pre-process and format ast via mvdan.cc/gofumpt extended ruleset
	gofumpt.File(tt, pp, gofumpt.Options{LangVersion: _goLangVersion, ExtraRules: true})
	// clean all comments and outdated directives in ast tree
	for i, c := range pp.Comments {
		for ii, s := range c.List {
			if len(s.Text) > 10 && s.Text[:4] == _go {
				continue // skip, valid new compiler directive format
			}
			switch s.Text[1] {
			case _slashfwd:
				pp.Comments[i].List[ii].Text = "//\n"
			case _asterix:
				ln := strings.Count(pp.Comments[i].List[ii].Text, "\n") + 1
				pp.Comments[i].List[ii].Text = strings.Repeat("//\n", ln)
			}
		}
	}
	// out
	oo, nn := bytes.Buffer{}, printer.Config{Mode: printer.RawFormat}
	if err = nn.Fprint(&oo, tt, pp); err != nil {
		errOut("[go] [unable to process] [" + filename + "]")
		return Result{}
	}
	compact, compactSize, lineSavings = removeEmptyGoComments(oo.Bytes())
	byteSavings = fileSize - compactSize
	return c.finalizeWorker(lineTotal, lineSavings, byteSavings, filename, compact)
}

// workerGO ...
func (c *Config) workerGO() {
	for i := 0; i < worker; i++ {
		go func() {
			lineTotal, lineSaved, byteSaved := 0, 0, 0
			for filename := range channel_go {
				r := c.parseGO(filename)
				lineTotal += r.TotalLines
				lineSaved += r.LineSavings
				byteSaved += r.ByteSavings
			}
			channel_statFile <- statFile{lineTotal: lineTotal, lineSavings: lineSaved, byteSavings: byteSaved}
			bg.Done()
		}()
	}
}
