package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jempe/whatsapp_backup/internal/data"
	"github.com/jempe/whatsapp_backup/internal/validator"
	"github.com/pgvector/pgvector-go"
)

func (app *application) createPhraseHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title        string `json:"title"`
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
		Title:        input.Title,
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
		app.handleCustomPhraseErrors(err, w, r, v)
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
		Title        *string `json:"title"`
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

	if input.Title != nil {
		phrase.Title = *input.Title
	}

	if input.Tokens != nil {
		phrase.Tokens = *input.Tokens
	}

	if input.Content != nil {
		phrase.Content = *input.Content

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

		if errors.Is(err, data.ErrEditConflict) {
			app.editConflictResponse(w, r)
			app.handleCustomPhraseErrors(err, w, r, v)
		} else {
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
		MessageID int64
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.MessageID = app.readInt64(qs, "message_id", 0, v)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{
		"id",
		"message_id",
		"-id",
		"-message_id",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	phrases, metadata, err := app.models.Phrases.GetAll(
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

/*handle_custom_errors_start*/

func (app application) handleCustomPhraseErrors(err error, w http.ResponseWriter, r *http.Request, v *validator.Validator) {
	switch {
	//	case errors.Is(err, data.ErrDuplicatePhraseTitleEn):
	//		v.AddError("title_en", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicatePhraseTitleEs):
	//		v.AddError("title_es", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicatePhraseTitleFr):
	//		v.AddError("title_fr", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicatePhraseURLEn):
	//		v.AddError("url_en", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicatePhraseURLEs):
	//		v.AddError("url_es", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicatePhraseURLFr):
	//		v.AddError("url_fr", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicatePhraseFolder):
	//		v.AddError("folder", "a video with this folder already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	default:
		app.serverErrorResponse(w, r, err)
	}
}

/*handle_custom_errors_end*/
/*list_phrase_semantic_start*/
func (app *application) listPhraseSemanticHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Search             string
		Similarity         float64
		EmbeddingsProvider string
		ContentFields      []string
		MessageID          int64
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	defaultEmbeddingsProvider := app.config.embeddings.defaultProvider

	input.Search = app.readString(qs, "search", "")
	input.Similarity = app.readFloat(qs, "similarity", 0.7, v)
	input.EmbeddingsProvider = app.readString(qs, "embeddings-provider", defaultEmbeddingsProvider)

	input.ContentFields = app.readCSV(qs, "content_fields", []string{})

	//Additional Semantic Search Filters
	input.MessageID = app.readInt64(qs, "message_id", 0, v)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 5, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{
		"id",
		"-id",
	}

	if input.Search == "" {
		app.serverErrorResponse(w, r, errors.New("missing required search parameter"))
		return
	}

	if !(input.EmbeddingsProvider == "sentence-transformers" || input.EmbeddingsProvider == "openai" || input.EmbeddingsProvider == "google") {
		app.serverErrorResponse(w, r, errors.New("invalid embeddings provider"))
		return
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	searchInput := []string{
		input.Search,
	}

	var embeddings [][]float32
	var err error
	if input.EmbeddingsProvider == "sentence-transformers" {
		embeddings, err = app.fetchSentenceTransformersEmbeddings(searchInput)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	if len(embeddings) == 0 {
		app.serverErrorResponse(w, r, errors.New("no embeddings returned"))
		return
	}

	phrases, metadata, err := app.models.Phrases.GetAllSemantic(
		pgvector.NewVector(embeddings[0]),
		input.Similarity,
		input.EmbeddingsProvider,
		input.ContentFields,
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

/*list_phrase_semantic_end*/
