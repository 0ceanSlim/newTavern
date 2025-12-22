package routes

import (
	"net/http"

	"goFrame/src/utils"
)

func Gnostream(w http.ResponseWriter, r *http.Request) {

	data := utils.PageData{
		Title: "GNOSTREAM",
	}

	// Render the template
	utils.RenderTemplate(w, data, "gnostream.html", false)
}
