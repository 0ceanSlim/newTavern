package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func NostrClients(w http.ResponseWriter, r *http.Request) {

    data := utils.PageData{
        Title:       "nostr",

    }

    utils.RenderTemplate(w, data, "nostr-clients.html", false)
}
