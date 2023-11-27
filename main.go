package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type ChatGTPRequest struct {
	Messages []ChatGTPMessages `json:"messages"`
}

type ChatGTPMessages struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatGTPResponse struct {
	Choices []ChatGTPResponseChoice `json:"choices"`
}

type ChatGTPResponseChoice struct {
	Message ChatGTPResponseMessage `json:"message"`
}

type ChatGTPResponseMessage struct {
	Index        int    `json:"index"`
	Role         string `json:"role"`
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason"`
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
}

func main() {
	loadEnv()

	r := mux.NewRouter()

	r.HandleFunc("/chat", chatHandler).Methods("POST")

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	var req ChatGTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	resp, err := chatGPTAPI(req)
	if err != nil {
		http.Error(w, "Error from GPT API", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func chatGPTAPI(request ChatGTPRequest) (ChatGTPResponse, error) {
	apiKey := os.Getenv("OPEN_API_KEY")
	url := "https://api.openai.com/v1/chat/completions"
	bodyData := map[string]interface{}{
		"messages": request.Messages,
		"model":    "gpt-3.5-turbo",
	}

	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return ChatGTPResponse{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return ChatGTPResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ChatGTPResponse{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return ChatGTPResponse{}, err
	}

	var respData ChatGTPResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		return ChatGTPResponse{}, err
	}

	return respData, nil
}
