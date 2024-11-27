package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func NostrMobileIos(w http.ResponseWriter, r *http.Request) {

    data := utils.PageData{
        Title:       "nostr iOS",

    }

    utils.RenderTemplate(w, data, "nostr-mobile-ios.html", false)
}
