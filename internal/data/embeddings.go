package data

import "fmt"

func getEmbeddingsField(provider string) (embeddingsField string, err error) {
	if provider == "sentence-transformers" {
		return "st_embeddings", nil
	} else if provider == "openai" {
		return "openai_embeddings", nil
	}

	return "", fmt.Errorf("Unknown Embeddings provider %s", provider)
}
