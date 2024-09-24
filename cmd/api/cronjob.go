package main

import (
	"fmt"
	"strconv"

	"github.com/jempe/whatsapp_backup/internal/data"
	"github.com/pgvector/pgvector-go"
)

func (app *application) runCronJob() error {
	app.logger.PrintInfo("starting the cron job", nil)

	var embeddingsProviders []string
	if app.config.embeddings.sentenceTransformersServerURL != "" {
		embeddingsProviders = append(embeddingsProviders, "sentence-transformers")
	}

	maxTokens := 200

	itemsPerBatch := app.config.embeddings.embeddingsPerBatch

	if len(embeddingsProviders) == 0 {
		return fmt.Errorf("No embeddings providers configured", nil)
	}

	filter := data.Filters{
		Page:     1,
		PageSize: itemsPerBatch,
		Sort:     "-id",
		SortSafelist: []string{
			"-id",
		},
	}

	var fields []string

	fields = []string{
		"messages.message",
	}

	for _, field := range fields {

		messages, metadata, err := app.models.Messages.GetAllNotInSemantic(filter, field)
		if err != nil {
			return err
		}

		app.logger.PrintInfo("Messages to process", map[string]string{
			"field":      field,
			"total":      strconv.Itoa(metadata.TotalRecords),
			"processing": strconv.Itoa(len(messages)),
		})

		for _, message := range messages {

			content := ""

			switch field {
			case "messages.message":
				content = message.Message
			}

			titleField := message.Message

			if content != "" {
				countedTokens, err := app.countTokens(content)
				if err != nil {
					return err
				}

				var parts []string
				var partsErr error

				if countedTokens < maxTokens {
					parts = []string{content}
				} else {
					parts, partsErr = app.splitText(content, maxTokens)
					if partsErr != nil {
						return partsErr
					}
				}

				for seq, part := range parts {

					countedTokens, err := app.countTokens(part)
					if err != nil {
						return err
					}

					phrasePart := &data.Phrase{
						Title:        titleField,
						Content:      part,
						Tokens:       countedTokens,
						Sequence:     seq + 1,
						MessageID:    message.ID,
						ContentField: field,
					}

					if countedTokens > app.config.embeddings.maxTokens {
						return fmt.Errorf("Document of Message %d has too many tokens %d, sequence: %d", message.ID, countedTokens, seq)
					} else {
						err = app.models.Phrases.Insert(phrasePart)
						if err != nil {
							return err
						}
					}
				}
			} else {
				app.logger.PrintInfo("No content to process", map[string]string{
					"message_id": strconv.Itoa(int(message.ID)),
				})
			}
		}
	}

	for _, provider := range embeddingsProviders {
		phrases, metadata, err := app.models.Phrases.GetAllWithoutEmbeddings(itemsPerBatch, provider)
		if err != nil {
			return err
		}

		app.logger.PrintInfo("Phrases to process", map[string]string{
			"total":      strconv.Itoa(metadata.TotalRecords),
			"processing": strconv.Itoa(len(phrases)),
		})

		if len(phrases) == 0 {
			continue
		}

		var embeddingsContent []string

		for _, phrase := range phrases {
			embeddingsContent = append(embeddingsContent, phrase.Content)
		}

		var embeddings [][]float32
		if provider == "sentence-transformers" {
			embeddings, err = app.fetchSentenceTransformersEmbeddings(embeddingsContent)
		}

		if err != nil {
			return err
		}

		app.logger.PrintInfo(fmt.Sprintf("Embeddings fetched from %s", provider), map[string]string{
			"total": strconv.Itoa(len(embeddings)),
		})

		for i, phrase := range phrases {
			app.models.Phrases.UpdateEmbedding(phrase, pgvector.NewVector(embeddings[i]), provider)
		}
	}
	return nil

}
