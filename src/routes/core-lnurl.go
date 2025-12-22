package routes

import (
	"net/http"

	"goFrame/src/utils"
)

func CoreLnurl(w http.ResponseWriter, r *http.Request) {

	data := utils.PageData{
		Title: "Core LNURL",
	}

	// Render the template
	utils.RenderTemplate(w, data, "core-lnurl.html", false)
}
