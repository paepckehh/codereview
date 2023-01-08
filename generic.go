// package codereview ...
package codereview

// import
import (
	"github.com/zeebo/blake3"
)

//
// LITTLE HELPER FUNCTIONS
//

const (
	maxTargets   = 4
	chanBuff     = 1000
	_tab         = '\t'
	_linefeed    = '\n'
	_lineFeed    = "\n"
	_whitespace  = ' '
	_singleqoute = '\''
	_doubleqoute = '"'
	_asterix     = '*'
	_slashfwd    = '/'
	_slashbwd    = '\\'
	_hashmark    = '#'
	_cbracketON  = '{'
	_cbracketOFF = '}'
	_dollar      = '$'
	_hex         = "0123456789abcdef"
	_go          = "//go"  // compat hack, go compiler 'directive' hidout as 'used' by 1.19+ packges now)
	_goD         = "//go:" // go compiler directive
	_goDO        = "// +b" // outdated go asm compiler directive workaround
)

// hash ...
func hash(in []byte) []byte {
	h := blake3.Sum256(in)
	return h[:]
}

// slice2Hex ...
func slice2Hex(in []byte) []byte {
	r := make([]byte, len(in)*2)
	for i, v := range in {
		r[i*2] = _hex[v>>4]
		r[i*2+1] = _hex[v&0x0f]
	}
	return r
}

// removeEmptyLines
func removeEmptyLines(in []byte) ([]byte, int) {
	size, void, removed, buff, clean := len(in), true, 0, []byte{}, []byte{}
	for i := 0; i < size; i++ {
		switch in[i] {
		case _whitespace, _tab:
			if void {
				buff = append(buff, in[i])
				continue
			}
			clean = append(clean, in[i])
		case _linefeed:
			if void {
				buff = []byte{}
				removed++
				continue
			}
			void = true
			clean = append(clean, _linefeed)
		default:
			if void {
				clean = append(clean, buff...)
				buff = []byte{}
			}
			clean = append(clean, in[i])
			void = false
		}
	}
	return clean, removed
}

// removeEmptyGoComments ...
func removeEmptyGoComments(in []byte) ([]byte, int, int) {
	// directive = false
	size, void, removed, buff, clean := len(in), true, 0, []byte{}, []byte{}
	for i := 0; i < size; i++ {
		switch in[i] {
		case _slashfwd:
			if i+1 < size && in[i+1] == _slashfwd && in[i+2] == _linefeed {
				if !void {
					clean = append(clean, _linefeed)
				}
				void, buff = true, nil
				i += 2
				removed++
				continue
			}
			clean = append(clean, _slashfwd)
		case _whitespace, _tab:
			if void {
				buff = append(buff, in[i])
				continue
			}
			clean = append(clean, in[i])
		case _linefeed:
			if void {
				buff = []byte{}
				removed++
				continue
			}
			clean, void = append(clean, _linefeed), true
		default:
			if void {
				clean, buff = append(clean, buff...), nil
			}
			clean, void = append(clean, in[i]), false
		}
	}
	return clean, len(clean), removed
}
