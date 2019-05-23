package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/gorilla/mux"
)

type ContactDetails struct {
	Email   string
	Subject string
	Message string
}

type Todo struct {
	Title string
	Done  bool
}

type TodoPageData struct {
	PageTile string
	Todos    []Todo
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

func Logging() Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			defer func() { log.Println(r.URL.Path, time.Since(start)) }()
			f(w, r)
		}
	}
}

func Method(m string) Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != m {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			log.Println(r.Method)
			f(w, r)
		}
	}
}

func Chain(f http.HandlerFunc, middleware ...Middleware) http.HandlerFunc {
	for _, m := range middleware {
		f = m(f)
	}
	return f
}

func ReadBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	title := vars["title"]
	page := vars["page"]

	fmt.Fprintf(w, "you are request title: %s page: %s", title, page)
}

func main() {
	r := mux.NewRouter()
	tmpl := template.Must(template.ParseFiles("layout.html"))
	tmplfrom := template.Must(template.ParseFiles("from.html"))

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := TodoPageData{
			PageTile: "my Todo List",
			Todos: []Todo{
				{Title: "Task 1", Done: false},
				{Title: "Task 2", Done: true},
				{Title: "Task 2", Done: false},
			},
		}

		tmpl.Execute(w, data)
	}).Methods("GET")
	r.HandleFunc("/from", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			tmplfrom.Execute(w, nil)
			return
		}

		details := ContactDetails{
			Email:   r.FormValue("email"),
			Subject: r.FormValue("subject"),
			Message: r.FormValue("message"),
		}
		_ = details
		tmplfrom.Execute(w, struct{ Success bool }{true})
	}).Methods("GET", "POST")
	r.HandleFunc("/books/{title}/page/{page}", Chain(ReadBook, Method("GET"), Logging()))

	http.ListenAndServe(":8080", r)
}
