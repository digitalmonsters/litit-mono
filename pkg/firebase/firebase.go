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
func (f *FirebaseClient) SendNotification(ctx context.Context, deviceToken, platform string, title, body, notificationType string, data map[string]string) (string, error) {
	if data == nil {
		data = make(map[string]string)
	}
	data["sound"] = "Sweet.mp3"

	customDataJSON, _ := json.Marshal(data)

	message := &messaging.Message{
		Token: deviceToken,
		Data: map[string]string{
			"custom_data": string(customDataJSON),
			"type":        notificationType,
			"title":       title,
			"body":        body,
		},
	}

	if platform == "ios" {
		message.Notification = &messaging.Notification{
			Title: title,
			Body:  body,
		}
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

type BatchMessage struct {
	DeviceToken      string
	Platform         string
	Title            string
	Body             string
	NotificationType string
	Data             map[string]string
}

type BatchNotificationSender struct {
	client        *messaging.Client
	messages      []BatchMessage
	messagesMutex sync.Mutex
	batchSize     int
	batchInterval time.Duration
	lastSentTime  time.Time
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewBatchNotificationSender(ctx context.Context, client *messaging.Client) *BatchNotificationSender {
	ctx, cancel := context.WithCancel(ctx)
	sender := &BatchNotificationSender{
		client:        client,
		messages:      make([]BatchMessage, 0),
		batchSize:     200,
		batchInterval: 5 * time.Second,
		lastSentTime:  time.Now(),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Start background worker
	go sender.startWorker()
	return sender
}

func (b *BatchNotificationSender) startWorker() {
	ticker := time.NewTicker(b.batchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			// Send remaining messages before shutting down
			b.sendBatch()
			return
		case <-ticker.C:
			b.sendBatch()
		}
	}
}

func (b *BatchNotificationSender) Stop() {
	b.cancel()
}

func (b *BatchNotificationSender) QueueNotification(deviceToken, platform string, title, body, notificationType string, data map[string]string) {
	b.messagesMutex.Lock()
	defer b.messagesMutex.Unlock()

	log.Printf("Queueing notification - Device: %s, Platform: %s, Type: %s", deviceToken, platform, notificationType)

	if data == nil {
		data = make(map[string]string)
	}
	data["sound"] = "Sweet.mp3"

	message := BatchMessage{
		DeviceToken:      deviceToken,
		Platform:         platform,
		Title:            title,
		Body:             body,
		NotificationType: notificationType,
		Data:             data,
	}

	b.messages = append(b.messages, message)
	currentQueueSize := len(b.messages)
	log.Printf("Current queue size: %d/%d", currentQueueSize, b.batchSize)

	if currentQueueSize >= b.batchSize {
		log.Printf("Batch size reached (%d). Triggering immediate send...", b.batchSize)
		go b.sendBatch()
	}
}

func (b *BatchNotificationSender) sendBatch() {
	b.messagesMutex.Lock()
	if len(b.messages) == 0 {
		log.Println("No messages to send in batch. Skipping...")
		b.messagesMutex.Unlock()
		return
	}

	batchSize := len(b.messages)
	log.Printf("Preparing to send batch of %d messages", batchSize)

	// Get current batch and reset the queue
	currentBatch := b.messages
	b.messages = make([]BatchMessage, 0)
	b.lastSentTime = time.Now()
	b.messagesMutex.Unlock()

	// Convert batch messages to Firebase messages
	log.Println("Converting batch messages to Firebase format...")
	messages := make([]*messaging.Message, len(currentBatch))
	for i, msg := range currentBatch {
		customDataJSON, err := json.Marshal(msg.Data)
		if err != nil {
			log.Printf("Error marshaling data for device %s: %v", msg.DeviceToken, err)
			continue
		}

		message := &messaging.Message{
			Token: msg.DeviceToken,
			Data: map[string]string{
				"custom_data": string(customDataJSON),
				"type":        msg.NotificationType,
				"title":       msg.Title,
				"body":        msg.Body,
			},
		}

		if msg.Platform == "ios" {
			message.Notification = &messaging.Notification{
				Title: msg.Title,
				Body:  msg.Body,
			}
		}

		messages[i] = message
	}

	log.Printf("Sending batch of %d messages to Firebase...", len(messages))
	startTime := time.Now()

	// Send batch
	response, err := b.client.SendAll(context.Background(), messages)
	if err != nil {
		log.Printf("Error sending batch notifications: %v", err)
		return
	}

	duration := time.Since(startTime)
	log.Printf("Batch sending completed in %v", duration)
	log.Printf("Batch notification results - Total: %d, Success: %d, Failure: %d",
		len(messages), response.SuccessCount, response.FailureCount)

	// Log failures in detail if any
	if response.FailureCount > 0 {
		log.Println("Failed notifications details:")
		for i, resp := range response.Responses {
			if !resp.Success {
				log.Printf("Failed - Device: %s, Error: %v",
					currentBatch[i].DeviceToken, resp.Error)
			}
		}
	}
}

// Add this method to your FirebaseClient struct
func (f *FirebaseClient) NewBatchNotificationSender(ctx context.Context) *BatchNotificationSender {
	return NewBatchNotificationSender(ctx, f.client)
}
