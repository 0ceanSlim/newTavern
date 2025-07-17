package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func FileUpload(w http.ResponseWriter, r *http.Request) {
	data := utils.PageData{
		Title: "File Upload",
	}

	utils.RenderTemplate(w, data, "file-upload.html", false)
}
