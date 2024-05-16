package main

import (
	"fmt"
	"strconv"

	"github.com/jempe/whatsapp_backup/internal/data"
	"github.com/pgvector/pgvector-go"
)

func (app *application) runCronJob() {
	app.logger.PrintInfo("starting the cron job", nil)

	var embeddingsProviders []string

	if app.config.sentenceTransformersEnable {
		embeddingsProviders = append(embeddingsProviders, "sentence-transformers")
	}

	if app.config.openAIApiKey != "" {
		embeddingsProviders = append(embeddingsProviders, "openai")
	}

	maxTokens := 200

	itemsPerBatch := app.config.embeddingsPerBatch

	if len(embeddingsProviders) > 0 {

		filter := data.Filters{
			Page:     1,
			PageSize: itemsPerBatch,
			Sort:     "-id",
			SortSafelist: []string{
				"-id",
			},
		}

		messages, metadata, err := app.models.Messages.GetAllNotInSemantic(filter)
		if err != nil {
			app.logger.PrintError(err, nil)
			return
		}

		app.logger.PrintInfo("Messages to process", map[string]string{
			"total":      strconv.Itoa(metadata.TotalRecords),
			"processing": strconv.Itoa(len(messages)),
		})

		for _, message := range messages {

			fields := []string{
				"messages.message",
			}

			for _, field := range fields {
				content := ""

				switch field {
				case "messages.message":
					content = message.Message
				}

				if content != "" {
					countedTokens, err := app.countTokens(content)
					if err != nil {
						app.logger.PrintError(err, nil)
						return
					}

					if countedTokens < maxTokens {

						phrase := &data.Phrase{
							Content:      content,
							Tokens:       countedTokens,
							Sequence:     1,
							MessageID:    message.ID,
							ContentField: field,
						}

						err := app.models.Phrases.Insert(phrase)
						if err != nil {
							app.logger.PrintError(err, nil)
							return
						}
					} else {
						parts, partsErr := app.splitText(content, maxTokens)
						if partsErr != nil {
							app.logger.PrintError(partsErr, nil)
							return
						}

						for seq, part := range parts {

							countedTokens, err := app.countTokens(part)
							if err != nil {
								app.logger.PrintError(err, nil)
								return
							}

							phrasePart := &data.Phrase{
								Content:      part,
								Tokens:       countedTokens,
								Sequence:     seq + 1,
								MessageID:    message.ID,
								ContentField: field,
							}

							err = app.models.Phrases.Insert(phrasePart)
							if err != nil {
								app.logger.PrintError(err, nil)
								return
							}
						}
					}
				}
			}
		}

	}

	for _, provider := range embeddingsProviders {
		phrases, metadata, err := app.models.Phrases.GetAllWithoutEmbeddings(itemsPerBatch, provider)
		if err != nil {
			app.logger.PrintError(err, nil)
			return
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

			if err != nil {
				app.logger.PrintError(err, nil)
				return
			}
		} else if provider == "openai" {
			embeddings, err = fetchOpenaiEmbeddings(embeddingsContent, app.config.openAIApiKey)

			if err != nil {
				app.logger.PrintError(err, nil)
				return
			}
		}

		app.logger.PrintInfo(fmt.Sprintf("Embeddings fetched from %s", provider), map[string]string{
			"total": strconv.Itoa(len(embeddings)),
		})

		for i, phrase := range phrases {
			app.models.Phrases.UpdateEmbedding(phrase, pgvector.NewVector(embeddings[i]), provider)
		}

	}
}
