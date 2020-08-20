package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/benpate/derp"
	"github.com/benpate/html"
	"github.com/benpate/remote"
	"github.com/labstack/echo/v4"
)

type formatFunc func(interface{}) string

func main() {

	rand.Seed(time.Now().UnixNano())

	/// Load configuration file
	var data map[string][]interface{}

	dataBytes, err := ioutil.ReadFile("./static/data.json")

	if err != nil {
		panic(err.Error())
	}

	if err := json.Unmarshal(dataBytes, &data); err != nil {
		panic("Could not unmarshal data: " + err.Error())
	}

	/// Configure Web Server

	e := echo.New()

	e.Static("/", "static")

	// JSON Event Streams
	e.GET("/posts/sse.json", handleStream(makeStream(data["posts"], jsonFormatFunc)))
	e.GET("/comments/sse.json", handleStream(makeStream(data["comments"], jsonFormatFunc)))
	e.GET("/albums/sse.json", handleStream(makeStream(data["albums"], jsonFormatFunc)))
	e.GET("/todos/sse.json", handleStream(makeStream(data["todos"], jsonFormatFunc)))
	e.GET("/users/sse.json", handleStream(makeStream(data["users"], jsonFormatFunc)))

	// HTML Event Streams (with HTMX extension tags)
	e.GET("/posts/sse.html", handleStream(makeStream(data["posts"], postTemplate())))
	e.GET("/comments/sse.html", handleStream(makeStream(data["comments"], commentTemplate())))
	e.GET("/albums/sse.html", handleStream(makeStream(data["albums"], albumTemplate())))
	e.GET("/todos/sse.html", handleStream(makeStream(data["todos"], todoTemplate())))
	e.GET("/users/sse.html", handleStream(makeStream(data["users"], userTemplate())))

	e.GET("/htmx", func(ctx echo.Context) error {
		var result string

		if err := remote.Get(ctx.QueryParam("src")).Response(&result, &result).Send(); err != nil {
			return err
		}

		ctx.Request().Header.Set("mime-type", "text/javascript")
		return ctx.String(200, result)
	})

	e.GET("/page/:number", func(ctx echo.Context) error {

		pageNumber, err := strconv.Atoi(ctx.Param("number"))

		if err != nil {
			pageNumber = 1
		}

		thisPage := strconv.Itoa(pageNumber)
		nextPage := strconv.Itoa(pageNumber + 1)
		random := strconv.Itoa(rand.Int())

		template := html.CollapseWhitespace(`
			<div class="container" hx-get="/page/%s" hx-swap="afterend limit:10" hx-trigger="revealed">
				This is page %s<br><br>
				Randomly generated <b>HTML</b> %s<br><br>
				I wish I were a haiku.
			</div>`)

		content := fmt.Sprintf(template, nextPage, thisPage, random)
		return ctx.HTML(200, content)
	})

	e.Logger.Fatal(e.Start(":80"))
}

func postTemplate() formatFunc {

	// return templateFormatFunc("post.html", "DDD")
	return templateFormatFunc("post.html", `
		<div>
			<div class="bold">Post: {{.title}}</div>
			<div>{{.body}}</div>
			<div>id: {{.id}} user: {{.userId}}</div>
		</div>`)
}

func commentTemplate() formatFunc {

	// return templateFormatFunc("comment.html", "CCC")
	return templateFormatFunc("comment.html", `
		<div>
			<div class="bold">Comment: {{.name}}</div>
			<div>{{.email}}</div>
			<div>{{.body}}</div>
		</div>`)
}

func albumTemplate() formatFunc {

	return templateFormatFunc("album.html", `
		<div>
			<div class="bold">Album: {{.title}}</div>
			<div>id: {{.id}}</div>
		</div>`)
}

func todoTemplate() formatFunc {

	return templateFormatFunc("todo.html", `
		<div>
			<div class="bold">ToDo:{{.id}}: {{.title}}</div>
			<div>Complete? {{.completed}}</div>
		</div>`)
}

func userTemplate() formatFunc {

	return templateFormatFunc("user.html", `
		<div>
			<div class="bold">User: {{.name}} / {{.username}}</div>
			<div>{{.email}}</div>
			<div>{{.address.street}} {{.address.suite}}<br>{{.address.city}}, {{.address.zipcode}}</div>
		</div>`)
}

// handleStream creates an echo.HandlerFunc that streams events from an eventSource to client browsers
func handleStream(eventSource chan string) echo.HandlerFunc {

	return func(ctx echo.Context) error {

		done := make(chan bool)

		r := ctx.Request()
		w := ctx.Response().Writer

		// Make sure that the writer supports flushing.
		f, ok := w.(http.Flusher)

		if !ok {
			return derp.New(500, "handler.ServerSentEvent", "Streaming Not Supported")
		}

		// Listen to the closing of the http connection via the CloseNotifier
		if closeNotifier, ok := w.(http.CloseNotifier); ok {
			notify := closeNotifier.CloseNotify()
			go func() {
				<-notify
				log.Println("HTTP connection just closed.")
				done <- true
			}()
		}

		// Set the headers related to event streaming.
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		types := []string{}

		if param := ctx.QueryParam("types"); param != "" {
			types = strings.Split(param, ",")
		}

		// Don't close the connection, instead loop endlessly.
		for {

			select {
			case <-done:
				break

			case message := <-eventSource:

				if len(types) > 0 {
					eventType := types[rand.Int()%len(types)]
					fmt.Fprintf(w, "event: %s\n", eventType)
				}

				fmt.Fprintf(w, "data: %s\n\n", message)

				// Flush the response.  This is only possible if the response supports streaming.
				f.Flush()
			}
		}

		// Done
		// b.RemoveClient <- client
		log.Println("Finished HTTP request at ", r.URL.Path)

		return nil
	}
}

// makeStream loops through  an array of interfaces
func makeStream(data []interface{}, format formatFunc) chan string {

	result := make(chan string)

	go func() {

		for {
			for _, record := range data {
				result <- format(record)
				time.Sleep((time.Duration(100 + rand.Int()%1000)) * time.Millisecond)

			}
		}
	}()

	return result
}

func jsonFormatFunc(data interface{}) string {
	result, _ := json.Marshal(data)
	return string(result)
}

func templateFormatFunc(name string, text string) formatFunc {

	f, _ := template.New(name).Parse(html.CollapseWhitespace(text))

	return func(data interface{}) string {

		var buffer bytes.Buffer

		f.Execute(&buffer, data)

		return buffer.String()
	}
}
