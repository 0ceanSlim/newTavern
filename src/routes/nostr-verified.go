package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func NostrVerified(w http.ResponseWriter, r *http.Request) {

	data := utils.PageData{
		Title: "Nostr Verified",
	}

	utils.RenderTemplate(w, data, "nostr-verified.html", false)
}
