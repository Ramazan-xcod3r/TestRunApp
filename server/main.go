package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	MyUnzip "github.com/Ramazan-xcod3r/go-react-testrun/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/julienschmidt/httprouter"
)

type Todo struct {
	mu          sync.Mutex
	ID          int    `json:"id"`
	Task        string `json:"task"`
	Done        bool   `json:"status"`
	Description string `json:"description"`
	File        string `json:"file"`
}
type Report struct {
	Id       int
	Name     string
	Details  string
	LoggedAt time.Time
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

var todos []Todo
var UsersSPath string = "./UsersStorage/" + "allFiles" + "/"
var UsersSReportPath string = "./UsersStorage/Reports/"

func init() {
	todos = []Todo{}
}

func main() {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", //http://localhost:3000
		AllowMethods: "GET, POST, PUT, DELETE",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	todos := []Todo{}
	app.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("OK!")
	})
	app.Get("/api/todos", func(c *fiber.Ctx) error {
		return c.JSON(todos)
	})
	app.Get("/api/todos/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(401).SendString("Invalid ID")
		}
		report := Report{}
		go func() {
			buf := new(bytes.Buffer)
			todos[id-1].mu.Lock()
			s := strings.Split(todos[id-1].File, ".zip")[0]
			errs := os.Chdir(s)
			if errs != nil {
				log.Println(errs)
			}
			pwd, _ := os.Getwd()
			println(pwd)
			cmd := exec.Command("mvn", "verify", "test")

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Fatal(err)
			}
			if err := cmd.Start(); err != nil {
				log.Fatal(err)
			}
			buf.ReadFrom(stdout)
			out := buf.String()

			re := regexp.MustCompile("(?m)^.[*].*$") //[\r\n]+
			out = re.ReplaceAllString(out, "")
			report = Report{Id: id - 1, Name: todos[id-1].Task, Details: out, LoggedAt: time.Now()}
			content, _ := json.Marshal(report)
			os.Chdir("../../..")
			fname := (UsersSReportPath + strconv.Itoa(report.Id) + "--" + report.LoggedAt.String() + ".json")
			fname = strings.ReplaceAll(fname, ":", "")
			fname = strings.ReplaceAll(fname, " ", "")
			if _, err := os.Stat(UsersSReportPath); os.IsNotExist(err) {
				os.MkdirAll(UsersSReportPath, os.ModePerm)
			}
			f, _ := os.Create(fname)
			defer f.Close()
			n, err := f.Write(content)
			if err != nil {
				fmt.Println(n, err)
			}
			if n, err = f.WriteString("\n"); err != nil {
				fmt.Println(n, err)
			}

			todos[id-1].mu.Unlock()

		}()
		if todos[id-1].Done = true; todos[id-1].Done {
			return c.JSON(fmt.Sprint(todos[id-1].Task, " Task is Running! Don`t forget look at the reports!"))
		}
		return c.SendString("id not found!")
	})
	app.Get("/reports/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		id--
		if err != nil {
			return c.Status(401).SendString("Invalid ID")
		}
		files, err := filepath.Glob(UsersSReportPath + "*.json")
		if err != nil {
			log.Fatal(err)
		}
		var reports []Report
		for _, file := range files {
			f, err := os.Open(file)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			var report Report
			jsonParser := json.NewDecoder(f)
			if err := jsonParser.Decode(&report); err != nil {
				log.Fatal(err)
			}
			if report.Id == id {
				reports = append(reports, report)
			}
		}
		return c.JSON(reports)
	})
	app.Post("/api/todos", func(c *fiber.Ctx) error {
		todo := &Todo{}
		if err := c.BodyParser(todo); err != nil {
			return err
		}
		todo.ID = len(todos) + 1
		fmt.Println("TODOLIST: ", todo)
		//b, _ := c.MultipartForm()
		//fmt.Println("files : ", b.Value["file"])
		a, err := c.FormFile("file")
		if err != nil {
			return err
		}
		fmt.Println((a.Size / 1024), "KB")

		if _, err := os.Stat(UsersSPath); os.IsNotExist(err) {
			os.MkdirAll(UsersSPath, os.ModePerm)
		}
		path := UsersSPath + a.Filename
		todo.File = path
		dst, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
		}
		defer dst.Close()
		println("paths: " + path)
		c.SaveFile(a, path)
		defer MyUnzip.MyUnzip(UsersSPath, a.Filename)
		todos = append(todos, *todo)
		return c.JSON(todos)
	})

	app.Patch("/api/todos/:id/done", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(401).SendString("Invalid ID")
		}
		if todos[id-1].Done = true; todos[id-1].Done {
			return c.JSON(todos)
		}
		return c.SendString("id not found!")
	})
	log.Fatal(app.Listen(":8080"))
}
