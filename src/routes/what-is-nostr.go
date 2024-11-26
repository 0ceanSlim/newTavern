package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func WhatIsNostr(w http.ResponseWriter, r *http.Request) {

    data := utils.PageData{
        Title:       "nostr",

    }

    utils.RenderTemplate(w, data, "what-is-nostr.html", false)
}
