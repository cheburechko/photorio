package internal

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Config              *Config
	ElasticsearchClient *elasticsearch.TypedClient
	Postgres            *pgxpool.Pool
}

func NewApp(c *Config) (*App, error) {
	client, err := elasticsearch.NewTypedClient(c.Elasticsearch)
	if err != nil {
		return nil, err
	}

	psqlUrl := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", c.Postgres.User, c.Postgres.Password, c.Postgres.Host, c.Postgres.Port, c.Postgres.Database)
	dbpool, err := pgxpool.New(context.Background(), psqlUrl)
	if err != nil {
		return nil, err
	}

	return &App{
		Config:              c,
		ElasticsearchClient: client,
		Postgres:            dbpool,
	}, nil
}

func (a *App) Home(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", CheckAdmin(c))
}

func (a *App) SignIn(c *gin.Context) {
	if CheckAdmin(c) {
		c.Redirect(http.StatusFound, "/admin/")
		return
	}

	c.HTML(http.StatusOK, "sign_in.html", nil)
}

func (a *App) Admin(c *gin.Context) {
	c.HTML(http.StatusOK, "admin.html", nil)
}

func (a *App) Search(c *gin.Context) {
	query := strings.ToLower(c.Request.PostFormValue("prompt"))

	resp, err := Search(query, a.ElasticsearchClient, c)

	if err != nil {
		AbortWithHTML(c, http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "search_result.html", resp)
}

func (a *App) SubmitSignIn(c *gin.Context) {
	login := c.Request.PostFormValue("login")
	password := c.Request.PostFormValue("password")

	if login != "login" || password != "password" {
		AbortWithHTML(c, http.StatusUnauthorized, fmt.Errorf("Bad creds"))
		return
	}

	err := SetAdmin(c)

	if err != nil {
		AbortWithHTML(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, "/admin/")
	return
}

func AbortWithHTML(c *gin.Context, code int, err error) {
	c.Abort()
	c.Error(err)
	c.HTML(code, "error.html", err.Error())
}
