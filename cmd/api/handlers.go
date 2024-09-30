package main

import "net/http"

func (app *application) userPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "User Page"
	app.render(w, r, http.StatusOK, "auth_pages.tmpl", data)
}

func (app *application) homeHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Dashboard"
	app.render(w, r, http.StatusOK, "dashboard.tmpl", data)
}

func (app *application) contactsPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Contacts"
	app.render(w, r, http.StatusOK, "contacts.tmpl", data)
}

func (app *application) contactPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Contact"
	app.render(w, r, http.StatusOK, "contacts_item.tmpl", data)
}

func (app *application) messagesPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Messages"
	app.render(w, r, http.StatusOK, "messages.tmpl", data)
}

func (app *application) messagePageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Message"
	app.render(w, r, http.StatusOK, "messages_item.tmpl", data)
}

func (app *application) chatsPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Chats"
	app.render(w, r, http.StatusOK, "chats.tmpl", data)
}

func (app *application) chatPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Chat"
	app.render(w, r, http.StatusOK, "chats_item.tmpl", data)
}

func (app *application) phrasesPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Phrases"
	app.render(w, r, http.StatusOK, "phrases.tmpl", data)
}

func (app *application) phrasePageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Phrase"
	app.render(w, r, http.StatusOK, "phrases_item.tmpl", data)
}
