package main

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"

	"github.com/gorilla/mux"
)

func main() {
	t := reflect.TypeOf(http.Request{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fmt.Printf("%s: offset=%d, size=%d\n", field.Name, field.Offset, field.Type.Size())
	}
	fmt.Printf("\n")
	t = reflect.TypeOf(url.URL{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fmt.Printf("%s: offset=%d, size=%d\n", field.Name, field.Offset, field.Type.Size())
	}

	// Print the memory address of the GetPage function
	fmt.Printf("GetPage: %p\n", GetPage)

	r := mux.NewRouter()

	r.HandleFunc("/books/{title}/page/{page}", GetPage)

	http.ListenAndServe(":80", r)
}

func GetPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	title := vars["title"]
	page := vars["page"]

	w.Header().Set("Access-Control-Allow-Origin", "*")

	fmt.Fprintf(w, "You've requested the book: %s on page %s, req @ (%p, %p, %p)\n", title, page, r, &r.Method, r.URL)
}
