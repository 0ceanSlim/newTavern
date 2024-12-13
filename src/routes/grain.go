package routes

import (
	"net/http"

	"goFrame/src/utils"
)

func Grain(w http.ResponseWriter, r *http.Request) {

	data := utils.PageData{
		Title: "Grain",
	}

	// Render the template
	utils.RenderTemplate(w, data, "grain.html", false)
}
