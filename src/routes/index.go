package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		fileServer := http.FileServer(http.Dir("web"))
		http.StripPrefix("/", fileServer).ServeHTTP(w, r)
		return
	}

	data := utils.PageData{
		Title: "🏠",
	}

	utils.RenderTemplate(w, data, "index.html", false)
}
