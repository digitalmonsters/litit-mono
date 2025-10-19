package kafka_listener

// todo later
//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"github.com/docker/distribution/configuration"
//	"github.com/gocql/gocql"
//	"github.com/rs/zerolog/log"
//	"github.com/scylladb/gocqlx/v2"
//	"github.com/segmentio/kafka-go"
//	"os"
//	"testing"
//	"time"
//)
//
//var cfg *configuration.Configuration
//var session gocqlx.Session
//
//func TestMain(m *testing.M) {
//
//	config, err := configuration.GetConfiguration()
//
//	if err != nil {
//		log.Panic().Err(err).Msg("cannot initialize configuration")
//		return
//	}
//
//	cfg = config
//
//	//scylla start
//
//	cluster := gocql.NewCluster(cfg.Scylla.Hosts...)
//	cluster.Keyspace = cfg.Scylla.Keyspace
//	cluster.Authenticator = gocql.PasswordAuthenticator{
//		Username: cfg.Scylla.UserName,
//		Password: cfg.Scylla.Password,
//	}
//
//	s, err := gocqlx.WrapSession(cluster.CreateSession())
//	session = s
//
//	if err != nil {
//		log.Panic().Err(err).Msg("cannot initialize scylla session")
//		return
//	}
//
//	defer session.Close()
//
//	structs.InitTables(cfg.Scylla.Keyspace)
//
//	code := m.Run()
//	os.Exit(code)
//}
//
//func TestListener(t *testing.T) {
//	//testListener(t)
//}
//
//func testListener(t *testing.T) {
//	ctx, _ := context.WithCancel(context.Background())
//
//	b := NewBatchListener(*cfg.Kafka, structs2.NewCommand("like", func(request ...kafka.Message) (interface{}, error) {
//
//		fmt.Println(fmt.Printf("Processed %v number of messages in batch", len(request)))
//
//		for _, m := range request {
//			fmt.Println(fmt.Sprintf("Messages with offset %v", m.Offset))
//		}
//		return nil, nil
//	}, false), ctx, 100*time.Second, 1000)
//
//	s := NewSingleListener(*cfg.Kafka, structs2.NewCommand("like", func(request ...kafka.Message) (interface{}, error) {
//		fmt.Println(fmt.Printf("Processed %v number of single message", len(request)))
//
//		for _, m := range request {
//			fmt.Println(fmt.Sprintf("Single sessage with offset %v", m.Offset))
//		}
//		return nil, nil
//	}, false), ctx)
//
//	go func() {
//		_, err := s.Connect()
//
//		if err != nil {
//			t.Fatal(err)
//		}
//
//		s.Listen()
//
//		defer s.Close()
//
//	}()
//
//	go func() {
//		_, err := b.Connect()
//
//		if err != nil {
//			t.Fatal(err)
//		}
//
//		b.Listen()
//
//		defer b.Close()
//
//	}()
//
//	go func() {
//		w := &kafka.Writer{
//			Addr:         kafka.TCP(fmt.Sprintf("%s:%v", cfg.Kafka.Host, cfg.Kafka.Port)),
//			Topic:        cfg.Kafka.ConsumerTopic,
//			Balancer:     &kafka.LeastBytes{},
//			RequiredAcks: kafka.RequireNone,
//		}
//
//		for ctx.Err() == nil {
//			time.Sleep(4 * time.Second)
//
//			messages := generateMessages(t, 10)
//
//			err := w.WriteMessages(context.Background(), messages...)
//
//			if err != nil {
//				t.Fatal(err)
//			}
//		}
//
//		if err := w.Close(); err != nil {
//			t.Fatal(err)
//		}
//	}()
//
//	<-ctx.Done()
//}
//
//func generateMessages(t *testing.T, count int) []kafka.Message {
//	var resp = structs.KafkaResponse{
//		OwnerId:    "b2b5f5a9-5e20-48cb-8d13-1cec2b9dfda4",
//		ContentId:  "1316c797-e2dd-493d-9a74-7a746eb5fbc3",
//		Deleted:    false,
//		ViewerId:   "45025428-73e9-46a6-81e9-086ab5d75ea0",
//		CategoryId: "8ebe643c-028e-4b89-8d3c-c00a576fa8b8",
//		HashtagId:  "8c08782c-28eb-47ef-bde7-7d17f83ed034",
//		CountryId:  "bcb2692a-6d73-448f-986c-eef00af5e0fc",
//		Date:       time.Now().UTC(),
//	}
//
//	js, err := json.Marshal(&resp)
//
//	if err != nil {
//		t.Fatal(err)
//	}
//	var messages []kafka.Message
//
//	for i := 0; i < count; i++ {
//		messages = append(messages, kafka.Message{
//			Time:  time.Now().Add(time.Duration(i) * time.Millisecond).Truncate(time.Millisecond),
//			Value: js,
//		})
//	}
//
//	return messages
//}
