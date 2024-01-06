package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/erdemkosk/golang-twitter-timeline-service/internal/models"
	"github.com/segmentio/kafka-go"
)

type KafkaService struct {
	brokers       []string
	topic         string
	consumerGroup string
	tweetChannel  chan models.Tweet
}

func CreateKafkaService() *KafkaService {
	brokers := []string{"localhost:29092"}
	topic := "twitter.twitter.tweet"
	consumerGroup := "twitter"

	config := kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  consumerGroup,
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	}

	tweetChannel := make(chan models.Tweet)

	kafkaService := &KafkaService{
		brokers,
		topic,
		consumerGroup,
		tweetChannel,
	}

	go func() {
		reader := kafka.NewReader(config)
		defer func() {
			if err := reader.Close(); err != nil {
				log.Fatalf("Error closing reader: %v", err)
			}
		}()

		var kafkaMessage struct {
			Schema  map[string]interface{} `json:"schema"`
			Payload map[string]interface{} `json:"payload"`
		}

		for {
			m, err := reader.ReadMessage(context.Background())
			if err != nil {
				break
			}

			if err := json.Unmarshal(m.Value, &kafkaMessage); err != nil {
				log.Printf("Error unmarshalling JSON. Message: %s, Error: %v\n", string(m.Value), err)
				continue
			}

			// Payload kısmını Tweet modeline çözümle
			tweet := models.Tweet{
				ID:     kafkaMessage.Payload["_id"].(string),
				UserId: kafkaMessage.Payload["userid"].(string),
				Tweet:  kafkaMessage.Payload["tweet"].(string),
			}

			kafkaService.tweetChannel <- tweet

			//fmt.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
			//fmt.Printf(tweet.Tweet)
		}
	}()

	return kafkaService
}

func readMessage(reader *kafka.Reader) (kafka.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	return reader.ReadMessage(ctx)
}
