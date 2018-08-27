/******************************************************************************
https://www.hackerrank.com/challenges/contacts
******************************************************************************/

package main

import (
	"fmt"
)

type Trie struct {
	letter       byte
	next_letters []Trie
	count        uint
}

func addName(trie *Trie, name string, index int) {
	var length, last_elem int
	var ch byte
	var found_letter bool
	var next_trie *Trie
	
	length = len(name)
	ch = name[index]
	
	for i, _ := range trie.next_letters {
		next_trie = &trie.next_letters[i] // Get address so we can modify
		if (next_trie.letter == ch) {
			next_trie.count += 1
			
			if (index < length - 1) {
				addName(next_trie, name, index + 1)
			}
			
			found_letter = true
			break
		}
	}
	
	if (! found_letter) {
		trie.next_letters = append(trie.next_letters, Trie{ch, nil, 1})
		
		if (index < length - 1) {
			last_elem = len(trie.next_letters) - 1
			addName(&trie.next_letters[last_elem], name, index + 1)
		}
	}
	
	return
} // end addName

func findPartial(trie *Trie, partial string, index int) uint {
	var length int
	var count uint
	var ch byte
	
	length = len(partial)
	ch = partial[index]
	
	for _, t := range trie.next_letters {
		if (t.letter == ch) {
			if (index == length -1) {
				count = t.count;
			} else {
				count = findPartial(&t, partial, index + 1)
			}
			break
		}
	}
	
	return count
} // end findPartial

func main() {
	trie := Trie{};
	var ops int
	var count uint
	var operation, name string

	fmt.Scanf("%d", &ops)

	for i := 0; i < ops; i++ {
		fmt.Scanf("%s", &operation)
		fmt.Scanf("%s", &name)

		if operation == "add" {
			addName(&trie, name, 0)
			//fmt.Println("main after add: Data", trie)
		} else if operation == "find" {
			//fmt.Println("main before find: Data", trie)
			count = findPartial(&trie, name, 0)
			fmt.Printf("%d\n", count)
		} else {
			fmt.Printf("Invalid operation: %s\n", operation)
			return
		}
	}
} // end main