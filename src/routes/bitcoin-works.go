package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func BitcoinWorks(w http.ResponseWriter, r *http.Request) {

	data := utils.PageData{
		Title: "Bitcoin Works!",
	}

	utils.RenderTemplate(w, data, "bitcoin-works.html", false)
}
