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
		ChatID               int64     `json:"chat_id"`
		EnableSemanticSearch bool      `json:"enable_semantic_search"`
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
		ChatID:               input.ChatID,
		EnableSemanticSearch: input.EnableSemanticSearch,
	}

	v := validator.New()

	if data.ValidateMessage(v, message, validator.ActionCreate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Messages.Insert(message)
	if err != nil {
		app.handleCustomMessageErrors(err, w, r, v)
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
		ChatID               *int64     `json:"chat_id"`
		EnableSemanticSearch *bool      `json:"enable_semantic_search"`
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

	if input.ChatID != nil {
		message.ChatID = *input.ChatID
	}

	if input.EnableSemanticSearch != nil {
		message.EnableSemanticSearch = *input.EnableSemanticSearch
	}

	v := validator.New()

	if data.ValidateMessage(v, message, validator.ActionUpdate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Messages.Update(message)
	if err != nil {

		if errors.Is(err, data.ErrEditConflict) {
			app.editConflictResponse(w, r)
			app.handleCustomMessageErrors(err, w, r, v)
		} else {
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
		"-id",
		"-message_date",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	messages, metadata, err := app.models.Messages.GetAll(
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

/*handle_custom_errors_start*/

func (app application) handleCustomMessageErrors(err error, w http.ResponseWriter, r *http.Request, v *validator.Validator) {
	switch {
	//	case errors.Is(err, data.ErrDuplicateMessageTitleEn):
	//		v.AddError("title_en", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateMessageTitleEs):
	//		v.AddError("title_es", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateMessageTitleFr):
	//		v.AddError("title_fr", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateMessageURLEn):
	//		v.AddError("url_en", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateMessageURLEs):
	//		v.AddError("url_es", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateMessageURLFr):
	//		v.AddError("url_fr", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateMessageFolder):
	//		v.AddError("folder", "a video with this folder already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	default:
		app.serverErrorResponse(w, r, err)
	}
}

/*handle_custom_errors_end*/
