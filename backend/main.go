package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

type (
	App struct {
		SearchTemplate *template.Template
	}

	SearchResponse struct {
		Text string
	}
)

func NewApp() (*App, error) {
	tmplText, err := os.ReadFile("templates/search_result.html")
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("search_results").Parse(string(tmplText))
	if err != nil {
		return nil, err
	}

	return &App{
		SearchTemplate: tmpl,
	}, nil
}

func main() {
	app, err := NewApp()

	if err != nil {
		log.Fatal(err)
		return
	}

	mux := http.NewServeMux()

	mux.Handle("/", http.HandlerFunc(app.Home))
	mux.Handle("/search", http.HandlerFunc(app.Search))

	err = http.ListenAndServe(":9000", mux)
	log.Fatal(err)
}

func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func (a *App) Search(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.PostFormValue("prompt"))

	if query == "" {
		_, _ = w.Write([]byte("Please enter something"))
		return
	}

	resp := SearchResponse{
		Text: query,
	}

	var buf bytes.Buffer

	err := a.SearchTemplate.Execute(&buf, resp)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}

	_, err = buf.WriteTo(w)
}
