package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func Mempool(w http.ResponseWriter, r *http.Request) {

	data := utils.PageData{
		Title: "mempool",
	}

	utils.RenderTemplate(w, data, "mempool.html", false)
}
