package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
)

type App struct {
	errorLog  *log.Logger
	infoLog   *log.Logger
	templates *template.Template
}

func main() {

	addr := flag.String("addr", ":4001", "HTTP network address")
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
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)

}
