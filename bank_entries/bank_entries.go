package main

import (
	"fmt"
//	"io"
	"log"
	"os"
	"bufio"
	"strings"
	"strconv"
)

const MIN_FIELDS int = 6
const DELIMITER string = ";"

// Convert text files entries such as the following to a format suitable
// for import to spreadsheet
//
// Sep 16, 02	071-058077-220			 7,871.14			41,319.75
// Sep 21, 02	CHECK DEPOSIT 956097			10,110.00	51,429.75
// Delimiter will be ';'

func main() {
	var err error
	var lastBalance float64 = 0.0
	// Read input log records and print relevant fields
	scanner := bufio.NewScanner(os.Stdin)

	// Print column headings
	fmt.Printf("Date%sDescription%sDebit%sCredit%sBalance\n",
		DELIMITER, DELIMITER, DELIMITER, DELIMITER)
	
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		numFields := len(fields)
		if numFields < MIN_FIELDS {
			log.Fatalf(
				"Input record should have at least %d fields, found %d",
				MIN_FIELDS, numFields)
		}
		// Date fields like 'Sep 9, 02'
		fmt.Printf("%s %s %s%s", fields[0], fields[1], fields[2],
			DELIMITER)
		
		// Transaction description
		for i := 3; i < (numFields - 2); i++ {
			if i != 3 { // No space before the first of tokens
				fmt.Printf(" ")
			}
			fmt.Printf("%s", fields[i])
		}
		fmt.Printf(DELIMITER)

		// Penultimate field is transaction amount (debit or credit)
		var transAmount float64 = 0.0
		transField := fields[numFields - 2]
		transField = strings.Replace(transField, ",", "", -1)
		transAmount, err = strconv.ParseFloat(transField, 32);
		if err != nil {
			log.Fatalf("Error converting %s to decimal number",
				transField)
		}

		// Last field is new balance
		var newBalance float64 = 0.0
		transField = fields[numFields - 1]
		transField = strings.Replace(transField, ",", "", -1)
		newBalance, err = strconv.ParseFloat(transField, 32);
		if err != nil {
			log.Fatalf("Error converting %s to decimal number",
				transField)
		}

		// Tranaction amount should be in Debit or Credit column as
		// appropriate (with the other one set to a blank string)
		if newBalance >= lastBalance { // Credit (comes after)
			fmt.Printf("%s%.2f%s%.2f\n", DELIMITER, transAmount,
				DELIMITER, newBalance)
		} else { // Debit (comes first)
			fmt.Printf("%.2f%s%s%.2f\n", transAmount, DELIMITER,
				DELIMITER, newBalance)
		}
		lastBalance = newBalance
	}
	if err = scanner.Err(); err != nil {
		fmt.Println(os.Stderr, "reading entries from input:", err) 
	}
} // end main
