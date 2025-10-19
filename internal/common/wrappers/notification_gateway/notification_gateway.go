package notification_gateway

import (
    "context"
    "fmt"
    "time"

    "github.com/digitalmonsters/go-common/boilerplate"
    "github.com/digitalmonsters/go-common/common"
    "github.com/digitalmonsters/go-common/eventsourcing"
    "github.com/digitalmonsters/go-common/wrappers"
    "github.com/pkg/errors"
    "github.com/rs/zerolog/log"
    "go.elastic.co/apm"
)

type Wrapper struct {
    baseWrapper    *wrappers.BaseWrapper
    defaultTimeout time.Duration
    apiUrl         string
    serviceName    string
    pushPublisher  *eventsourcing.KafkaEventPublisher
    emailPublisher *eventsourcing.KafkaEventPublisher
}

type INotificationGatewayWrapper interface {
    SendSmsInternal(message string, phoneNumber string, apmTransaction *apm.Transaction, forceLog bool) chan SendSmsMessageResponseChan
    SendEmailInternal(ccAddresses, toAddresses []string, htmlBody, textBody, subject string, apmTransaction *apm.Transaction, forceLog bool) chan SendEmailMessageResponseChan
    EnqueuePushForUser(msg []SendPushRequest, ctx context.Context) chan error
    EnqueueEmail(msg []SendEmailMessageRequest, ctx context.Context) chan error
}

func NewNotificationGatewayWrapper(config boilerplate.WrapperConfig) INotificationGatewayWrapper {
    timeout := 5 * time.Second

    if config.TimeoutSec > 0 {
        timeout = time.Duration(config.TimeoutSec) * time.Second
    }

    if len(config.ApiUrl) == 0 {
        config.ApiUrl = "http://notification-gateway"

        log.Warn().Msgf("Api Url is missing for NotificationGateway. Setting as default : %v", config.ApiUrl)
    }

    w := &Wrapper{
        baseWrapper:    wrappers.GetBaseWrapper(),
        defaultTimeout: timeout,
        apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
        serviceName:    "notification_gateway",
    }

    pushPublisher := config.PushPublisher
    w.pushPublisher = eventsourcing.NewKafkaEventPublisher(
        boilerplate.KafkaWriterConfiguration{
            Hosts:     pushPublisher.Hosts,
            Tls:       pushPublisher.Tls,
            KafkaAuth: pushPublisher.KafkaAuth,
        }, pushPublisher.Topic)

    emailPublisher := config.EmailPublisher
    w.emailPublisher = eventsourcing.NewKafkaEventPublisher(
        boilerplate.KafkaWriterConfiguration{
            Hosts:     emailPublisher.Hosts,
            Tls:       emailPublisher.Tls,
            KafkaAuth: emailPublisher.KafkaAuth,
        }, emailPublisher.Topic)

    return w
}

func (w *Wrapper) EnqueuePushForUser(msg []SendPushRequest, ctx context.Context) chan error {
    ch := make(chan error, 2)

    go func() {
        defer func() {
            close(ch)
        }()

        if w.pushPublisher == nil {
            ch <- errors.New("publisher is nil")
        }
        var i []eventsourcing.IEventData

        for _, m := range msg {
            i = append(i, m)
        }

        if err := w.pushPublisher.Publish(apm.TransactionFromContext(ctx),
            i...); len(err) > 0 {
            ch <- err[0]

            return
        }

        ch <- nil
    }()

    return ch
}

func (w *Wrapper) EnqueueEmail(msg []SendEmailMessageRequest, ctx context.Context) chan error {
    ch := make(chan error, 2)

    go func() {
        defer func() {
            close(ch)
        }()

        if w.emailPublisher == nil {
            ch <- errors.New("publisher is nil")
        }
        var i []eventsourcing.IEventData

        for _, m := range msg {
            i = append(i, m)
        }

        if err := w.emailPublisher.Publish(apm.TransactionFromContext(ctx),
            i...); len(err) > 0 {
            ch <- err[0]

            return
        }

        ch <- nil
    }()

    return ch
}

func (w *Wrapper) SendSmsInternal(message string, phoneNumber string, apmTransaction *apm.Transaction, forceLog bool) chan SendSmsMessageResponseChan {
    respCh := make(chan SendSmsMessageResponseChan, 2)

    respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "SendSmsInternal", SendSmsMessageRequest{
        Message:     message,
        PhoneNumber: phoneNumber,
    }, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

    go func() {
        defer func() {
            close(respCh)
        }()

        resp := <-respChan

        result := SendSmsMessageResponseChan{
            Error: resp.Error,
        }
        respCh <- result
    }()

    return respCh
}

func (w *Wrapper) SendEmailInternal(ccAddresses, toAddresses []string, htmlBody, textBody, subject string, apmTransaction *apm.Transaction, forceLog bool) chan SendEmailMessageResponseChan {
    respCh := make(chan SendEmailMessageResponseChan, 2)

    respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "SendEmailInternal", SendEmailMessageRequest{
        CcAddresses: ccAddresses,
        ToAddresses: toAddresses,
        HtmlBody:    htmlBody,
        TextBody:    textBody,
        Subject:     subject,
    }, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

    go func() {
        defer func() {
            close(respCh)
        }()

        resp := <-respChan

        result := SendEmailMessageResponseChan{
            Error: resp.Error,
        }
        respCh <- result
    }()

    return respCh
}
