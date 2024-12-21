package api

import (
	"encoding/json"
	"errors"
	"github.com/jaam8/web_calculator/internal/calculator"
	"net/http"
)

func RunServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", CalculateHandler)
	http.ListenAndServe(":8080", mux)
}

type Request struct {
	Expression string `json:"expression"`
}

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	request := new(Request)
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	result, err := calculator.Calculate(request.Expression)
	if err != nil {
		res := map[string]string{"error": ""}
		if errors.Is(err, calculator.ErrInvalidExpression) {
			res["error"] = err.Error()
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			res["error"] = "Internal server error"
			w.WriteHeader(http.StatusInternalServerError)
		}
		resBytes, _ := json.Marshal(res)
		w.Write(resBytes)
	} else {
		res := map[string]float64{"result": result}
		resBytes, _ := json.Marshal(res)
		w.WriteHeader(http.StatusOK)
		w.Write(resBytes)
	}
}
