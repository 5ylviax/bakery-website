package main

import "net/http"

func (app *App) routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", app.handleIndex)
	mux.HandleFunc("/menu", app.handleMenu)
	mux.HandleFunc("/about", app.handleAbout)
	mux.HandleFunc("/contact", app.handleContact)

	//CSS
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("static"))))

	return mux
}
