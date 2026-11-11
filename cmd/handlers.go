package main

import (
	"net/http"
	"os"
	"path/filepath"
)

// bread is one box on the front page.
type bread struct {
	Name     string
	Image    string // file name inside static/img
	HasImage bool   // set per request: false shows the placeholder
}

// frontPageBreads are the four boxes on the home page. To add a photo,
// save it as static/img/<Image>. Until that file exists, the box shows a
// "Foto próximamente" placeholder. Change Name to the bread's real name.
var frontPageBreads = []bread{
	{Name: "Pan 1", Image: "pan1.jpg"},
	{Name: "Pan 2", Image: "pan2.jpg"},
	{Name: "Pan 3", Image: "pan3.jpg"},
	{Name: "Pan 4", Image: "pan4.jpg"},
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
		_, err := os.Stat(filepath.Join("static", "img", breads[i].Image))
		breads[i].HasImage = err == nil
	}

	app.render(w, "index", breads)
}

func (app *App) handleAbout(w http.ResponseWriter, r *http.Request) {
	app.render(w, "about", nil)
}

func (app *App) handleContact(w http.ResponseWriter, r *http.Request) {
	app.render(w, "contact", nil)
}
