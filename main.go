package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
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

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		// Don't close the connection, instead loop endlessly.
		for {

			select {
			case <-done:
				break

			case t := <-ticker.C:
				msg := fmt.Sprintf("Connection: %s<br>Message: %s<br>Current Time: %s<br>Random Content: %s<br>", connectionID, convert.String(counter), t.Format("2006-01-02 3:04:05"), convert.String(rand.Int()))
				if len(types) > 0 {
					eventType := types[rand.Int()%len(types)]
					fmt.Fprintf(w, "event: %s\n", eventType)
					msg = msg + "Event Type: " + eventType
				}
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
