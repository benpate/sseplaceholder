package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/benpate/convert"
	"github.com/benpate/derp"
	"github.com/labstack/echo/v4"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	e := echo.New()

	e.Static("/", "static")
	e.Static("/htmx", "../htmx")

	e.GET("/page/:number", func(ctx echo.Context) error {

		pageNumber, err := strconv.Atoi(ctx.Param("number"))

		if err != nil {
			pageNumber = 1
		}

		thisPage := strconv.Itoa(pageNumber)
		nextPage := strconv.Itoa(pageNumber + 1)
		random := strconv.Itoa(rand.Int())

		content := fmt.Sprintf(`<div class="container" hx-get="/page/%s" hx-swap="afterend limit:10" hx-trigger="revealed">This is page %s<br><br>Randomly generated <b>HTML</b> %s<br><br>I wish I were a haiku.</div>`, nextPage, thisPage, random)
		return ctx.HTML(200, content)
	})

	e.GET("/eventStream", func(ctx echo.Context) error {

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

		connectionID := convert.String(rand.Int())
		counter := 0
		types := []string{}

		if param := ctx.QueryParam("types"); param != "" {
			types = strings.Split(param, ",")
		}

		ticker := make(chan time.Time)

		go func() {
			for {
				ticker <- time.Now()
				time.Sleep((time.Duration(rand.Int() % 2000)) * time.Millisecond)
			}
		}()

		// Don't close the connection, instead loop endlessly.
		for {

			select {
			case <-done:
				break

			case t := <-ticker:

				msg := fmt.Sprintf("<div>Connection: %s<br>Message: %s<br>Current Time: %s<br>Random Content: %s<br>", connectionID, convert.String(counter), t.Format("2006-01-02 3:04:05"), convert.String(rand.Int()))

				if len(types) > 0 {
					eventType := types[rand.Int()%len(types)]
					fmt.Fprintf(w, "event: %s\n", eventType)
					msg = msg + "Event Type: " + eventType
				}
				msg = msg + fmt.Sprintf("</div><div id=\"oob-target\" hx-swap-oob=\"true\" class=\"container\">OOB Swap from ConnectionID: %s<br>Message Number: %s</div>", connectionID, convert.String(counter))
				fmt.Fprintf(w, "data: %s\n\n", msg)

				// Flush the response.  This is only possible if the response supports streaming.
				f.Flush()
			}

			counter = counter + 1
		}

		// Done
		// b.RemoveClient <- client
		log.Println("Finished HTTP request at ", r.URL.Path)

		return nil

	})

	fmt.Println("Starting web server..")
	e.Logger.Fatal(e.Start(":80"))
}
