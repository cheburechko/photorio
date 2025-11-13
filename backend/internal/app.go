package internal

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	Config              *AppConfig
	ElasticsearchClient *elasticsearch.TypedClient
	Postgres            *pgxpool.Pool
	Upgrader            *websocket.Upgrader
	Templates           *template.Template
	KafkaProducer       *Producer
}

func NewApp(c *AppConfig) (*App, error) {
	client, err := elasticsearch.NewTypedClient(c.Elasticsearch)
	if err != nil {
		return nil, err
	}

	dbpool, err := pgxpool.New(context.Background(), c.Postgres.ConnectionUrl)
	if err != nil {
		return nil, err
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	templates, err := template.ParseGlob(c.TemplateGLOB)
	if err != nil {
		return nil, err
	}

	producer, err := NewProducer(c.Kafka)
	if err != nil {
		return nil, err
	}

	return &App{
		Config:              c,
		ElasticsearchClient: client,
		Postgres:            dbpool,
		Upgrader:            &upgrader,
		Templates:           templates,
		KafkaProducer:       producer,
	}, nil
}

func (a *App) Close() {
	a.Postgres.Close()
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
		AbortWithInternalError(c, err)
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
			AbortWithInternalError(c, err)
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
		AbortWithInternalError(c, err)
		return
	}

	c.Status(http.StatusOK)
	c.Header("hx-redirect", "/admin/")
}

func (a *App) SubmitCaptionTask(c *gin.Context) {
	prompt := c.Request.PostFormValue("prompt")

	var id int

	err := a.Postgres.QueryRow(c, "insert into caption_tasks(prompt) values($1) on conflict do nothing returning id", prompt).Scan(&id)
	if err == pgx.ErrNoRows {
		AbortWithHTML(c, http.StatusBadRequest, fmt.Errorf("task exists"))
		return
	} else if err != nil {
		AbortWithInternalError(c, err)
		return
	}

	task := CaptionTaskEvent{
		Id:     id,
		Prompt: prompt,
	}

	err = a.KafkaProducer.SendTaskEvent(&task)

	if err != nil {
		AbortWithInternalError(c, err)
		return
	}

	c.HTML(http.StatusOK, "submit_caption_task_success.html", nil)
}

func (a *App) CaptionTasks(c *gin.Context) {
	db, err := a.Postgres.Acquire(c)
	if err != nil {
		AbortWithInternalError(c, err)
		return
	}
	defer db.Release()

	_, err = db.Exec(c, "listen caption_task_status_channel")
	if err != nil {
		AbortWithInternalError(c, err)
		return
	}

	conn, err := a.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		AbortWithInternalError(c, err)
		return
	}
	defer conn.Close()

	for {
		rows, err := db.Query(c, "select prompt, status, total_subtasks, completed_subtasks, created_at from caption_tasks order by created_at;")
		if err != nil {
			a.HandleWebsocketError(c, "Failed to execute query", conn, err)
			return
		}

		var tasks []CaptionTask

		for rows.Next() {
			var task CaptionTask
			err := rows.Scan(&task.Prompt, &task.Status, &task.TotalSubtasks, &task.CompletedSubtasks, &task.CreatedAt)
			task.PercentComplete = 0
			if task.TotalSubtasks > 0 {
				task.PercentComplete = 100 * task.CompletedSubtasks / task.TotalSubtasks
			}
			if err != nil {
				a.HandleWebsocketError(c, "Failed to scan get row from table", conn, err)
				return
			}
			tasks = append(tasks, task)
		}

		template, err := a.RenderTemplate(c, "caption_tasks.html", tasks)
		if err != nil {
			a.HandleWebsocketError(c, "Failed to render template", conn, err)
			return
		}

		err = conn.WriteMessage(websocket.TextMessage, template)
		if err != nil {
			a.HandleWebsocketError(c, "Failed to write to ws", conn, err)
			return
		}

		_, err = db.Conn().WaitForNotification(c)
		if err != nil {
			a.HandleWebsocketError(c, "Failed to get notification", conn, err)
			return
		}
	}
}

func (a *App) HandleWebsocketError(c *gin.Context, msg string, conn *websocket.Conn, err error) {
	slog.Error(msg, slog.String("error", err.Error()))
	template, _ := a.RenderTemplate(c, "error.html", err)

	conn.WriteMessage(websocket.CloseMessage, template)
}

func (a *App) RenderTemplate(c *gin.Context, name string, data any) ([]byte, error) {
	var buffer bytes.Buffer
	err := a.Templates.ExecuteTemplate(&buffer, name, data)
	return buffer.Bytes(), err
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

func AbortWithInternalError(c *gin.Context, err error) {
	AbortWithHTML(c, http.StatusInternalServerError, err)
}
