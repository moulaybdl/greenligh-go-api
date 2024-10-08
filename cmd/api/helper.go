package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)


type envlope map[string]interface{}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}


func (app *application) writeJSON(w  http.ResponseWriter,data envlope, status int, header http.Header ) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range header {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil

}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {

	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()


	
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError // problem with the syntax of json being decoded
		var unmarshalTypeError *json.UnmarshalTypeError //JSON value type not compatible with the Go destination type
		var invalidUnmarshalError *json.InvalidUnmarshalError // 	destination not valid (problem with the server side)



		switch {
			case errors.As(err, &syntaxError):
				return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

			case errors.Is(err, io.ErrUnexpectedEOF):
				return errors.New("body contain badly-formated JSON")

			case errors.As(err, &unmarshalTypeError):
				if unmarshalTypeError.Field != "" {
					return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
				}
				return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
			
			case errors.Is(err, io.EOF):
				return errors.New("body must not be empty")

			case strings.HasPrefix(err.Error(), "json:unkown field "):
				fieldName := strings.TrimPrefix(err.Error(), "json:unkown field")
				return fmt.Errorf("body contains unkown key %s ", fieldName)

			case err.Error() == "http: request body too large":
				return fmt.Errorf("body must not be larget than %d bytes", maxBytes)
			
			case errors.As(err, &invalidUnmarshalError):
				panic(err)

			default:
				return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil

}

