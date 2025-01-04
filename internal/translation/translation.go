package translation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/JohnSalinas123/linguachat-backend-go/internal/models"
)

func TranslateMessage(content string, startLang string, endLang string) (models.TranslationResponse, error) {

	startLangTrim := strings.Trim(startLang, "{}")
	endLangTrim := strings.Trim(endLang, "{}")

	apiURL := fmt.Sprintf("http://127.0.0.1:5000/api/translate_%s_to_%s", startLangTrim, endLangTrim)
	log.Println(apiURL)

	// create request body
	requestBody, err := json.Marshal(map[string]string {
		"input": content,
	})
	if err != nil {
		return models.TranslationResponse{}, fmt.Errorf("failed to encode request body: %w", err)
	}

	// make POST request to translation api
	response, err := http.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return models.TranslationResponse{}, fmt.Errorf("failed to translate message: %w", err)
	}
	defer response.Body.Close()

	// check for non-200 status code
	if response.StatusCode != http.StatusOK {
		return models.TranslationResponse{}, fmt.Errorf("translation API return status: %d", response.StatusCode)
	}

	// parse response body
	var responseData models.TranslationResponse
	err = json.NewDecoder(response.Body).Decode(&responseData)
	if err != nil {
		return models.TranslationResponse{}, fmt.Errorf("failed to parse translation response: %w", err)
	}

	return responseData, nil

}