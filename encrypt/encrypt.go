// Solution to https://www.hackerrank.com/challenges/encryption/problem

package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

// Complete the encryption function below.
func encryption(s string) string {
	var strBuilder strings.Builder

	// Remove all spaces in the string
	noSpaceStr := strings.Replace(s, " ", "", -1)

	// Get number of rows and columns based on string length, keeping
	// rows * column smallest that can accommodate the string
	length := len(noSpaceStr)
	sqrt := math.Sqrt(float64(length))
	rows := int(math.Floor(sqrt))
	columns := int(math.Ceil(sqrt))

	// If not perfect square, bump up rows if necessary
	if (rows != columns) && (rows*columns < length) {
		rows++
	}

	fmt.Printf("Length: %d, Rows: %d, Columns: %d\n", length,
		rows, columns)

	// Allocate space, covering space character after every column but
	// the last one
	var newStrLen = length + (columns - 1)
	strBuilder.Grow(newStrLen)

	// Treating the characters in the given string to be sequenced as
	// row by row, process 0th character in all rows first to get the
	// first column of characters, then 1st character and so on
	for c := 0; c < columns; c++ {
		for r := 0; r < rows; r++ {
			// Last source row need not be full
			index := (r * columns) + c
			if (index < length) {
				_ = strBuilder.WriteByte(noSpaceStr[index])
			}
		}
		if c != (columns - 1) {
			_ = strBuilder.WriteByte(' ')
		}
	}

	return strBuilder.String()
}

func main() {
	reader := bufio.NewReaderSize(os.Stdin, 1024*1024)

//	stdout, err := os.Create(os.Getenv("OUTPUT_PATH"))
//	checkError(err)
//
//	defer stdout.Close()
//
//	writer := bufio.NewWriterSize(stdout, 1024*1024)
	writer := bufio.NewWriterSize(os.Stdout, 1024*1024)

	s := readLine(reader)

	result := encryption(s)

	fmt.Fprintf(writer, "%s\n", result)

	writer.Flush()
}

func readLine(reader *bufio.Reader) string {
	str, _, err := reader.ReadLine()
	if err == io.EOF {
		return ""
	}

	return strings.TrimRight(string(str), "\r\n")
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
