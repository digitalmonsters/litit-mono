package internal

import (
	"crypto/tls"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
	"time"
)

func GetKafkaDialer(tlsEnabled bool, auth boilerplate.KafkaAuth) (*kafka.Dialer, error) {
	var tlsConfig *tls.Config

	if tlsEnabled {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	if len(auth.Type) == 0 {
		return &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			TLS:       tlsConfig,
		}, nil
	}

	var mech sasl.Mechanism

	if auth.Type == "plain" {
		mech = plain.Mechanism{
			Username: "username",
			Password: "password",
		}

	} else {
		mechanism, err := scram.Mechanism(scram.SHA512, auth.User, auth.Password)

		if err != nil {
			return nil, err
		}
		mech = mechanism
	}

	return &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mech,
		TLS:           tlsConfig,
	}, nil
}
