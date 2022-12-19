package codereview

import (
	"os"
	"syscall"
)

const (
	_modeDir     uint32 = 1 << (32 - 1 - 0)
	_modeSymlink uint32 = 1 << (32 - 1 - 4)
)

// fastWalker ...
func (c *Config) fastWalker() {
	fi, err := os.Stat(c.Path)
	if err != nil {
		errExit("[stat deviceID root dir] [" + c.Path + "] [" + err.Error() + "]")
	}
	channel_dir <- c.Path
	d, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		errExit("[stat deviceID root dir] [" + c.Path + "]")
	}
	for i := 0; i < worker; i++ {
		go c.walkParse(d, uint64(d.Dev))
	}
}

// walkParse ...
func (c *Config) walkParse(d *syscall.Stat_t, rootNodeDeviceID uint64) {
	exclude, skipme := false, false
	if len(c.Exclude) > 0 {
		exclude = true
	}
	inodeSeen, total, processed := make(map[uint64]struct{}), 0, 0
	for path := range channel_dir {
		list, err := os.ReadDir(path)
		if err != nil {
			errOut("unable to read directory [" + path + "] [" + err.Error() + "]")
			walk.Done()
			continue
		}
		for _, item := range list {
			total++
			fi, _ := item.Info()
			fname := item.Name()
			if c.SkipHidden {
				if fname[0] == '.' {
					continue // skip hidden files if requested
				}
			}
			ftype := uint32(item.Type())
			name := path + "/" + fname
			inode := fi.Sys().(*syscall.Stat_t).Ino
			if _, ok := inodeSeen[inode]; ok {
				continue // skip inode if we seen it already
			}
			inodeSeen[inode] = struct{}{}
			if ftype&_modeSymlink != 0 {
				continue // skip symlinks
			}
			if ftype&_modeDir != 0 {
				st, _ := fi.Sys().(*syscall.Stat_t)
				if uint64(st.Dev) != rootNodeDeviceID {
					continue // skip dirtargets outside fs boundary
				}
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
