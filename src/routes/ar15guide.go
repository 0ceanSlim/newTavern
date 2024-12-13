package routes

import (
	"net/http"

	"goFrame/src/utils"
)

func ArGuide(w http.ResponseWriter, r *http.Request) {

	data := utils.PageData{
		Title: "AR15 Bild Guide",
	}

	// Render the template
	utils.RenderTemplate(w, data, "ar15guide.html", false)
}
