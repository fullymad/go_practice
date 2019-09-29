////
// https://gophercises.com/exercises/cyoa
////

/*
Choose Your Own Adventure is (was?) a series of books intended for children
where as you read you would occasionally be given options about how you want to
proceed. For instance, you might read about a boy walking in a cave when he
stumbles across a dark passage or a ladder leading to an upper level and the
reader will be presented with two options like:

Turn to page 44 to go up the ladder.
Turn to page 87 to venture down the dark passage.

The goal of this exercise is to recreate this experience via a web application
where each page will be a portion of the story, and at the end of every page
the user will be given a series of options to choose from (or be told that they
have reached the end of that particular story arc).
*/

package main

import (
	"errors"
	"log"
	"fmt"
	"strings"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"html/template"
)

const bookFile string = "Book.json"

type Arc struct {
	Title	string
	Story	[]string
	Options []struct {
		Text	string `json:",omitempty"`
		Arc		string `json:",omitempty"`
	}
}

type Book map[string]Arc

// NOTE: The following globals could possibly be made locals and passed to a
// function that dynamically generates the handler function that can thus get
// access to these
var arcsMap Book // Global map of story arc names to story arc
var tmpl *template.Template // Global template

const storyURL = "/story"

const tmplText = `
<!DOCTYPE html>
<html>
<body>
	<h1>Create your own adventure!</h1>
	<h2>{{.Title}}</h2>
	<p>{{range .Story}} {{.}} {{end}}</p>
	<form>{{range .Options}}
		<label>{{.Text}}</label>
		<input type="radio" name="next_arc" value={{.Arc}} required><br>{{end}}
		{{if .Options}}
  		<input type="submit" value="Submit">{{end}}
	</form>
</body>
</html>`

func storyHandler(w http.ResponseWriter, r *http.Request) {
	// Rid of initial slash and final one (if any)
	arcName := (*r).URL.Path[1:]
	arcName = strings.TrimSuffix(arcName, "/")

	// Consider as main story arc if initial URL (including NO storyURL prefix)
	if arcName == "" || arcName == storyURL[1:] {
		arcName = "intro"
	}

	// If next story arc requested, display that
	next_arc := r.FormValue("next_arc")
	if next_arc != "" {
		arcName = next_arc
	}

	arc, ok := arcsMap[arcName]
	if !ok {
		switch arcName {
		case "intro":
			log.Fatal(errors.New("Error finding story arc 'intro'"))
		default:
			http.NotFound(w, r)
		}
	} else {
		err := tmpl.Execute(w, arc)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	book, err := ioutil.ReadFile(bookFile)
	if err != nil {
		log.Fatal(errors.New(fmt.Sprintf("error reading book file %s: ",
			bookFile) + err.Error()))
	}

	// Top level JSON key is arbitrary and represents story arc name
	// Key: Story arc name, Value: JSON string for each story arc 
	var m map[string]*json.RawMessage
	err = json.Unmarshal(book, &m)
	if err != nil {
		log.Fatal(err)
	}

	// Final map containing all the story arcs with arc name as the key
	arcsMap = make(Book, len(m))

	// Parse each value into a story arc and add to map of story arcs
	for arcName, v := range m {
		var arc Arc
		err = json.Unmarshal(*v, &arc)
		if err != nil {
			log.Fatal(err)
		}
		arcsMap[arcName] = arc
	}

	// Create template with web page contents
	tmpl, err = template.New("Story").Parse(tmplText)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", storyHandler)
	http.HandleFunc(storyURL, storyHandler)

	fmt.Println("Starting story server at 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
