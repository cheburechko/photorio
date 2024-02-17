package internal

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	Config              *Config
	ElasticsearchClient *elasticsearch.TypedClient
	Postgres            *pgxpool.Pool
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Message string `json:"message"`
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

	var hashedPassword string
	err := a.Postgres.QueryRow(c, "select password from users where username = $1", login).Scan(&hashedPassword)

	if err != nil {
		if err == pgx.ErrNoRows {
			AbortWithHTML(c, http.StatusUnauthorized, fmt.Errorf("bad creds"))
		} else {
			AbortWithHTML(c, http.StatusInternalServerError, err)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

	if err != nil {
		AbortWithHTML(c, http.StatusUnauthorized, fmt.Errorf("bad creds"))
		return
	}

	err = SetAdmin(c)

	if err != nil {
		AbortWithHTML(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
	c.Header("hx-redirect", "/admin/")
}

func (a *App) CreateUser(c *gin.Context) {
	var user User
	err := c.BindJSON(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, Response{Message: "bad json"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}

	_, err = a.Postgres.Exec(c, "insert into users(username, password) values($1, $2);", user.Username, hashedPassword)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, Response{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Message: "registered"})
}

func AbortWithHTML(c *gin.Context, code int, err error) {
	c.Abort()
	c.Error(err)
	c.HTML(code, "error.html", err.Error())
}
