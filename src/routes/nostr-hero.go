package routes

import (
	"net/http"

	"goFrame/src/utils"
)

func NostrHero(w http.ResponseWriter, r *http.Request) {

	data := utils.PageData{
		Title: "Nostr Hero",
	}

	// Render the template
	utils.RenderTemplate(w, data, "nostr-hero.html", false)
}
