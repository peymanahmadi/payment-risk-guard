package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/peymanahmadi/payment-risk-guard/internal/adapter/kafkaproducer"
)

var countries = []string{"US", "GB", "DE", "CY", "SG", "JP"}

func main() {
	brokers := flag.String("brokers", envOr("KAFKA_BROKERS", "localhost:9092"), "comma-separated Kafka broker addresses")
	topic := flag.String("topic", envOr("KAFKA_TOPIC", "payments.transactions"), "Kafka topic to publish to")
	numAccounts := flag.Int("accounts", 8, "number of simulated accounts")
	interval := flag.Duration("interval", 500*time.Millisecond, "delay between published events")
	anomalyEvery := flag.Int("anomaly-every", 6, "publish one deliberately anomalous scenario every N ticks")
	flag.Parse()

	producer := kafkaproducer.New(strings.Split(*brokers, ","), *topic)
	defer producer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accountIDs := make([]string, *numAccounts)
	for i := range accountIDs {
		accountIDs[i] = fmt.Sprintf("acc-%02d", i+1)
	}

	log.Printf("simulator publishing to topic=%s brokers=%s (ctrl+c to stop)", *topic, *brokers)

	n := 0
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	publish := func(evt kafkaproducer.TransactionEvent) {
		if err := producer.Publish(ctx, evt); err != nil {
			log.Printf("publish error: %v", err)
		}
	}

	for range ticker.C {
		n++
		acc := accountIDs[rand.Intn(len(accountIDs))]

		if *anomalyEvery > 0 && n%*anomalyEvery == 0 {
			switch rand.Intn(3) {
			case 0:
				evt := amountSpikeEvent(acc)
				log.Printf("[anomaly:amount-spike] account=%s amount=%.2f", evt.AccountID, evt.Amount)
				publish(evt)
			case 1:
				evt := geoJumpEvent(acc)
				log.Printf("[anomaly:geo-jump] account=%s country=%s", evt.AccountID, evt.Country)
				publish(evt)
			default:
				log.Printf("[anomaly:velocity-burst] account=%s", acc)
				for i := 0; i < 6; i++ {
					publish(normalEvent(acc))
					time.Sleep(50 * time.Millisecond)
				}
			}
			continue
		}

		evt := normalEvent(acc)
		log.Printf("[normal] account=%s amount=%.2f country=%s", evt.AccountID, evt.Amount, evt.Country)
		publish(evt)
	}
}

func normalEvent(accountID string) kafkaproducer.TransactionEvent {
	return kafkaproducer.TransactionEvent{
		AccountID:   accountID,
		Amount:      50 + rand.Float64()*150,
		Currency:    "USD",
		Country:     "US",
		IPAddress:   randomIP(),
		PaymentType: "deposit",
		OccurredAt:  time.Now().UTC(),
	}
}

func amountSpikeEvent(accountID string) kafkaproducer.TransactionEvent {
	return kafkaproducer.TransactionEvent{
		AccountID:   accountID,
		Amount:      5000 + rand.Float64()*10000,
		Currency:    "USD",
		Country:     "US",
		IPAddress:   randomIP(),
		PaymentType: "withdrawal",
		OccurredAt:  time.Now().UTC(),
	}
}

func geoJumpEvent(accountID string) kafkaproducer.TransactionEvent {
	return kafkaproducer.TransactionEvent{
		AccountID:   accountID,
		Amount:      80,
		Currency:    "USD",
		Country:     countries[1+rand.Intn(len(countries)-1)],
		IPAddress:   randomIP(),
		PaymentType: "deposit",
		OccurredAt:  time.Now().UTC(),
	}
}

func randomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", 1+rand.Intn(254), rand.Intn(255), rand.Intn(255), 1+rand.Intn(254))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
