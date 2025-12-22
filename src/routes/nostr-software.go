package routes

import (
	"net/http"

	"goFrame/src/utils"
)

func NostrSoftware(w http.ResponseWriter, r *http.Request) {

	data := utils.PageData{
		Title: "Nostr Software",
	}

	// Render the template
	utils.RenderTemplate(w, data, "nostr-software.html", false)
}
