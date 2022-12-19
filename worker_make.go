package codereview

import "os"

// parseMake ...
func (c *Config) parseMake(filename string) Result {
	// setup
	f, err := os.ReadFile(filename)
	if err != nil {
		errOut("[error] [unable to read file] [" + filename + "]")
		return Result{}
	}
	fileSize := len(f)
	if fileSize < 6 {
		return c.finalizeWorker(0, 0, 0, filename, []byte{})
	}
	lineTotal, lineSavings, comment, single, double, space, value := 0, 0, false, false, false, false, false
	compact := make([]byte, 0, fileSize)
	// parse
	for i := 0; i < fileSize; i++ {
		switch {
		case value:
			if f[i] == _cbracketOFF {
				value = false
			}
			compact = append(compact, f[i])
			continue
		case space:
			switch f[i] {
			case _whitespace, _tab:
				continue
			default:
				space = false
			}
		case comment:
			switch f[i] {
			case _linefeed:
				lineSavings++
				comment = false
			default:
				continue
			}
		case single:
			if f[i] == _singleqoute {
				single = false
				if i-1 > 0 && f[i-1] == _slashbwd {
					single = true
					if i-2 > 0 && f[i-2] == _slashbwd {
						single = false
					}
				}
			}
			compact = append(compact, f[i])
			continue
		case double:
			if f[i] == _doubleqoute {
				double = false
				if i-1 > 0 && f[i-1] == _slashbwd {
					double = true
					if i-2 > 0 && f[i-2] == _slashbwd {
						double = false
					}
				}
			}
			compact = append(compact, f[i])
			continue
		}
		switch f[i] {
		case _whitespace:
			space = true
			compact = append(compact, _whitespace)
		case _tab:
			space = true
			compact = append(compact, _tab)
		case _dollar:
			switch {
			case i+1 < fileSize && f[i+1] == _cbracketON:
				i++
				value = true
				compact = append(compact, _dollar)
				compact = append(compact, _cbracketON)
			default:
				compact = append(compact, _dollar)
			}
		case _linefeed:
			lineTotal++
			if i+1 < fileSize {
				switch f[i+1] {
				case _linefeed:
					lineSavings++
				case _hashmark:
					i++
					lineSavings++
					comment = true
				default:
					if len(compact) > 0 {
						compact = append(compact, _linefeed)
					}
				}
			}
		case _hashmark:
			comment = true
		case _singleqoute:
			single = true
			compact = append(compact, _singleqoute)
		case _doubleqoute:
			double = true
			compact = append(compact, _doubleqoute)
		case _slashbwd:
			i++
			switch {
			case i < fileSize && f[i] == _linefeed:
			default:
				compact = append(compact, _slashbwd)
				compact = append(compact, f[i])
			}
		default:
			compact = append(compact, f[i])
		}
	}
	return c.finalizeWorker(lineTotal, lineSavings, fileSize-len(compact), filename, compact)
}

// workerMake ...
func (c *Config) workerMake() {
	for i := 0; i < worker; i++ {
		go func() {
			lineTotal, lineSaved, byteSaved := 0, 0, 0
			for filename := range channel_make {
				r := c.parseMake(filename)
				lineTotal += r.TotalLines
				lineSaved += r.LineSavings
				byteSaved += r.ByteSavings
			}
			channel_statFile <- statFile{lineTotal: lineTotal, lineSavings: lineSaved, byteSavings: byteSaved}
			bg.Done()
		}()
	}
}
