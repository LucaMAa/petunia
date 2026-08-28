package service

import (
	"context"
	"log"
	"os"
	"petunia/internal/repository"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
)

type PushService interface {
	SendToUser(userID string, title, body string, data map[string]string)
	SendToUsers(userIDs []string, title, body string, data map[string]string)
}

type pushService struct {
	client    *messaging.Client
	tokenRepo repository.PushTokenRepository
}

func NewPushService(tokenRepo repository.PushTokenRepository) PushService {
	credPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	if credPath == "" {
		log.Println("[push] FIREBASE_CREDENTIALS_PATH not set — push disabled")
		return &pushService{tokenRepo: tokenRepo}
	}

	if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credPath); err != nil {
		log.Printf("[push] cannot set GOOGLE_APPLICATION_CREDENTIALS: %v — push disabled", err)
		return &pushService{tokenRepo: tokenRepo}
	}

	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Printf("[push] firebase init failed: %v — push disabled", err)
		return &pushService{tokenRepo: tokenRepo}
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Printf("[push] messaging client failed: %v — push disabled", err)
		return &pushService{tokenRepo: tokenRepo}
	}

	return &pushService{client: client, tokenRepo: tokenRepo}
}

func (s *pushService) SendToUser(userID string, title, body string, data map[string]string) {
	s.SendToUsers([]string{userID}, title, body, data)
}

func (s *pushService) SendToUsers(userIDs []string, title, body string, data map[string]string) {
	if s.client == nil {
		return
	}
	tokens, err := s.tokenRepo.FindByUserIDs(userIDs)
	if err != nil || len(tokens) == 0 {
		return
	}

	go func() {
		ctx := context.Background()
		for _, t := range tokens {
			_, err := s.client.Send(ctx, &messaging.Message{
				Fid: t.Token,
				Notification: &messaging.Notification{
					Title: title,
					Body:  body,
				},
				Data: data,
			})
			if err != nil {
				log.Printf("[push] send failed for token %s: %v", t.Token, err)
				if messaging.IsUnregistered(err) {
					_ = s.tokenRepo.Delete(t.Token)
				}
			}
		}
	}()
}
