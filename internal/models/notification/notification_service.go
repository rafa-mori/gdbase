// Package notification fornece a interface e o modelo para notificações.
package notification

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	gl "github.com/kubex-ecosystem/logz"
)

// Service: NotificationService encapsula a lógica de negócio e envio.
type NotificationService struct {
	repo NotificationRepo
}

func NewNotificationService(repo NotificationRepo) *NotificationService {
	return &NotificationService{repo: repo}
}

// CreateNotification cria a notificação, define status e, se necessário, processa o envio.
func (s *NotificationService) CreateNotification(notification Notification) (Notification, error) {
	notification.Status = "pendente"
	created, err := s.repo.Create(notification)
	if err != nil {
		return Notification{}, err
	}
	// Se a data agendada já passou ou é igual ao momento atual, envia instantaneamente.
	if created.ScheduledAt.Before(time.Now()) || created.ScheduledAt.Equal(time.Now()) {
		// Executa o processamento em uma goroutine para não bloquear o endpoint.
		go s.ProcessNotification(created)
	}
	return created, nil
}

// ProcessNotification decide qual função de envio utilizar conforme o canal.
func (s *NotificationService) ProcessNotification(notification Notification) error {
	var err error
	switch notification.Channel {
	case "email":
		err = s.sendEmail(notification)
	case "whatsapp":
		err = s.sendWhatsApp(notification)
	case "push":
		err = s.sendPush(notification)
	default:
		err = errors.New("canal desconhecido")
	}
	if err == nil {
		notification.Status = "enviado"
		fmt.Printf("Notificação ID %d enviada com sucesso via %s\n", notification.ID, notification.Channel)
	} else {
		notification.Status = "erro"
		fmt.Printf("Erro ao enviar notificação ID %d: %v\n", notification.ID, err)
	}
	// Em um cenário real, você atualizaria o status no banco aqui.
	return err
}

// Função simulada para envio de e-mail.
func (s *NotificationService) sendEmail(notification Notification) error {
	gl.Log("info", "Enviando Email")
	fmt.Printf("Enviando Email: %s - %s\n", notification.Title, notification.Message)
	return nil
}

// Função simulada para envio de notificação push.
func (s *NotificationService) sendPush(notification Notification) error {
	gl.Log("info", "Enviando Push Notification")
	fmt.Printf("Enviando Push Notification: %s - %s\n", notification.Title, notification.Message)
	return nil
}

// Função para envio de mensagem via WhatsApp usando a WhatsApp Cloud API real.
func (s *NotificationService) sendWhatsApp(notification Notification) error {
	// Placeholders para dados sensíveis – substitua pelos seus dados reais.
	const (
		phoneNumberID   = "YOUR_PHONE_NUMBER_ID"   // Seu ID do número do WhatsApp Business
		accessToken     = "YOUR_ACCESS_TOKEN"      // Token de acesso gerado para a WhatsApp Cloud API
		recipientNumber = "RECIPIENT_PHONE_NUMBER" // Número do destinatário no formato internacional (ex.: "5511999999999")
	)

	// Monta o endpoint concatenando o ID e o token.
	endpoint := "https://graph.facebook.com/v15.0/" + phoneNumberID + "/messages?access_token=" + accessToken

	// Constrói o payload para enviar uma mensagem de texto via WhatsApp.
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                recipientNumber,
		"type":              "text",
		"text": map[string]interface{}{
			"body": notification.Message, // Utiliza a mensagem definida na notificação.
		},
	}

	// Serializa o payload para JSON.
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		gl.Log("error", "Erro ao gerar payload JSON")
		return fmt.Errorf("erro ao gerar payload JSON: %v", err)
	}

	// Cria a requisição POST.
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		gl.Log("error", "Erro ao criar requisição HTTP")
		return fmt.Errorf("erro ao criar requisição HTTP: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Configura o client HTTP com timeout.
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		gl.Log("error", "Erro ao enviar requisição HTTP")
		return fmt.Errorf("erro na requisição HTTP: %v", err)
	}
	defer resp.Body.Close()

	// Verifica se a resposta indica sucesso.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		gl.Log("error", "Erro ao enviar mensagem via WhatsApp")
		return fmt.Errorf("falha no envio via WhatsApp, status code: %d", resp.StatusCode)
	}

	// Se necessário, você pode ler a resposta para maiores detalhes.
	// body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Printf("Resposta do WhatsApp: %s\n", string(body))

	return nil
}
