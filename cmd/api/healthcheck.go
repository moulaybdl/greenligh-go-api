package main

import (
	"net/http"
)


func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string {
		"status": "Available",
		"environment": app.config.env,
		"version": version,
	}

	err := app.writeJSON(w, data, http.StatusOK, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}