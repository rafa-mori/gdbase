package factory

import (
	"log"

	"github.com/rafa-mori/gdbase/internal/services"
	"github.com/streadway/amqp"
)

type Broker = services.BrokerImpl
type BrokerInfo = services.BrokerInfoLock
type BrokerManager = services.BrokerManager

func NewBrokerService(verbose bool, port string) (*Broker, error) { return services.NewBroker(verbose) }
func NewBrokerManager() *BrokerManager                            { return services.NewBrokerManager() }
func NewBrokerInfo(port string) *BrokerInfo                       { return services.NewBrokerInfo("", port) }

func PublishMessage(queueName string, message string) error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Printf("Erro ao conectar ao RabbitMQ: %s", err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Erro ao abrir um canal: %s", err)
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Erro ao declarar a fila: %s", err)
		return err
	}

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		log.Printf("Erro ao publicar a mensagem: %s", err)
		return err
	}

	log.Printf("Mensagem publicada na fila %s: %s", queueName, message)
	return nil
}
