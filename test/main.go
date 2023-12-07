package main

import (
	"fmt"
	"io"
	"os"
)

const FNAME = "test.txt"

func printFileInfo(fp *os.File) {
	info, err := fp.Stat()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(info.Size())
	bytes, err := io.ReadAll(fp)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("(%s)\n", string(bytes))
}

func logErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	fp, err := os.OpenFile(FNAME, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	_, err = fp.WriteString("One two three")
	logErr(err)
	err = fp.Truncate(10)
	logErr(err)
	_, err = fp.Seek(10, 0)
	logErr(err)

	// _, err = fp.WriteString("Four five siz")
	// logErr(err)
	_, err = fp.Seek(0, 0)
	logErr(err)

	printFileInfo(fp)
}
