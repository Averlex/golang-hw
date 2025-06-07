// main contains a function for copying a file from a source path to
// a destination path with optional offset and limit.
package main

import (
	"flag"
	"log"
)

var (
	from, to      string
	limit, offset int64
)

func init() {
	flag.StringVar(&from, "from", "", "file to read from")
	flag.StringVar(&to, "to", "", "file to write to")
	flag.Int64Var(&limit, "limit", 0, "limit of bytes to copy")
	flag.Int64Var(&offset, "offset", 0, "offset in input file")
}

func main() {
	flag.Parse()
	res := Copy(from, to, offset, limit)
	if res != nil {
		log.Fatal(res.Error())
	}
}
