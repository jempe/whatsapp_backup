package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	BRACKET_START = "BRACKET_START"
)

func (app *application) importWhatsappExportHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Export string `json:"export"`
		ChatID int64  `json:"chat_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	/*message := &data.Message{
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
	}*/
}

func getMessagesFromWhatsappExport(dataFile string) (messages []map[string]string, err error) {

	dateRegex := regexp.MustCompile(`^\[\d+/\d+/\d+, \d+:\d+:\d+\u202f(A|P)M\]`)
	startBracketRegex := regexp.MustCompile(`^\[`)
	startSpaceRegex := regexp.MustCompile(`^\u200e`)

	data := ""
	lines := strings.Split(dataFile, "\n")

	for _, line := range lines {
		if startBracketRegex.MatchString(line) && !dateRegex.MatchString(line) {
			line = startBracketRegex.ReplaceAllString(line, BRACKET_START)
		}

		if startSpaceRegex.MatchString(line) {
			line = startSpaceRegex.ReplaceAllString(line, "")
		}

		data += line + "\n"
	}

	// Replace the first character of the file
	data = strings.Replace(data, "[", "", 1)

	// Split the data into lines
	lines = strings.Split(data, "\n[")

	messages = []map[string]string{}

	for _, line := range lines {
		dateParts := strings.Split(line, "]")

		dateString := strings.Replace(dateParts[0], "[", "", 1)
		message := strings.Replace(dateParts[1], BRACKET_START, "[", 1)
		message = strings.TrimSpace(message)

		attachment := findStringEnclosedIn(line, "<attached:", ">")
		phoneNumber := findStringEnclosedIn(line, "\u202a", "\u202c")

		message = strings.Replace(message, "<attached:"+attachment+">", "", -1)
		message = strings.Replace(message, "\u202a"+phoneNumber+"\u202c:", "", -1)

		date, err := time.Parse("02/01/2006, 03:04:05 PM", dateString)
		if err != nil {
			fmt.Println("Error converting date string:", err)
			continue
		}

		formattedDate := date.Format("2006-01-02 15:04:05")

		messageData := map[string]string{
			"date":         formattedDate,
			"phone_number": phoneNumber,
			"message":      message,
			"attachment":   strings.TrimSpace(attachment),
		}

		//saveMessageToDatabase(messageData)

		// TODO: Fix issue with messages that contains [

		fmt.Println(messageData)
	}

	return messages, nil
}

func findStringEnclosedIn(input, start, end string) string {
	// Implement this function to find a string enclosed between `start` and `end`
	findStartOccurrence := strings.Index(input, start) + len(start)
	if findStartOccurrence == -1 {
		return ""
	}

	findEndOccurrence := strings.Index(string(input[findStartOccurrence:]), end)

	if findEndOccurrence == -1 {
		return ""
	}

	return strings.TrimSpace(input[findStartOccurrence : findStartOccurrence+findEndOccurrence])
}
