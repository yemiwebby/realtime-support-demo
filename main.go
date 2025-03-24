package main

import (
	"log"
	"net/http"
	"os"
	"realtime-support-demo/handler"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("[ERROR]: Error loading .env file")
	}

	serviceId := os.Getenv("TWILIO_CONVERSATION_SERVICE_SID")
	if serviceId == "" {
		log.Fatal("[ERROR]: TWILIO_CONVERSATION_SERVICE_SID is required")
	}
	
	log.Printf("[INFO]Using Conversation Service SID: %s", serviceId)

	// Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/token", handler.TokenHandler)
	mux.HandleFunc("/create-conversation", handler.CreateConversationHandler)
	mux.HandleFunc("/", handler.ServeStaticFiles)

	log.Println("[INFO]: Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
