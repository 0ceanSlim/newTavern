package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func NostrMobileAndroid(w http.ResponseWriter, r *http.Request) {

    data := utils.PageData{
        Title:       "nostr Android",

    }

    utils.RenderTemplate(w, data, "nostr-mobile-android.html", false)
}
