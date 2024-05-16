package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jempe/whatsapp_backup/internal/data"
	"github.com/jempe/whatsapp_backup/internal/validator"
)

func (app *application) createMessageHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		MessageDate          time.Time `json:"message_date"`
		Message              string    `json:"message"`
		PhoneNumber          string    `json:"phone_number"`
		Attachment           string    `json:"attachment"`
		EnableSemanticSearch bool      `json:"enable_semantic_search"`
		ChatID               int64     `json:"chat_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	message := &data.Message{
		MessageDate:          input.MessageDate,
		Message:              input.Message,
		PhoneNumber:          input.PhoneNumber,
		Attachment:           input.Attachment,
		EnableSemanticSearch: input.EnableSemanticSearch,
		ChatID:               input.ChatID,
	}

	v := validator.New()

	if data.ValidateMessage(v, message, validator.ActionCreate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Messages.Insert(message)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/messages/%d", message.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"message": message}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showMessageHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	message, err := app.models.Messages.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": message}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateMessageHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	message, err := app.models.Messages.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(message.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}

	var input struct {
		MessageDate          *time.Time `json:"message_date"`
		Message              *string    `json:"message"`
		PhoneNumber          *string    `json:"phone_number"`
		Attachment           *string    `json:"attachment"`
		EnableSemanticSearch *bool      `json:"enable_semantic_search"`
		ChatID               *int64     `json:"chat_id"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.MessageDate != nil {
		message.MessageDate = *input.MessageDate
	}

	if input.Message != nil {
		message.Message = *input.Message
	}

	if input.PhoneNumber != nil {
		message.PhoneNumber = *input.PhoneNumber
	}

	if input.Attachment != nil {
		message.Attachment = *input.Attachment
	}

	if input.EnableSemanticSearch != nil {
		message.EnableSemanticSearch = *input.EnableSemanticSearch
	}

	if input.ChatID != nil {
		message.ChatID = *input.ChatID
	}

	v := validator.New()

	if data.ValidateMessage(v, message, validator.ActionUpdate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Messages.Update(message)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Messages.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "message successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listMessageHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EnableSemanticSearch bool
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.EnableSemanticSearch = app.readBool(qs, "enable_semantic_search", false)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{
		"id",
		"message_date",
		"enable_semantic_search",
		"-id",
		"-message_date",
		"-enable_semantic_search",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	messages, metadata, err := app.models.Messages.GetAll(
		input.EnableSemanticSearch,
		input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"messages": messages, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
