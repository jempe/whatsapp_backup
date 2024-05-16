package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jempe/whatsapp_backup/internal/validator"

	"github.com/julienschmidt/httprouter"
)

type envelope map[string]any

func (app *application) readBool(qs url.Values, key string, defaultValue bool) bool {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		return defaultValue
	}

	return b
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *application) readFloat(qs url.Values, key string, defaultValue float64, v *validator.Validator) float64 {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	f, err := strconv.ParseFloat(s, 64)

	if err != nil {
		v.AddError(key, "must be a float value")
		return defaultValue
	}

	return f
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)

	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

func (app *application) readInt64(qs url.Values, key string, defaultValue int64, v *validator.Validator) int64 {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)

	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return int64(i)
}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid ID parameter")
	}
	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json utf-8")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	// Limit the size of the request body to 1MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize the json.Decoder, and call the DisallowUnknownFields() method on it
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}

			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// Call Decode() again, using a pointer to an empty anonymous struct as the destination. to find if there is an additional JSON object
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// runCommandWithInput executes a shell command, sends input via stdin, and checks its return status.
func (app *application) runCommandWithInput(command string, args []string, input string) (string, error) {
	// Define the command that we want to run
	var allowedCommands = []string{
		app.config.scriptsPath.pythonBinary,
	}

	var allowedPaths = []string{
		app.config.scriptsPath.path,
	}

	// Check if the command is allowed
	for _, allowedCommand := range allowedCommands {
		if command != allowedCommand {
			return "", fmt.Errorf("forbidden command: %s", command)
		}
	}

	if len(args) > 0 {
		firstArg := args[0]

		path, err := filepath.Abs(firstArg)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Check if the path is allowed
		for _, allowedPath := range allowedPaths {

			allowedFullPath, err := filepath.Abs(allowedPath)
			if err != nil {
				return "", fmt.Errorf("failed to get absolute path: %w", err)
			}

			if !strings.HasPrefix(path, allowedFullPath) {
				return "", fmt.Errorf("forbidden path: %s", path)
			}
		}
	}

	cmd := exec.Command(command, args...)

	// Set up stdin, stdout, and stderr pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	defer stdin.Close()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Send input via stdin
	if _, err := io.WriteString(stdin, input); err != nil {
		return "", fmt.Errorf("failed to write to stdin: %w", err)
	}
	stdin.Close()

	// Wait for the command to finish and check the exit status
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Non-zero exit code indicates an error
			return stdout.String(), fmt.Errorf("command failed with exit code %d: %s", exitErr.ExitCode(), stderr.String())
		} else {
			return stdout.String(), fmt.Errorf("command failed: %w", err)
		}
	}

	// Successful execution
	return stdout.String(), nil
}
