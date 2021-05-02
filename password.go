package main

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type PasswordInput struct {
	Password string
}

type PasswordOutput struct {
	Result string
}

func checkPassword(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var data PasswordInput
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		var responseData PasswordOutput
		responseData.Result = "nok"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(responseData)
		return
	}
	if data.Password == "3600" {
		var responseData PasswordOutput
		responseData.Result = "ok"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(responseData)
		return
	}
	var responseData PasswordOutput
	responseData.Result = "nok"
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseData)
}
