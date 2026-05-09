package routes

import (
	"net/http"

	"goFrame/src/utils"
)

func Mill(w http.ResponseWriter, r *http.Request) {

	data := utils.PageData{
		Title: "Mill",
	}

	utils.RenderTemplate(w, data, "mill.html", false)
}
