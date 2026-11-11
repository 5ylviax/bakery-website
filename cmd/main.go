package main

import (
	"context"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	errorLog  *log.Logger
	infoLog   *log.Logger
	templates *template.Template
}

func main() {

	addr := flag.String("addr", "127.0.0.1:4001", "HTTP network address")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	templates, err := parseTemplates()
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &App{
		errorLog:  errorLog,
		infoLog:   infoLog,
		templates: templates,
	}

	srv := &http.Server{
		Addr:              *addr,
		ErrorLog:          errorLog,
		Handler:           app.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Ctrl+C locally, or SIGTERM from systemctl restart/stop.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		infoLog.Printf("Starting server on %s", *addr)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		errorLog.Fatal(err) // e.g. port already in use: exit so systemd notices
	case <-ctx.Done():
	}

	infoLog.Println("Shutting down, finishing open requests...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		errorLog.Println(err)
	}
}
