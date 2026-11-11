package main

import (
	"net/http"
	"os"
	"path/filepath"
)

type bread struct {
	Name   string
	Images []string
	Photos []string
}

var frontPageBreads = []bread{
	{Name: "Pan de muerto", Images: []string{"pan-de-muerto-1.jpg", "pan-de-muerto-2.jpg", "pan-de-muerto-3.jpg"}},
	{Name: "Pan 2", Images: []string{"pan2.jpg"}},
	{Name: "Pan 3", Images: []string{"pan3.jpg"}},
	{Name: "Pan 4", Images: []string{"pan4.jpg"}},
}

func (app *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	// Checked on every request so a new photo shows up without a restart.
	breads := make([]bread, len(frontPageBreads))
	copy(breads, frontPageBreads)
	for i := range breads {
		for _, img := range breads[i].Images {
			if _, err := os.Stat(filepath.Join("static", "img", img)); err == nil {
				breads[i].Photos = append(breads[i].Photos, img)
			}
		}
	}

	app.render(w, "index", breads)
}

func (app *App) handleAbout(w http.ResponseWriter, r *http.Request) {
	app.render(w, "about", nil)
}

func (app *App) handleContact(w http.ResponseWriter, r *http.Request) {
	app.render(w, "contact", nil)
}
