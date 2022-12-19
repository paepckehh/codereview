package codereview

import "os"

// parse and compact an c/c++ and related header files
func (c *Config) parseC(filename string) Result {
	// setup
	f, err := os.ReadFile(filename)
	if err != nil {
		errOut("[error] [unable to read file] [" + filename + "]")
		return Result{}
	}
	fileSize := len(f)
	if fileSize < 3 {
		return c.finalizeWorker(0, 0, 0, filename, []byte{})
	}
	compact := make([]byte, 0, fileSize)
	lineTotal, lineSavings, comment, singleComment, single, double, pendingLF := 0, 0, false, false, false, false, false
	// parse
	for i := 0; i < fileSize; i++ {
		switch {
		case comment:
			switch {
			case f[i] == _linefeed:
				pendingLF = false
				lineTotal++
				lineSavings++
			case f[i] == _asterix:
				if i+1 < fileSize {
					if f[i+1] == _slashfwd {
						i++
						comment = false
						if pendingLF {
							pendingLF = false
							if i+1 < fileSize {
								if f[i+1] == _linefeed {
									lineSavings++
									continue
								}
							}
							compact = append(compact, _linefeed)
							continue
						}
						lineSavings++
					}
				}
			}
			continue
		case singleComment:
			switch f[i] {
			case _linefeed:
				singleComment = false
			default:
				continue
			}
		case single:
			if f[i] == _singleqoute {
				single = false
				if i-1 > 0 {
					if f[i-1] == _slashbwd {
						single = true
						if i-2 > 0 && f[i-2] == _slashbwd {
							single = false
						}
					}
				}
			}
			compact = append(compact, f[i])
			continue
		case double:
			if f[i] == _doubleqoute {
				double = false
				if i-1 > 0 {
					if f[i-1] == _slashbwd {
						double = true
						if i-2 > 0 && f[i-2] == _slashbwd {
							double = false
						}
					}
				}
			}
			compact = append(compact, f[i])
			continue
		}
		switch f[i] {
		case _singleqoute:
			single = true
			compact = append(compact, _singleqoute)
		case _doubleqoute:
			double = true
			compact = append(compact, _doubleqoute)
		case _slashbwd:
			compact = append(compact, _slashbwd)
			i++
			if i < fileSize {
				compact = append(compact, f[i])
			}
		case _slashfwd:
			if i+1 < fileSize {
				switch f[i+1] {
				case _slashfwd:
					singleComment = true
					i++
					continue
				case _asterix:
					comment = true
					i++
					continue
				}
			}
			compact = append(compact, _slashfwd)
		case _linefeed:
			lineTotal++
			if i+1 < fileSize {
				switch f[i+1] {
				case _linefeed:
					lineSavings++
					continue
				case _slashfwd:
					if i+2 < fileSize {
						switch f[i+2] {
						case _slashfwd:
							lineSavings++
							singleComment = true
							i += 2
							continue
						case _asterix:
							lineSavings++
							comment = true
							pendingLF = true
							i += 2
							continue
						}
						lineSavings++
						continue
					}
				}
			}
			if len(compact) > 0 {
				compact = append(compact, _linefeed)
			}
		default:
			compact = append(compact, f[i])
		}
	}
	return c.finalizeWorker(lineTotal, lineSavings, fileSize-len(compact), filename, compact)
}

func (c *Config) workerC() {
	for i := 0; i < worker; i++ {
		go func() {
			lineTotal, lineSaved, byteSaved := 0, 0, 0
			for filename := range channel_c {
				r := c.parseC(filename)
				lineTotal += r.TotalLines
				lineSaved += r.LineSavings
				byteSaved += r.ByteSavings
			}
			channel_statFile <- statFile{lineTotal: lineTotal, lineSavings: lineSaved, byteSavings: byteSaved}
			bg.Done()
		}()
	}
}
