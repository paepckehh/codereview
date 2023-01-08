// package codereview ...
package codereview

// import
import (
	"os"
	"runtime"
	"sync"
)

//
// CORE TYPES, VARS, CONTS
//

type statFile struct {
	lineTotal   int
	lineSavings int
	byteSavings int
}

type statFiles struct {
	totalFiles     int
	processedFiles int
}

type (
	files   struct{ totalFiles, processedFiles int }
	savings struct{ linesTotal, sLines, sBytes int }
)

var (
	bg, walk, walkStat sync.WaitGroup
	worker             = runtime.NumCPU()
	channel_statFile   = make(chan statFile, maxTargets)
	channel_statFiles  = make(chan statFiles, maxTargets)
	channel_dir        = make(chan string, chanBuff*25)
	channel_c          = make(chan string, chanBuff)
	channel_go         = make(chan string, chanBuff)
	channel_sh         = make(chan string, chanBuff/4)
	channel_asm        = make(chan string, chanBuff/4)
	channel_make       = make(chan string, chanBuff/4)
	channelFiles       = make(chan files, 1)
	channelSavings     = make(chan savings, 1)
	_goDS              = []byte(_goD) // thanks, for not having immutable constant slices in go!
	_goDSO             = []byte(_goDO)
	_lf                = []byte(_lineFeed)
)

//
// CORE FUNCTIONS
//

// dirWalk ...
func (c *Config) dirWalk() TotalStats {
	// spin up worker
	go func() {
		if c.LangC {
			bg.Add(worker)
			go c.workerC()
		}
		if c.LangGO {
			bg.Add(worker)
			go c.workerGO()
		}
		if c.LangSH {
			bg.Add(worker)
			go c.workerSH()
		}
		if c.LangASM {
			bg.Add(worker)
			go c.workerASM()
		}
		if c.LangMAKE {
			bg.Add(worker)
			go c.workerMake()
		}
	}()
	// spin up stats collection
	walkStat.Add(worker)
	go func() {
		go func() {
			var linesTotal, sLines, sBytes int
			for s := range channel_statFile {
				linesTotal += s.lineTotal
				sLines += s.lineSavings
				sBytes += s.byteSavings
			}
			channelSavings <- savings{linesTotal: linesTotal, sLines: sLines, sBytes: sBytes}
		}()
		go func() {
			var totalFiles, processedFiles int
			for s := range channel_statFiles {
				totalFiles += s.totalFiles
				processedFiles += s.processedFiles
			}
			channelFiles <- files{totalFiles: totalFiles, processedFiles: processedFiles}
		}()
	}()
	// spin up directory walker
	{
		walk.Add(1)
		go c.fastWalker()
	}
	walk.Wait()
	{
		close(channel_dir)
		close(channel_c)
		close(channel_sh)
		close(channel_go)
		close(channel_asm)
		close(channel_make)
	}
	bg.Wait()
	walkStat.Wait()
	{
		close(channel_statFile)
		close(channel_statFiles)
	}
	filesResult := <-channelFiles
	savingsResult := <-channelSavings
	return TotalStats{
		TotalFiles:        filesResult.totalFiles,
		ProcessedFiles:    filesResult.processedFiles,
		TotalLines:        savingsResult.linesTotal,
		TotalSavingsLines: savingsResult.sLines,
		TotalSavingsBytes: savingsResult.sBytes,
	}
}

// finalizeWorker ...
func (c *Config) finalizeWorker(totalLines, lineSavings, byteSavings int, filename string, compact []byte) Result {
	r := Result{}
	switch {
	case c.Inplace:
		if byteSavings > 0 {
			if err := os.WriteFile(filename, compact, 0o644); err != nil {
				errOut("[unable to write] [" + filename + "] [" + err.Error() + "]")
			}
		}
	case c.HashReturn:
		r.Result = slice2Hex(hash(compact))
	case c.HashOut && c.SingleFile:
		out(string(slice2Hex(hash(compact))))
	case c.HashOut && c.Verbose:
		out("[" + string(slice2Hex(hash(compact))) + "] [" + filename + "]")
	case !c.Inplace && c.SingleFile:
		out(string(compact))
	}
	r.TotalLines = totalLines
	r.LineSavings = lineSavings
	r.ByteSavings = byteSavings
	return r
}
