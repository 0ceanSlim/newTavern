package routes

import (
	"net/http"

	"goFrame/src/utils"
)

func Contact(w http.ResponseWriter, r *http.Request) {

	data := utils.PageData{
		Title: "Contact",
	}

	// Render the template
	utils.RenderTemplate(w, data, "contact.html", false)
}
