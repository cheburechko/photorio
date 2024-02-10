package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Port         string `yaml:"port" env:"PORT" env-default:"9000"`
		TemplateRoot string `yaml:"template_root"`
	}

	App struct {
		Config         *Config
		SearchTemplate *template.Template
	}

	SearchResponse struct {
		Text string
	}
)

func NewApp(c *Config) (*App, error) {
	tmpl, err := template.ParseFiles(filepath.Join(c.TemplateRoot, "search_result.html"))
	if err != nil {
		return nil, err
	}

	return &App{
		SearchTemplate: tmpl,
		Config: c,
	}, nil
}

func readConfig() (*Config, error) {
	var cfg Config

	configPath := flag.String("config", "config.yaml", "Path to config")

	flag.Parse()

	err := cleanenv.ReadConfig(*configPath, &cfg)

	return &cfg, err
}

func main() {
	cfg, err := readConfig()
	if err != nil {
		log.Fatal(err)
		return
	}

	app, err := NewApp(cfg)

	if err != nil {
		log.Fatal(err)
		return
	}

	mux := http.NewServeMux()

	mux.Handle("/", http.HandlerFunc(app.Home))
	mux.Handle("/search", http.HandlerFunc(app.Search))

	addr := fmt.Sprintf(":%s", app.Config.Port)
	log.Printf("Address: %s", addr)

	err = http.ListenAndServe(addr, mux)
	log.Fatal(err)
}

func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(a.Config.TemplateRoot, "index.html"))
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
