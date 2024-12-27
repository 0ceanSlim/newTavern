package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func BitcoinDashboard(w http.ResponseWriter, r *http.Request) {
	data := utils.PageData{
		Title: "Bitcoin Dashboard",
	}

	utils.RenderTemplate(w, data, "bitcoin-dashboard.html", false)
}
