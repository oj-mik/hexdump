package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

const version = "1.0.1"
const date = "2020-06-07"

var hex, help, noByteNr, noAsc, ver bool
var lnFmt = "%08d:   "

func init() {
	flag.BoolVar(&hex, "x", false, "print line numbers as hexadecimal numbers")
	flag.BoolVar(&noByteNr, "b", false, "don't print byte numbers")
	flag.BoolVar(&noAsc, "a", false, "don't print ASCII values")
	flag.BoolVar(&help, "h", false, "print help information")
	flag.BoolVar(&ver, "v", false, "print version number and exit")
}

func main() {
	flag.Parse()

	if help {
		fmt.Fprint(os.Stderr,
			"hexdump\n",
			"  Reads bytestream from stdin or input file and outputs byte in human readable format.\n",
			"\n",
			"  Usage is:\n",
			"  hexdump [options] [input file]\n",
			"\n",
		)
		flag.PrintDefaults()
		return
	}

	if ver {
		fmt.Fprintf(os.Stderr, "hexdump version %s, dated %s", version, date)
		return
	}

	var input io.ReadCloser
	var err error

	switch flag.NArg() {
	case 0:
		input = os.Stdin
	case 1:
		input, err = os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open file: %s", err)
			return
		}
		defer input.Close()
	default:
		fmt.Fprintf(os.Stderr, "Too many arguments, missing quotes around file name containing space?")
		return
	}

	if hex {
		lnFmt = "%08x:   "
	}

	b := make([]byte, 16384, 16384)
	prevb := make([]byte, 0, 16)
	var preveq bool
	var i, bn int

	for err != io.EOF {
		i, err = input.Read(b)
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "Error reading file: %s", err)
			return
		}

		sb := sliceBuf(b[:i])
		for n := range sb {
			if sliceEqual(sb[n], prevb) && !noByteNr {
				if !preveq {
					fmt.Println("   *")
					preveq = true
				}
			} else {
				preveq = false
				prevb = prevb[:len(sb[n])]
				copy(prevb, sb[n])
				slicePrint(bn+n*16, sb[n])
			}
		}
		bn += i
	}
}

func sliceBuf(b []byte) [][]byte {
	l := len(b) / 16
	if len(b)%16 != 0 {
		l += 1
	}
	r := make([][]byte, l)
	for n := range r {
		if len(b[n*16:]) > 16 {
			r[n] = b[n*16 : n*16+16]
		} else {
			r[n] = b[n*16:]
		}
	}
	return r
}

func slicePrint(ln int, b []byte) {
	if len(b) > 16 {
		b = b[:16]
	}
	if !noByteNr {
		fmt.Printf(lnFmt, ln)
	}
	for n := 0; n < 16; n++ {
		if n < len(b) {
			fmt.Printf("%02x", b[n])
		} else {
			fmt.Printf("  ")
		}

		switch n {
		case 7: // extra space between 8 first and 8 last bytes
			fmt.Printf("  ")
		case 15: // no space after last byte
		default:
			fmt.Printf(" ")
		}
	}
	if !noAsc {
		fmt.Printf("   |%s|", sliceToASCII(b))
	}
	fmt.Print("\n")
}

func sliceToASCII(b []byte) string {
	r := make([]byte, len(b), len(b))
	for n := range b {
		if b[n] < 32 || b[n] > 126 {
			r[n] = '.'
		} else {
			r[n] = b[n]
		}
	}
	return string(r)
}

func sliceEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for n := range a {
		if a[n] != b[n] {
			return false
		}
	}
	return true
}
