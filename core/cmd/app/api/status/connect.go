package status

import (
	"encoding/json"
	"net/http"
)

func GetActiveConnect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "test"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResp)
	if err != nil {
		return
	}
}

func GetPassiveConnect(w http.ResponseWriter, r *http.Request) {

}
