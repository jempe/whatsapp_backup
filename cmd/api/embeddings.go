package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jempe/text_splitter/splitter"
	"github.com/tiktoken-go/tokenizer"
)

func (app *application) countTokens(input string) (tokens int, err error) {
	model := "gpt-3.5-turbo"

	codec, err := tokenizer.ForModel(tokenizer.Model(model))
	if err != nil {
		return 0, err
	}

	ids, _, err := codec.Encode(input)
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}

func (app *application) splitText(input string, tokenLimit int) (output []string, err error) {

	delimiters := []rune{' ', '\n', '\t'}

	chunks := splitter.SplitTextInChunks(input, tokenLimit*5, tokenLimit, delimiters)

	return chunks, nil
}
func (app *application) fetchSentenceTransformersEmbeddings(input []string) ([][]float32, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("No input provided")
	}

	var data struct {
		Documents []string `json:"documents"`
	}

	var result struct {
		Embeddings [][]float32 `json:"embeddings"`
	}

	for _, item := range input {
		data.Documents = append(data.Documents, item)
	}

	url := fmt.Sprintf("%s/embedding", app.config.embeddings.sentenceTransformersServerURL)

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		apiErrorMessage, _ := ioutil.ReadAll(resp.Body)

		return nil, fmt.Errorf("Sentence Transformers Bad status code: %d, message %s", resp.StatusCode, apiErrorMessage)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	var embeddings [][]float32
	for _, item := range result.Embeddings {
		var embedding []float32
		for _, v := range item {
			embedding = append(embedding, v)
		}
		embeddings = append(embeddings, embedding)
	}
	return embeddings, nil
}
