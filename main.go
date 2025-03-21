package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go"
	conversations "github.com/twilio/twilio-go/rest/conversations/v1"
)

var client *twilio.RestClient

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client = twilio.NewRestClient()

	http.HandleFunc("/token", tokenHandler)
	http.HandleFunc("/create-conversation", createConversationHandler)
	http.HandleFunc("/", serveStaticFiles)

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	identity := r.URL.Query().Get("identity")
	if identity == "" {
		http.Error(w, "Identity required", http.StatusBadRequest)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": os.Getenv("TWILIO_API_KEY"),
		"sub": os.Getenv("TWILIO_ACCOUNT_SID"),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 1).Unix(),
		"grants": map[string]interface{}{
			"identity": identity,
			"chat": map[string]interface{}{},
		},
	})


	tokenString, err := token.SignedString([]byte(os.Getenv("TWILIO_API_SECRET")))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate token: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func createConversationHandler(w http.ResponseWriter, r *http.Request) {
	// Create a new conversation in the default service
	convoParams := &conversations.CreateConversationParams{}
	convoParams.SetFriendlyName("Support Chat")
	convo, err := client.ConversationsV1.CreateConversation(convoParams)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create conversation: %v", err), http.StatusInternalServerError)
		return
	}

	// Add customer
	customerParams := &conversations.CreateConversationParticipantParams{}
	customerParams.SetIdentity("customer")
	_, err = client.ConversationsV1.CreateConversationParticipant(*convo.Sid, customerParams)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add customer: %v", err), http.StatusInternalServerError)
		return
	}

	// Add agent
	agentParams := &conversations.CreateConversationParticipantParams{}
	agentParams.SetIdentity("agent")
	_, err = client.ConversationsV1.CreateConversationParticipant(*convo.Sid, agentParams)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add agent: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"conversationSid": *convo.Sid})
}


func serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" || path == "/customer" {
		http.ServeFile(w, r, "static/customer.html")
	} else if path == "/agent" {
		http.ServeFile(w, r, "static/agent.html")
	} else if path == "/styles.css" {
		http.ServeFile(w, r, "static/styles.css")
	} else {
		http.ServeFile(w, r, "static"+path)
	}
}