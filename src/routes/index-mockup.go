package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func IndexMockup(w http.ResponseWriter, r *http.Request) {
	data := utils.PageData{
		Title: "🏠",
	}

	utils.RenderTemplate(w, data, "index-new-mockup.html", false)
}
