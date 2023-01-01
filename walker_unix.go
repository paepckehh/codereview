//go:build windows

// walker windows does not implement (yet) skiping hardlinked
// files(inode seen based on windows 'fileid') and no support
// for filesystem boundary crossing checks
package codereview

import (
	"os"
)

const (
	_modeDir     uint32 = 1 << (32 - 1 - 0)
	_modeSymlink uint32 = 1 << (32 - 1 - 4)
)

// fastWalker ...
func (c *Config) fastWalker() {
	channel_dir <- c.Path
	for i := 0; i < worker; i++ {
		go c.walkParse()
	}
}

// walkParse ...
func (c *Config) walkParse() {
	exclude, skipme := false, false
	if len(c.Exclude) > 0 {
		exclude = true
	}
	total, processed := 0, 0
	for path := range channel_dir {
		list, err := os.ReadDir(path)
		if err != nil {
			errOut("unable to read directory [" + path + "] [" + err.Error() + "]")
			walk.Done()
			continue
		}
		for _, item := range list {
			total++
			fname := item.Name()
			if c.SkipHidden {
				if fname[0] == '.' {
					continue // skip hidden files if requested
				}
			}
			ftype := uint32(item.Type())
			name := path + "/" + fname
			if ftype&_modeSymlink != 0 {
				continue // skip symlinks
			}
			if ftype&_modeDir != 0 {
				if exclude {
					skipme = false
					for _, exclude := range c.Exclude {
						if fname == exclude {
							skipme = true
							break // skip exclude list
						}
					}
					if skipme {
						continue
					}
				}
				walk.Add(1)
				channel_dir <- name
				continue
			}
			l := len(fname)
			if l > 2 {
				if c.LangC {
					switch fname[l-2:] {
					case ".c", ".h":
						processed++
						channel_c <- name
						continue
					}
				}
				if c.LangASM {
					if fname[l-2:] == ".s" || fname[l-2:] == ".S" {
						processed++
						channel_asm <- name
						continue
					}
				}
			}
			if l > 3 {
				if c.LangGO && fname[l-3:] == ".go" {
					processed++
					channel_go <- name
					continue
				}
				if c.LangSH && fname[l-3:] == ".sh" {
					processed++
					channel_sh <- name
					continue
				}
				if c.LangC {
					if fname[l-3:] == ".cc" || fname[l-3:] == ".hh" {
						processed++
						channel_c <- name
						continue
					}
				}
			}
			if l > 4 {
				if c.LangC && fname[l-4:] == ".cpp" {
					processed++
					channel_c <- name
					continue
				}
			}
			if l > 7 {
				if c.LangMAKE && fname[:8] == "Makefile" {
					processed++
					channel_make <- name
					continue
				}
			}
			// if c.LangC && shebang(name) == "#!/bin/sh" {
			//	processed++
			//	channel_sh <- name
			//	continue
			// }
		}
		walk.Done()
	}
	channel_statFiles <- statFiles{totalFiles: total, processedFiles: processed}
	walkStat.Done()
}
