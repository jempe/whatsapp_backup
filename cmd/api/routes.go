package main

import (
	"net/http"

	"github.com/jempe/whatsapp_backup/ui"
	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/admin/", app.homeHandler)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/contacts", app.requireActivatedUser(app.listContactHandler))
	router.HandlerFunc(http.MethodPost, "/v1/contacts", app.requireActivatedUser(app.createContactHandler))
	router.HandlerFunc(http.MethodGet, "/v1/contacts/:id", app.requireActivatedUser(app.showContactHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/contacts/:id", app.requireActivatedUser(app.updateContactHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/contacts/:id", app.requireActivatedUser(app.deleteContactHandler))

	router.HandlerFunc(http.MethodGet, "/v1/messages", app.requireActivatedUser(app.listMessageHandler))
	router.HandlerFunc(http.MethodPost, "/v1/messages", app.requireActivatedUser(app.createMessageHandler))
	router.HandlerFunc(http.MethodGet, "/v1/messages/:id", app.requireActivatedUser(app.showMessageHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/messages/:id", app.requireActivatedUser(app.updateMessageHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/messages/:id", app.requireActivatedUser(app.deleteMessageHandler))

	router.HandlerFunc(http.MethodGet, "/v1/chats", app.requireActivatedUser(app.listChatHandler))
	router.HandlerFunc(http.MethodPost, "/v1/chats", app.requireActivatedUser(app.createChatHandler))
	router.HandlerFunc(http.MethodGet, "/v1/chats/:id", app.requireActivatedUser(app.showChatHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/chats/:id", app.requireActivatedUser(app.updateChatHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/chats/:id", app.requireActivatedUser(app.deleteChatHandler))

	router.HandlerFunc(http.MethodGet, "/v1/phrases", app.requireActivatedUser(app.listPhraseHandler))
	router.HandlerFunc(http.MethodPost, "/v1/phrases", app.requireActivatedUser(app.createPhraseHandler))
	router.HandlerFunc(http.MethodGet, "/v1/phrases/:id", app.requireActivatedUser(app.showPhraseHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/phrases/:id", app.requireActivatedUser(app.updatePhraseHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/phrases/:id", app.requireActivatedUser(app.deletePhraseHandler))

	router.HandlerFunc(http.MethodGet, "/v1/phrases_search", app.requireActivatedUser(app.listPhraseSemanticHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/password", app.updateUserPasswordHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/activation", app.createActivationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/password_reset", app.createPasswordResetTokenHandler)

	router.HandlerFunc(http.MethodGet, "/admin/login.html", app.userPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/signup.html", app.userPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/activate.html", app.userPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/reset_password.html", app.userPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/forgot_password.html", app.userPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/request_activation.html", app.userPageHandler)

	router.HandlerFunc(http.MethodGet, "/admin/contacts.html", app.contactsPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/contact.html", app.contactPageHandler)

	router.HandlerFunc(http.MethodGet, "/admin/messages.html", app.messagesPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/message.html", app.messagePageHandler)

	router.HandlerFunc(http.MethodGet, "/admin/chats.html", app.chatsPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/chat.html", app.chatPageHandler)

	router.HandlerFunc(http.MethodGet, "/admin/phrases.html", app.phrasesPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/phrase.html", app.phrasePageHandler)

	router.Handler(http.MethodGet, "/static/*filepath", http.FileServerFS(ui.Files))

	//custom_routes

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
}
