package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func NostrDesktop(w http.ResponseWriter, r *http.Request) {

    data := utils.PageData{
        Title:       "nostr Desktop Clients",

    }

    utils.RenderTemplate(w, data, "nostr-desktop.html", false)
}
