package codereview

// Config ...
type Config struct {
	Path       string
	Inplace    bool
	SingleFile bool
	Exclude    []string
	HashOut    bool
	HashReturn bool
	Verbose    bool
	Silent     bool
	SkipHidden bool
	Debug      bool
	LangC      bool
	LangGO     bool
	LangSH     bool
	LangASM    bool
	LangMAKE   bool
}

// Result holds stats and the final object result
type Result struct {
	TotalLines, LineSavings, ByteSavings int
	Result                               []byte
	Found                                bool
}

// TotalStats ...
type TotalStats struct {
	TotalFiles, ProcessedFiles, TotalLines, TotalSavingsLines, TotalSavingsBytes int
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		LangC:    true,
		LangGO:   true,
		LangSH:   true,
		LangASM:  true,
		LangMAKE: true,
	}
}

// NewConfigBatch  ...
func NewConfigBatch() *Config {
	return &Config{
		HashReturn: true,
		Silent:     true,
		LangC:      true,
		LangGO:     true,
		LangSH:     true,
		LangASM:    true,
		LangMAKE:   true,
	}
}

// CodeHash ...
func CodeHash(filename string) {
	c := NewConfig()
	c.Path = filename
	c.SingleFile = true
	c.HashOut = true
	_ = c.ParseFile()
}

// WalkDir ...
func (c *Config) WalkDir() TotalStats {
	if c.Silent {
		silent = true
	}
	return c.dirWalk()
}

// ParserC ...
func (c *Config) ParserC() Result {
	if c.Silent {
		silent = true
	}
	return c.parseC(c.Path)
}

// ParserGO ...
func (c *Config) ParserGO() Result {
	if c.Silent {
		silent = true
	}
	return c.parseGO(c.Path)
}

// ParserSH ...
func (c *Config) ParserSH() Result {
	if c.Silent {
		silent = true
	}
	return c.parseSH(c.Path)
}

// ParserASM ...
func (c *Config) ParserASM() Result {
	if c.Silent {
		silent = true
	}
	return c.parseASM(c.Path)
}

// ParserMake ...
func (c *Config) ParserMake() Result {
	if c.Silent {
		silent = true
	}
	return c.parseMake(c.Path)
}

// ParseFileSheBang ...
func (c *Config) ParseFileSheBang() Result {
	if c.Silent {
		silent = true
	}
	r := c.ParseFile()
	if r.Found {
		if b := shebang(c.Path); len(b) == 9 {
			if b == "#!/bin/sh" {
				r = c.ParserSH()
				r.Found = true
				return r
			}
		}
		r.Result = []byte("file type not supported [" + c.Path + "]")
		return r
	}
	return r
}

// ParseFile ...
func (c *Config) ParseFile() Result {
	if c.Silent {
		silent = true
	}
	l := len(c.Path)
	if l > 2 {
		if c.Path[l-2:] == ".c" || c.Path[l-2:] == ".h" {
			r := c.ParserC()
			r.Found = true
			return r
		}
		if c.Path[l-2:] == ".s" || c.Path[l-2:] == ".S" {
			r := c.ParserASM()
			r.Found = true
			return r
		}
	}
	if l > 3 {
		switch {
		case c.Path[l-3:] == ".go":
			r := c.ParserGO()
			r.Found = true
			return r
		case c.Path[l-3:] == ".sh":
			r := c.ParserSH()
			r.Found = true
			return r
		case c.Path[l-3:] == ".cc" || c.Path[l-3:] == ".hh":
			r := c.ParserC()
			r.Found = true
			return r
		}
	}
	if l > 4 {
		switch {
		case c.Path[l-4:] == ".cpp":
			r := c.ParserC()
			r.Found = true
			return r
		default:
		}
	}
	if l > 7 {
		switch {
		case c.Path == "Makefile":
			r := c.ParserMake()
			r.Found = true
			return r
		default:
		}
	}
	r := Result{}
	r.Result = []byte("file type not supported [" + c.Path + "]")
	return r
}
