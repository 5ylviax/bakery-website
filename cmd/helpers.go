package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
)

func (app *App) render(w http.ResponseWriter, name string, data any) {
	var buf bytes.Buffer
	if err := app.templates.ExecuteTemplate(&buf, name, data); err != nil {
		app.serverError(w, err)
		return
	}
	buf.WriteTo(w)
}

func (app *App) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *App) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *App) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}
