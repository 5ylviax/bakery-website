package main

import "net/http"

func (app *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}
	app.render(w, "index", nil)
}

func (app *App) handleAbout(w http.ResponseWriter, r *http.Request) {
	app.render(w, "about", nil)
}

func (app *App) handleContact(w http.ResponseWriter, r *http.Request) {
	app.render(w, "contact", nil)
}
