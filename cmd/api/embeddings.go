package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type ContentTokens struct {
	Content string `json:"content"`
	Tokens  int    `json:"tokens"`
}

type openaiApiRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

func (app *application) countTokens(input string) (tokens int, err error) {
	var tokensJSON struct {
		Tokens int `json:"num_tokens"`
	}

	command := app.config.scriptsPath.pythonBinary
	commandArgs := []string{app.config.scriptsPath.path + "/python/count_tokens.py"}

	var output string

	output, err = app.runCommandWithInput(command, commandArgs, input)
	if err != nil {
		return 0, err
	} else {
		//TODO unify the error handling with readJSON helper
		dec := json.NewDecoder(strings.NewReader(output))
		dec.DisallowUnknownFields()

		err = dec.Decode(&tokensJSON)
		if err != nil {
			return 0, err
		}

		return tokensJSON.Tokens, nil
	}
}

func (app *application) splitText(input string, tokenLimit int) (output []string, err error) {

	partLength := tokenLimit * 5

	if len(input) < partLength {
		output = append(output, input)
		return output, nil
	}

	inputParts := []string{}

	currentPosition := 0

	for currentPosition < len(input) {
		if currentPosition+partLength > len(input) {
			inputParts = append(inputParts, input[currentPosition:])
			currentPosition = len(input)
		} else {
			firstSpaceAfterLength := strings.Index(input[currentPosition+partLength:], " ")
			firstSpaceBeforeLength := strings.LastIndex(input[0:currentPosition+partLength], " ")

			if firstSpaceAfterLength != -1 {
				inputParts = append(inputParts, input[currentPosition:currentPosition+partLength+firstSpaceAfterLength])
				currentPosition += partLength + firstSpaceAfterLength + 1
			} else if firstSpaceBeforeLength != -1 {
				inputParts = append(inputParts, input[0:firstSpaceBeforeLength])
				currentPosition += firstSpaceBeforeLength + 1
			} else {
				inputParts = append(inputParts, input[currentPosition:])
				currentPosition += len(input[currentPosition:])
			}
		}
	}

	return inputParts, nil
}

func (app *application) fetchSentenceTransformersEmbeddings(input []string) ([][]float32, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("No input provided")
	}

	var embeddingsJSON struct {
		Embeddings [][]float32 `json:"embeddings"`
		Error      string      `json:"error"`
		Message    string      `json:"message"`
	}

	command := app.config.scriptsPath.pythonBinary
	commandArgs := []string{app.config.scriptsPath.path + "/python/fetch_st_embeddings.py"}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	output, err := app.runCommandWithInput(command, commandArgs, string(inputJSON))
	if err != nil {
		return nil, err
	} else {
		//TODO unify the error handling with readJSON helper
		dec := json.NewDecoder(strings.NewReader(output))
		dec.DisallowUnknownFields()

		err = dec.Decode(&embeddingsJSON)
		if err != nil {
			return nil, err
		}

		return embeddingsJSON.Embeddings, nil
	}
}

func fetchOpenaiEmbeddings(input []string, apiKey string) ([][]float32, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("No input provided")
	}

	url := "https://api.openai.com/v1/embeddings"
	data := &openaiApiRequest{
		Input: input,
		Model: "text-embedding-ada-002",
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		apiErrorMessage, _ := ioutil.ReadAll(resp.Body)

		return nil, fmt.Errorf("OpenAI API Bad status code: %d, message %s", resp.StatusCode, apiErrorMessage)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	var embeddings [][]float32
	for _, item := range result["data"].([]interface{}) {
		var embedding []float32
		for _, v := range item.(map[string]interface{})["embedding"].([]interface{}) {
			embedding = append(embedding, float32(v.(float64)))
		}
		embeddings = append(embeddings, embedding)
	}
	return embeddings, nil
}
