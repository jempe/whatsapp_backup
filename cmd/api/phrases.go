package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jempe/whatsapp_backup/internal/data"
	"github.com/jempe/whatsapp_backup/internal/validator"
)

func (app *application) createPhraseHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Content      string `json:"content"`
		Tokens       int    `json:"tokens"`
		Sequence     int    `json:"sequence"`
		ContentField string `json:"content_field"`
		MessageID    int64  `json:"message_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	phrase := &data.Phrase{
		Content:      input.Content,
		Tokens:       input.Tokens,
		Sequence:     input.Sequence,
		ContentField: input.ContentField,
		MessageID:    input.MessageID,
	}

	if input.Tokens == 0 {
		countedTokens, err := app.countTokens(input.Content)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		phrase.Tokens = countedTokens
	}

	v := validator.New()

	if data.ValidatePhrase(v, phrase, validator.ActionCreate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Phrases.Insert(phrase)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/phrases/%d", phrase.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"phrase": phrase}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showPhraseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	phrase, err := app.models.Phrases.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"phrase": phrase}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updatePhraseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	phrase, err := app.models.Phrases.Get(id)
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
		if strconv.FormatInt(int64(phrase.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}

	var input struct {
		Content      *string `json:"content"`
		Tokens       *int    `json:"tokens"`
		Sequence     *int    `json:"sequence"`
		ContentField *string `json:"content_field"`
		MessageID    *int64  `json:"message_id"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Content != nil {
		phrase.Content = *input.Content
	}

	if input.Tokens != nil {
		phrase.Tokens = *input.Tokens
	} else if input.Content != nil {
		countedTokens, err := app.countTokens(phrase.Content)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		phrase.Tokens = countedTokens
	}

	if input.Sequence != nil {
		phrase.Sequence = *input.Sequence
	}

	if input.ContentField != nil {
		phrase.ContentField = *input.ContentField
	}

	if input.MessageID != nil {
		phrase.MessageID = *input.MessageID
	}

	v := validator.New()

	if data.ValidatePhrase(v, phrase, validator.ActionUpdate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Phrases.Update(phrase)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"phrase": phrase}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deletePhraseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Phrases.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "phrase successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listPhraseHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ContentField string
		MessageID    int64
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.ContentField = app.readString(qs, "content_field", "")

	input.MessageID = app.readInt64(qs, "message_id", 0, v)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{
		"id",
		"content_field",
		"message_id",
		"-id",
		"-content_field",
		"-message_id",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	phrases, metadata, err := app.models.Phrases.GetAll(
		input.ContentField,
		input.MessageID,
		input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"phrases": phrases, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
