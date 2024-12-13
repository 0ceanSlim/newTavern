package routes

import (
	"net/http"

	"goFrame/src/utils"
)

func GunBlog(w http.ResponseWriter, r *http.Request) {

	data := utils.PageData{
		Title: "Gun Blog",
	}

	// Render the template
	utils.RenderTemplate(w, data, "gun-blog.html", false)
}
