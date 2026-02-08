package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Assignment_2/cmd/internal/handlers"
	"Assignment_2/cmd/internal/middleware"
	"Assignment_2/cmd/internal/store"
)

func main() {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "secret12345"
	}

	logMessage := os.Getenv("LOG_MESSAGE")
	if logMessage == "" {
		logMessage = "Task API"
	}

	st := store.New()

	taskHandler := handlers.NewTaskHandler(st)
	externalHandler := handlers.NewExternalHandler()

	mux := http.NewServeMux()
	mux.Handle("/tasks", taskHandler)
	mux.Handle("/external/todos", externalHandler)

	// Middlewares (required in PDF): API key + logging :contentReference[oaicite:3]{index=3}
	var h http.Handler = mux
	h = middleware.APIKey(h, apiKey)
	h = middleware.Logging(h, logMessage)
	h = middleware.RequestID(h)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Run server
	go func() {
		log.Println("Server running on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = srv.Shutdown(ctx)
	log.Println("Server stopped")
}

/*
	START SERVER

1) AUTH (no API key)
curl -i http://localhost:8080/tasks

2) CREATE TASK
curl -i -H "X-API-KEY: secret12345" -H "Content-Type: application/json" \
  -d '{"title":"Write unit tests"}' http://localhost:8080/tasks

3) LIST TASKS
curl -i -H "X-API-KEY: secret12345" http://localhost:8080/tasks

4) GET TASK BY ID
curl -i -H "X-API-KEY: secret12345" "http://localhost:8080/tasks?id=1"

5) ERROR: INVALID ID
curl -i -H "X-API-KEY: secret12345" "http://localhost:8080/tasks?id=abc"

6) ERROR: NOT FOUND
curl -i -H "X-API-KEY: secret12345" "http://localhost:8080/tasks?id=999"

7) UPDATE TASK (PATCH done=true)
curl -i -H "X-API-KEY: secret12345" -H "Content-Type: application/json" \
  -X PATCH -d '{"done":true}' "http://localhost:8080/tasks?id=1"

8) VERIFY UPDATED TASK
curl -i -H "X-API-KEY: secret12345" "http://localhost:8080/tasks?id=1"


OPTIONAL:

# A) FILTER DONE
curl -i -H "X-API-KEY: secret12345" "http://localhost:8080/tasks?done=true"

# B) DELETE TASK
curl -i -H "X-API-KEY: secret12345" -X DELETE "http://localhost:8080/tasks?id=1"

# C) VERIFY DELETED (should be not found)
curl -i -H "X-API-KEY: secret12345" "http://localhost:8080/tasks?id=1"

# D) EXTERNAL API (if implemented)
curl -i -H "X-API-KEY: secret12345" http://localhost:8080/external/todos

*/
