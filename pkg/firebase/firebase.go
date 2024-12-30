package firebase

import (
	"context"
	"log"
	"sync"

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
func (f *FirebaseClient) SendNotification(ctx context.Context, deviceToken, title, body string, data map[string]string) (string, error) {
	if data == nil {
		data = make(map[string]string)
	}
	data["sound"] = "Sweet.mp3"
	message := &messaging.Message{
		Token: deviceToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	// message := &messaging.Message{
	// 	Token: deviceToken,
	// 	Notification: &messaging.Notification{
	// 		Title: "Hello from Firebase!",
	// 		Body:  "This is a test notification.",
	// 	},
	// }

	response, err := f.client.Send(ctx, message)
	if err != nil {
		log.Printf("Error sending notification: %v", err)
		return "", err
	}

	log.Printf("Successfully sent message: %s", response)
	return response, nil
}
