package codereview

import (
	"bytes"
	"os"
)

// parseASM and compact assembly file
func (c *Config) parseASM(filename string) Result {
	// setup
	f, err := os.ReadFile(filename)
	if err != nil {
		errOut("[error] [unable to read file] [" + filename + "]")
		return Result{}
	}
	lineTotal, fileSize := bytes.Count(f, _lf), len(f)
	if fileSize < 1 {
		return c.finalizeWorker(0, 0, 0, filename, []byte{})
	}
	lineSavings, comment := 0, false
	compact := make([]byte, 0, fileSize)
	// parse
	for i := 0; i < fileSize; i++ {
		if comment {
			switch f[i] {
			case _linefeed:
				comment = false
				lineSavings++
				compact = append(compact, _linefeed)
			case 'g':
				if i+2 < fileSize {
					if f[i+1] == 'o' && f[i+2] == ':' { // valid compiler directive for go
						i += 2
						comment = false
						compact = append(compact, _goDS...)
						continue
					}
				}
			case ' ':
				if i+2 < fileSize {
					if f[i+1] == '+' && f[i+2] == 'b' { // ups, golang team forgot about linting internal asm
						i += 2
						comment = false
						compact = append(compact, _goDSO...)
						continue
					}
				}
			}
			continue
		}
		switch f[i] {
		case _linefeed:
			compact = append(compact, _linefeed)
		case _slashfwd:
			if i+1 < fileSize {
				if f[i+1] == _slashfwd {
					i++
					comment = true
					continue
				}
			}
			compact = append(compact, _slashfwd)
		default:
			compact = append(compact, f[i])
		}
	}
	// out
	return c.finalizeWorker(lineTotal, lineSavings, fileSize-len(compact), filename, compact)
}

// workerASM ...
func (c *Config) workerASM() {
	for i := 0; i < worker; i++ {
		go func() {
			lineTotal, lineSaved, byteSaved := 0, 0, 0
			for filename := range channel_asm {
				r := c.parseASM(filename)
				lineTotal += r.TotalLines
				lineSaved += r.LineSavings
				byteSaved += r.ByteSavings
			}
			channel_statFile <- statFile{lineTotal: lineTotal, lineSavings: lineSaved, byteSavings: byteSaved}
			bg.Done()
		}()
	}
}
