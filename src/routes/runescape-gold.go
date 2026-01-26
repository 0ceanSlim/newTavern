package routes

import (
	"goFrame/src/utils"
	"net/http"
)

func RunescapeGold(w http.ResponseWriter, r *http.Request) {
	data := utils.PageData{
		Title: "RuneScape Gold Tracker",
	}

	utils.RenderTemplate(w, data, "runescape-gold.html", false)
}
