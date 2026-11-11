package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type menuItem struct {
	Name        string   `json:"name"`
	Price       string   `json:"price"`
	Description string   `json:"description"`
	Images      []string `json:"images"`
	Photos      []string `json:"-"`
}

type menuCategory struct {
	Name  string     `json:"category"`
	ID    string     `json:"-"`
	Items []menuItem `json:"items"`
}

// loadMenu reads data/menu.json. It runs once at startup: restart the
// server after editing the menu.
func loadMenu(path string) ([]menuCategory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseMenu(data)
}

func parseMenu(data []byte) ([]menuCategory, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // catches typos like "prize" or "image"

	var menu []menuCategory
	if err := dec.Decode(&menu); err != nil {
		return nil, fmt.Errorf("menu.json: %w", err)
	}

	seen := make(map[string]bool)
	for i := range menu {
		c := &menu[i]
		c.ID = slug(c.Name)
		if c.ID == "" {
			return nil, fmt.Errorf("menu.json: category %d needs a name", i+1)
		}
		if seen[c.ID] {
			return nil, fmt.Errorf("menu.json: category %q is listed twice", c.Name)
		}
		seen[c.ID] = true

		for j, item := range c.Items {
			if item.Name == "" || item.Price == "" {
				return nil, fmt.Errorf("menu.json: item %d in %q needs a name and a price", j+1, c.Name)
			}
		}
	}
	return menu, nil
}

var accents = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")

// slug turns "Pan Dulce" into "pan-dulce" for use in links like #pan-dulce.
func slug(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range accents.Replace(strings.ToLower(name)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			if dash && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			dash = false
		} else {
			dash = true
		}
	}
	return b.String()
}
