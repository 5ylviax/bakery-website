package main

import "testing"

func TestMenuJSON(t *testing.T) {
	// Tests run inside cmd/, so the menu is one folder up.
	if _, err := loadMenu("../data/menu.json"); err != nil {
		t.Fatal(err)
	}
}

func TestParseMenuRejectsMistakes(t *testing.T) {
	bad := map[string]string{
		"typo in a field":    `[{"category": "Pan", "items": [{"name": "Concha", "prize": "$1"}]}]`,
		"missing price":      `[{"category": "Pan", "items": [{"name": "Concha"}]}]`,
		"missing category":   `[{"items": []}]`,
		"duplicate category": `[{"category": "Pan"}, {"category": "pan"}]`,
		"broken JSON":        `[{"category": "Pan",}]`,
	}
	for name, data := range bad {
		if _, err := parseMenu([]byte(data)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestSlug(t *testing.T) {
	tests := map[string]string{
		"Pan Dulce":      "pan-dulce",
		"Café de olla":   "cafe-de-olla",
		"  Moño / Bow! ": "mono-bow",
		"A La Cart":      "a-la-cart",
	}
	for in, want := range tests {
		if got := slug(in); got != want {
			t.Errorf("slug(%q) = %q, want %q", in, got, want)
		}
	}
}
