package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/messages", app.listMessageHandler)
	router.HandlerFunc(http.MethodPost, "/v1/messages", app.createMessageHandler)
	router.HandlerFunc(http.MethodGet, "/v1/messages/:id", app.showMessageHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/messages/:id", app.updateMessageHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/messages/:id", app.deleteMessageHandler)

	router.HandlerFunc(http.MethodGet, "/v1/chats", app.listChatHandler)
	router.HandlerFunc(http.MethodPost, "/v1/chats", app.createChatHandler)
	router.HandlerFunc(http.MethodGet, "/v1/chats/:id", app.showChatHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/chats/:id", app.updateChatHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/chats/:id", app.deleteChatHandler)

	router.HandlerFunc(http.MethodGet, "/v1/phrases", app.listPhraseHandler)
	router.HandlerFunc(http.MethodPost, "/v1/phrases", app.createPhraseHandler)
	router.HandlerFunc(http.MethodGet, "/v1/phrases/:id", app.showPhraseHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/phrases/:id", app.updatePhraseHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/phrases/:id", app.deletePhraseHandler)

	router.HandlerFunc(http.MethodGet, "/v1/phrases_search", app.listPhraseSemanticHandler)

	//custom_routes

	return app.recoverPanic(app.basicAuth(app.cors(app.rateLimit(router))))
}
