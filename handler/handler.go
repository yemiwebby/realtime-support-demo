package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client/jwt"
	conversations "github.com/twilio/twilio-go/rest/conversations/v1"
)

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	identity := r.URL.Query().Get("identity")
	if identity == "" {
		http.Error(w, "Identity is required", http.StatusBadRequest)
		return
	}

	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	apiKey := os.Getenv("TWILIO_API_KEY")
	apiSecret := os.Getenv("TWILIO_API_SECRET")
	serviceSid := os.Getenv("TWILIO_CONVERSATION_SERVICE_SID")

	if accountSid == "" || apiKey == "" || apiSecret == "" || serviceSid == "" {
		http.Error(w, "Internal server configuration error", http.StatusInternalServerError)
		return
	}

	params := jwt.AccessTokenParams{
		AccountSid:    accountSid,
		SigningKeySid: apiKey,
		Secret:        apiSecret,
		Identity:     identity,
	}

	jwtToken := jwt.CreateAccessToken(params)
	chatGrant := &jwt.ChatGrant{
		ServiceSid: serviceSid,
	}

	jwtToken.AddGrant(chatGrant)
	token, err := jwtToken.ToJwt()

	if err != nil {
		log.Printf("[INFO]: Failed to generate JWT: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})

}

func CreateConversationHandler(w http.ResponseWriter, r *http.Request) {
	client := twilio.NewRestClient()
	
	const uniqueName = "support-room-001"
	identity := r.URL.Query().Get("identity")

	// Try to fetch conversation
	convo, err := client.ConversationsV1.FetchConversation(uniqueName)
	if err != nil {
		log.Println("[INFO]: No existing conversation found, creating new one")

		createParams := &conversations.CreateConversationParams{}
		createParams.SetUniqueName(uniqueName)
		createParams.SetFriendlyName("Customer Support Room")

		convo, err = client.ConversationsV1.CreateConversation(createParams)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create conversation: %v", err), http.StatusInternalServerError)
			return
		}
	}
	convoSid := *convo.Sid
	log.Printf("[INFO]: Using conversation SID: %s", convoSid)

	if identity != "" {
		params := &conversations.CreateConversationParticipantParams{}
		params.SetIdentity(identity)

		_, err := client.ConversationsV1.CreateConversationParticipant(convoSid, params)
		if err != nil && !strings.Contains(err.Error(), "Participant already exists") {
			log.Printf("[WARNING]: Failed to add participant '%s': %v", identity, err)
		} else {
			log.Printf("[INFO]: Participant '%s' added or already exists", identity)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"conversationSid": convoSid})
}


// Serves frontend files
func ServeStaticFiles(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch path {
	case "/", "/customer":
		http.ServeFile(w, r, "static/customer.html")
	case "/agent":
		http.ServeFile(w, r, "static/agent.html")
	case "/styles.css":
		http.ServeFile(w, r, "static/styles.css")
	case "/chat.js":
		http.ServeFile(w, r, "static/chat.js")
	default:
		http.NotFound(w, r)
	}
}