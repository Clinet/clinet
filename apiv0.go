package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func APIv0() *chi.Mux {
	router := chi.NewRouter()

	//Layout endpoint
	router.Get("/layout/main", v0GetLayoutMain)            //Retrieves the main layout
	router.Get("/layout/guild", v0GetLayoutGuild)          //Retrieves the guild layout
	router.Get("/layout/guild/role", v0GetLayoutGuildRole) //Retrieves the guild roles layout
	router.Get("/layout/user", v0GetLayoutUser)            //Retrieves the user layout

	//Guild endpoint
	router.Get("/guild/{guildID}", v0GetGuild)                           //Retrieves info about a particular guild
	router.Get("/guild/{guildID}/settings", v0GetGuildSettings)          //Retrieves all settings and their values for a particular guild
	router.Put("/guild/{guildID}/settings/{setting}", v0PutGuildSetting) //Sets a new value to a particular guild setting

	//User endpoint
	router.Get("/user/{userID}", v0GetUser)                           //Retrieves info about a particular user
	router.Get("/user/{userID}/settings", v0GetUserSettings)          //Retrieves all settings and their values for a particular user
	router.Put("/user/{userID}/settings/{setting}", v0PutUserSetting) //Sets a new value to a particular user setting

	return router
}

func v0GetLayoutMain(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "stub")
}

func v0GetLayoutGuild(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "stub")
}

func v0GetLayoutGuildRole(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "stub")
}

func v0GetLayoutUser(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "stub")
}

func v0GetGuild(w http.ResponseWriter, r *http.Request) {
	guildID := chi.URLParam(r, "guildID")
	if guildID == "" {
		render.JSON(w, r, errAPI("guildID must not be empty"))
		return
	}

	guild, err := botData.DiscordSession.Guild(guildID)
	if err != nil {
		render.JSON(w, r, errAPI("guildID invalid"))
		return
	}

	render.JSON(w, r, guild)
}

func v0GetGuildSettings(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "stub")
}

func v0PutGuildSetting(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "stub")
}

func v0GetUser(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "stub")
}

func v0GetUserSettings(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "stub")
}

func v0PutUserSetting(w http.ResponseWriter, r *http.Request) {
	render.PlainText(w, r, "stub")
}
