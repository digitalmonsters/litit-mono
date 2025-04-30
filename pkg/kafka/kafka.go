package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/mail"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

func StartEmailConsumer(
	ctx context.Context,
	topic string,
	groupID string,
	brokerString string,
	emailSvc mail.IEmailService,
	userGoWrapper user_go.IUserGoWrapper,
) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerString},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	go func() {
		defer reader.Close()
		log.Info().Str("topic", topic).Str("group_id", groupID).Msg("Started consumer, listening")

		for {
			select {
			case <-ctx.Done():
				log.Warn().Msg("Context cancelled, stopping consumer")
				return

			default:
				msg, err := reader.ReadMessage(ctx)
				if err != nil {
					log.Error().Err(err).Msg("Error reading message from Kafka")
					time.Sleep(time.Second)
					continue
				}

				var object KafkaMessage
				if err := json.Unmarshal(msg.Value, &object); err != nil {
					log.Error().
						Err(err).
						Bytes("raw_message", msg.Value).
						Msg("Failed to unmarshal Kafka message")
					continue
				}

				switch object.EventType {
				case "email.password.forgot":
					var payload eventsourcing.EmailNotificationPasswordForgotPayload
					if !unmarshalPayload(object, &payload) {
						continue
					}

					log.Info().
						Interface("payload", payload).
						Int64("user_id", payload.UserId).
						Msg("Processing password forgot event")

					user := getUser(ctx, payload.UserId, userGoWrapper)
					if user == nil {
						continue
					}

					cwd, err := os.Getwd()
					if err != nil {
						log.Error().Err(err).Msg("Failed to get current working directory")
						continue
					}

					parentDir := filepath.Dir(cwd)
					templatePath := filepath.Join(parentDir, "templates", "forgot-password.html")

					bodyBytes, err := os.ReadFile(templatePath)
					if err != nil {
						log.Error().Err(err).
							Str("path", templatePath).
							Msg("Failed to read email template")
						continue
					}

					body := string(bodyBytes)
					body = strings.ReplaceAll(body, "{{.Username}}", fmt.Sprint(payload.UserId))
					body = strings.ReplaceAll(body, "{{.FormattedCode}}", fmt.Sprint(payload.Code))
					subject := "Password Reset Request"

					if err := emailSvc.SendGenericHTMLEmail(user.Email, subject, body); err != nil {
						log.Error().Err(err).Str("email", user.Email).Msg("Failed to send email")
					} else {
						log.Info().Str("email", user.Email).Msg("Email sent successfully")
					}

				case "email.confirm.address":
					var payload eventsourcing.EmailNotificationConfirmAddressPayload
					if !unmarshalPayload(object, &payload) {
						continue
					}

					log.Info().
						Interface("payload", payload).
						Int64("user_id", payload.UserId).
						Msg("Processing confirm address event")

				case "email.referral":
					var payload eventsourcing.EmailNotificationReferralPayload
					if !unmarshalPayload(object, &payload) {
						continue
					}

					log.Info().
						Interface("payload", payload).
						Int64("user_id", payload.UserId).
						Msg("Processing referral event")

				case "email.guest.temp_info":
					var payload eventsourcing.EmailNotificationTempGuestInfoPayload
					if !unmarshalPayload(object, &payload) {
						continue
					}

					log.Info().
						Interface("payload", payload).
						Str("email", payload.Email).
						Msg("Processing guest temp info event")

				default:
					log.Warn().
						Str("event_type", string(object.EventType)).
						Msg("Unknown event type")
					continue
				}
			}
		}
	}()
}

func getUser(ctx context.Context, userID int64, wrapper user_go.IUserGoWrapper) *user_go.UserRecord {
	userReq := <-wrapper.GetUsers([]int64{userID}, ctx, false)
	if userReq.Error != nil {
		log.Error().
			Err(userReq.Error.ToError()).
			Int64("user_id", userID).
			Msg("Failed to get user from wrapper")
		return nil
	}

	users := userReq.Response
	if users == nil || users[userID].Email == "" {
		log.Error().
			Int64("user_id", userID).
			Msg("User not found or email missing")
		return nil
	}

	user := users[userID]
	return &user
}

func unmarshalPayload(object KafkaMessage, target interface{}) bool {
	payloadBytes, err := json.Marshal(object.Payload)
	if err != nil {
		log.Error().
			Err(err).
			Str("event_type", string(object.EventType)).
			Msg("Failed to marshal payload map to JSON")
		return false
	}
	if err := json.Unmarshal(payloadBytes, target); err != nil {
		log.Error().
			Err(err).
			Str("event_type", string(object.EventType)).
			Msg("Failed to unmarshal typed payload")
		return false
	}
	return true
}
