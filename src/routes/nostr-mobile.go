package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func NostrMobile(w http.ResponseWriter, r *http.Request) {

    data := utils.PageData{
        Title:       "nostr Mobile",

    }

    utils.RenderTemplate(w, data, "nostr-mobile.html", false)
}
