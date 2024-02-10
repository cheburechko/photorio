package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/gin-gonic/gin"
	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Port          string               `yaml:"port" env:"PORT" env-default:"9000"`
		TemplateGLOB  string               `yaml:"template_glob"`
		Elasticsearch elasticsearch.Config `yaml:"elasticsearch"`
	}

	App struct {
		Config              *Config
		ElasticsearchClient *elasticsearch.TypedClient
	}

	SearchItem struct {
		URL     string `json:"url"`
		Caption string `json:"caption"`
	}

	SearchResponse struct {
		Items []*SearchItem
	}
)

func NewApp(c *Config) (*App, error) {
	client, err := elasticsearch.NewTypedClient(c.Elasticsearch)
	if err != nil {
		return nil, err
	}

	return &App{
		Config:              c,
		ElasticsearchClient: client,
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
		slog.Error("Failed to read config", slog.String("error", err.Error()))
		return
	}

	app, err := NewApp(cfg)

	if err != nil {
		slog.Error("Failed to create app", slog.String("error", err.Error()))
		return
	}

	router := gin.Default()
	router.LoadHTMLGlob(app.Config.TemplateGLOB)
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/", app.Home)
	router.POST("/search", app.Search)

	addr := fmt.Sprintf(":%s", app.Config.Port)

	err = router.Run(addr)
	slog.Error("Finishing", slog.String("error", err.Error()))
}

func (a *App) Home(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func (a *App) Search(c *gin.Context) {
	query := strings.ToLower(c.Request.PostFormValue("prompt"))

	result, err := a.ElasticsearchClient.Search().
		Index("captions").
		Request(&search.Request{
			Query: &types.Query{
				Match: map[string]types.MatchQuery{
					"caption": {Query: query},
				},
			},
		}).
		Do(c)

	if err != nil {
		AbortWithHTML(c, err)		
		return
	}

	resp := SearchResponse{
		Items: make([]*SearchItem, result.Hits.Total.Value),
	}

	for i, item := range result.Hits.Hits {
		var parsed SearchItem
		err = json.Unmarshal(item.Source_, &parsed)
		if err != nil {
			AbortWithHTML(c, err)

			marshalled, _ := item.Source_.MarshalJSON()
			slog.Error("failed to unmarshall json", slog.String("json", string(marshalled)))
			return
		}
		resp.Items[i] = &parsed
	}

	c.HTML(http.StatusOK, "search_result.html", resp)
}

func AbortWithHTML(c *gin.Context, err error) {
	c.Abort()
	c.Error(err)
	c.HTML(http.StatusInternalServerError, "error.html", err.Error())
}
