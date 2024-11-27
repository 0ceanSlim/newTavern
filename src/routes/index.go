package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {

	data := utils.PageData{
		Title: "🏠",
	}

	utils.RenderTemplate(w, data, "index.html", false)
}
