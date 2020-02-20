package main

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
)

const eventHubsConnStringEnvVar = "EVENTHUBS_CONNECTION_STRING"
const eventHubsBrokerEnvVar = "EVENTHUBS_BROKER"
const eventHubsTopicEnvVar = "EVENTHUBS_TOPIC"
const eventHubsUsernameEnvVar = "EVENTHUBS_USERNAME"

const timeFormat = "Mon Jan _2 15:04:05 2006"
const format = "{\"message\":\"%s\"}"

func main() {
	brokerList := []string{getEnv(eventHubsBrokerEnvVar)}
	fmt.Println("Event Hubs broker", brokerList)

	producer, err := sarama.NewSyncProducer(brokerList, getConfig())
	if err != nil {
		fmt.Println("Failed to start Sarama producer:", err)
		os.Exit(1)
	}

	eventHubsTopic := getEnv(eventHubsTopicEnvVar)
	fmt.Println("Event Hubs topic", eventHubsTopic)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT)

	close := make(chan int)
	var closed bool

	i := 1
	go func() {
		for !closed {
			select {
			case <-exit:
				fmt.Println("program stopped..")
				closed = true
				close <- 1

			default:
				t := fmt.Sprintf(format, "message-"+ strconv.Itoa(i))
				msg := &sarama.ProducerMessage{Topic: eventHubsTopic, Key: sarama.StringEncoder(strconv.Itoa(rand.Intn(100))), Value: sarama.StringEncoder(t)}
				p, o, err := producer.SendMessage(msg)
				if err != nil {
					fmt.Println("Failed to send msg:", err)
					continue
				}
				i++
				fmt.Printf("sent message %s to partition %d offset %d\n", t, p, o)
			}

			time.Sleep(3 * time.Second) //intentional pause
		}
	}()

	fmt.Println("Waiting for ctrl+c")

	<-close

	err = producer.Close()
	if err != nil {
		fmt.Println("failed to close producer", err)
	} else {
		fmt.Println("closed producer")
	}
}

func getConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Net.DialTimeout = 10 * time.Second

	config.Net.SASL.Enable = true
	config.Net.SASL.User = getEnv(eventHubsUsernameEnvVar)
	config.Net.SASL.Password = getEnv(eventHubsConnStringEnvVar)
	config.Net.SASL.Mechanism = "PLAIN"

	config.Net.TLS.Enable = true
	config.Net.TLS.Config = &tls.Config{
		InsecureSkipVerify: true,
		ClientAuth:         0,
	}
	config.Version = sarama.V1_0_0_0
	config.Producer.Return.Successes = true
	return config
}
func getEnv(envName string) string {
	value := os.Getenv(envName)
	if value == "" {
		fmt.Println("Environment variable " + envName + " is missing")
		os.Exit(1)
	}
	return value
}
