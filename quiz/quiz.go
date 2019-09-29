////
// https://gophercises.com/exercises/quiz
////

/*
Read in a quiz provided via a CSV file (<question>,<answer>)and give the quiz
to a user keeping track of how many questions they get right and how many they
get incorrect. Regardless of whether the answer is correct or wrong the next
question should be asked immediately afterwards.
Adapt the program from part 1 to add a timer. The default time limit should be
30 seconds, but should also be customizable via a flag.
Your quiz should stop as soon as the time limit has exceeded. That is, you
shouldn't wait for the user to answer one final question but should ideally
stop the quiz entirely even if you are currently waiting on an answer from the
end user.
*/

package main

import (
	"errors"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"bufio"
	"log"
	"os"
	"strings"
	"time"
)

const (
	numFields int = 2
)

var (
	quizFile		string
	timeLimit		uint
)

func init() {
	flag.StringVar(&quizFile, "csv", "problems.csv",
		"a csv file in the format of 'question,answer'")
	flag.UintVar(&timeLimit, "limit", 30,
		"the time limit for the quiz in seconds")
	flag.Parse()
}

func main() {

	var correct = 0
	var total = 0

	f, err := os.Open(quizFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fmt.Printf("Please press Enter when ready to start the quiz\n");
	fmt.Printf("Once you start, you will have %d seconds " +
		"to complete the quiz\n", timeLimit);

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if err = scanner.Err(); err != nil {
		log.Fatal(errors.New("starting quiz: " + err.Error()))
	}

	done := make(chan bool)

	// Goroutine to administer questions and signal when complete
	go func() {
		r := csv.NewReader(f)
		r.FieldsPerRecord = 2
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			} else {
				// Increment count of questions and print next question
				total += 1;
				fmt.Printf("%s: ", record[0]);
	
				// Read answer and check for correctness
				scanner.Scan()
				if err = scanner.Err(); err != nil {
					log.Fatal(errors.New("getting answer: " + err.Error()))
				}
	
				answer := scanner.Text();
				if (strings.TrimSpace(answer) == record[1]) {
					correct += 1;
				}
			}
		}

		done <- true
	}()

	timer := time.NewTimer(time.Duration(timeLimit) * time.Second)

	select {
	case <-done:
		fmt.Printf("Your score: %d out of %d\n", correct, total);
	case <-timer.C:
		fmt.Printf("\nExceeded time limit of %d seconds " +
			"to complete the quiz\n", timeLimit);
	}

} // end main
