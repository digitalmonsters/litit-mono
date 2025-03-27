package firebase

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/rs/zerolog/log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

// FirebaseClient wraps the Firebase Messaging client.
type FirebaseClient struct {
	client *messaging.Client
}

var (
	fbClient *FirebaseClient
	once     sync.Once
)

// Initialize creates and returns a singleton FirebaseClient.
func Initialize(ctx context.Context, serviceAccountJSON string) *FirebaseClient {
	once.Do(func() {
		opt := option.WithCredentialsJSON([]byte(serviceAccountJSON))
		app, err := firebase.NewApp(ctx, nil, opt)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Firebase app")
		}

		messagingClient, err := app.Messaging(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Firebase Messaging client")
		}

		fbClient = &FirebaseClient{
			client: messagingClient,
		}
		log.Info().Msg("Firebase client initialized successfully")
	})
	return fbClient
}

// SendNotification sends a push notification to a specific device token.
func (f *FirebaseClient) SendNotification(
	ctx context.Context,
	deviceToken, platform, collapseKey, title, imageUrl, body, notificationType string,
	data map[string]string,
) (string, error) {
	if data == nil {
		data = make(map[string]string)
	}
	data["sound"] = "Sweet.mp3"

	customDataJSON, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal custom data")
		return "", err
	}

	message := &messaging.Message{
		Token: deviceToken,
		Data: map[string]string{
			"custom_data": string(customDataJSON),
			"type":        notificationType,
			"title":       title,
			"body":        body,
		},
		// Notification: &messaging.Notification{
		// 	Title: title,
		// 	Body:  body,
		// },
		Android: &messaging.AndroidConfig{
			CollapseKey: collapseKey,
			Notification: &messaging.AndroidNotification{
				ImageURL: imageUrl,
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Category:       notificationType,
					Sound:          "Sweet.mp3",
					MutableContent: true,
				},
			},
			FCMOptions: &messaging.APNSFCMOptions{
				ImageURL: imageUrl,
			},
		},
	}

	response, err := f.client.Send(ctx, message)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send message")
		log.Debug().Msgf("Message details: %+v", message)
		return "", err
	}

	log.Info().Msgf("Successfully sent message: %s", response)
	return response, nil
}
