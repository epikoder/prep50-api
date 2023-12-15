package notifications

import (
	"context"
	"fmt"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/Prep50mobileApp/prep50-api/src/models"
	"github.com/Prep50mobileApp/prep50-api/src/services/database"
	"google.golang.org/api/option"
)

func SendNotifications(n *messaging.Notification, fcmTokens ...string) (err error) {
	var ctx = context.Background()
	opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return err
	}

	if len(fcmTokens) == 0 {
		fcms := []models.Fcm{}
		database.DB().Find(&fcms)
		for _, v := range fcms {
			fcmTokens = append(fcmTokens, v.Token)
		}
	}
	message := &messaging.MulticastMessage{
		Notification: n,
		Tokens:       fcmTokens,
	}

	br, err := client.SendMulticast(ctx, message)
	if err != nil {
		return
	}
	fmt.Printf("%d messages were sent successfully\n", br.SuccessCount)
	return
}
