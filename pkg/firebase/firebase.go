package firebase

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

// FirebaseClient wraps the Firebase Messaging client
type FirebaseClient struct {
	client *messaging.Client
}

var (
	fbClient *FirebaseClient
	once     sync.Once
)

// Initialize initializes the Firebase App and Messaging client
func Initialize(ctx context.Context, serviceAccountJSON string) *FirebaseClient {
	once.Do(func() {
		opt := option.WithCredentialsJSON([]byte(serviceAccountJSON))
		app, err := firebase.NewApp(ctx, nil, opt)
		if err != nil {
			log.Fatalf("Failed to initialize Firebase app: %v", err)
		}

		// Create a Messaging client
		messagingClient, err := app.Messaging(ctx)
		if err != nil {
			log.Fatalf("Failed to initialize Firebase Messaging client: %v", err)
		}

		fbClient = &FirebaseClient{
			client: messagingClient,
		}
		log.Println("Firebase client initialized successfully")
	})

	return fbClient
}

// SendNotification sends a push notification to a specific device token
func (f *FirebaseClient) SendNotification(ctx context.Context, deviceToken string, platform string, title string, imageUrl string, body, notificationType string, data map[string]string) (string, error) {
	if data == nil {
		data = make(map[string]string)
	}
	data["sound"] = "Sweet.mp3"

	customDataJSON, _ := json.Marshal(data)
	oneHour := time.Duration(1) * time.Hour
	message := &messaging.Message{
		Token: deviceToken,
		Data: map[string]string{
			"custom_data": string(customDataJSON),
			"type":        notificationType,
			"title":       title,
			"body":        body,
		},
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Android: &messaging.AndroidConfig{
			TTL: &oneHour,
			Notification: &messaging.AndroidNotification{
				ImageURL: imageUrl,
			},
		},
		APNS: &messaging.APNSConfig{
			FCMOptions: &messaging.APNSFCMOptions{
				ImageURL: imageUrl,
			},
		},
		Topic: "fcm_default_channel",
	}

	response, err := f.client.Send(ctx, message)
	if err != nil {
		log.Printf("Error sending notification: %v", err)
		return "", err
	}

	log.Printf("Successfully sent message: %s", response)
	return response, nil
}
