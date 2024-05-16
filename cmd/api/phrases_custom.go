package main

import (
	"errors"
	"net/http"

	"github.com/jempe/whatsapp_backup/internal/data"
	"github.com/jempe/whatsapp_backup/internal/validator"
	"github.com/pgvector/pgvector-go"
)

func (app *application) listPhraseSemanticHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Search             string
		Similarity         float64
		EmbeddingsProvider string
		ContentField       string
		MessageID          int64
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	defaultEmbeddingsProvider := "sentence-transformers"

	if app.config.openAIApiKey != "" || !app.config.sentenceTransformersEnable {
		defaultEmbeddingsProvider = "openai"
	}

	input.Search = app.readString(qs, "search", "")
	input.Similarity = app.readFloat(qs, "similarity", 0.7, v)
	input.EmbeddingsProvider = app.readString(qs, "embeddings-provider", defaultEmbeddingsProvider)

	input.ContentField = app.readString(qs, "content_field", "")
	input.MessageID = app.readInt64(qs, "generic_item_id", 0, v)

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

	if !(input.EmbeddingsProvider == "sentence-transformers" || input.EmbeddingsProvider == "openai") {
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
	} else if input.EmbeddingsProvider == "openai" {
		embeddings, err = fetchOpenaiEmbeddings(searchInput, app.config.openAIApiKey)
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

//custom_code
