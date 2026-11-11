package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type bread struct {
	Name   string
	Images []string
	Photos []string
	Price  string // from menu.json, matched by name
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
	prices := make(map[string]string)
	for _, c := range app.menu {
		for _, item := range c.Items {
			prices[strings.ToLower(item.Name)] = item.Price
		}
	}

	for i := range breads {
		breads[i].Photos = existingPhotos(breads[i].Images)
		breads[i].Price = prices[strings.ToLower(breads[i].Name)]
	}

	app.render(w, "index", breads)
}

func (app *App) handleMenu(w http.ResponseWriter, r *http.Request) {
	menu := make([]menuCategory, len(app.menu))
	for i, c := range app.menu {
		items := make([]menuItem, len(c.Items))
		for j, item := range c.Items {
			item.Photos = existingPhotos(item.Images)
			items[j] = item
		}
		c.Items = items
		menu[i] = c
	}

	app.render(w, "menu", menu)
}

// existingPhotos returns the names that have a file in static/img.
func existingPhotos(names []string) []string {
	var found []string
	for _, name := range names {
		if _, err := os.Stat(filepath.Join("static", "img", name)); err == nil {
			found = append(found, name)
		}
	}
	return found
}

func (app *App) handleAbout(w http.ResponseWriter, r *http.Request) {
	app.render(w, "about", nil)
}

func (app *App) handleContact(w http.ResponseWriter, r *http.Request) {
	app.render(w, "contact", nil)
}
