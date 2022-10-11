package awsses

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/toastate/toastainer/internal/api/dynamicroutes"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/email/awsses/awssns"
	"github.com/toastate/toastainer/internal/utils"
)

type seshandler struct {
	client *ses.SES
}

type snsmessage struct {
	NotificationType string `json:"notificationType"`
	Mail             struct {
		CommonHeaders struct {
			To []string `json:"to"`
		} `json:"commonHeaders"`
	} `json:"mail"`
}

func NewHandler() (*seshandler, error) {
	err := awssns.Init()
	if err != nil {
		return nil, err
	}

	awsCfg := aws.NewConfig().WithRegion(config.EmailProvider.AWSSES.Region)
	awsCfg = awsCfg.WithCredentials(credentials.NewStaticCredentials(config.EmailProvider.AWSSES.PubKey, config.EmailProvider.AWSSES.PrivKey, ""))

	sess, err := session.NewSession(awsCfg)
	if err != nil {
		panic(err)
	}
	client := ses.New(sess, awsCfg)

	dynamicroutes.AddUnauthedDynamicRoute("sys/rejectemail", func(w http.ResponseWriter, r *http.Request) {
		snsInput := &awssns.Payload{}

		err := json.NewDecoder(r.Body).Decode(&snsInput)
		if err != nil {
			utils.SendErrorAndLog(w, "could not read body", "invalidBody", 400, "sys/rejectemail", err)
			return
		}

		err = snsInput.VerifyPayload()
		if err != nil {
			utils.SendErrorAndLog(w, "invalid signature", "invalidBody", 400, "sys/rejectemail", err)
			return
		}

		switch snsInput.Type {
		case "SubscriptionConfirmation":
			_, err = snsInput.Subscribe()
			if err != nil {
				utils.SendErrorAndLog(w, "could not confirm sns topic subscription", "invalid", 400, "sys/rejectemail", err)
				return
			}
		case "UnsubscribeConfirmation":
			_, err = snsInput.Unsubscribe()
			if err != nil {
				utils.SendErrorAndLog(w, "could not confirm sns topic unsubscribe", "invalid", 400, "sys/rejectemail", err)
				return
			}
		case "Notification":
		default:
			utils.SendErrorAndLog(w, "unsupported aws sns message type: "+snsInput.Type, "invalidBody", 400, "sys/rejectemail", err)
			return
		}

		snsMessage := snsmessage{}
		err = json.Unmarshal([]byte(snsInput.Message), &snsMessage)
		if err != nil {
			utils.SendErrorAndLog(w, "could not parse json sns message", "invalidBody", 400, "sys/rejectemail", err)
			return
		}

		if snsMessage.NotificationType == "Delivery" {
			return
		}

		if snsMessage.NotificationType != "Bounce" && snsMessage.NotificationType != "Complaint" {
			utils.SendErrorAndLog(w, "unsupported sns email message notification type: "+snsMessage.NotificationType, "invalidBody", 400, "sys/rejectemail", err)
			return
		}

		for _, v := range snsMessage.Mail.CommonHeaders.To {
			err = objectdb.Client.BlockEmail(v, "")
			if err != nil && err != objectdberror.ErrAlreadyExists {
				utils.Error("msg", "sys/rejectemail", "could not block email in objectdb", v, err)
			}
		}

		return
	})

	return &seshandler{client: client}, nil
}

func (m *seshandler) Send(recipients []string, object, text, html string) error {
	rec := make([]*string, len(recipients))
	for i := 0; i < len(recipients); i++ {
		rec[i] = &recipients[i]
	}

	repto := make([]*string, len(config.EmailProvider.AWSSES.ReplyTo))
	for i := 0; i < len(config.EmailProvider.AWSSES.ReplyTo); i++ {
		repto[i] = &config.EmailProvider.AWSSES.ReplyTo[i]
	}

	_, err := m.client.SendEmail(&ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: rec,
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    &html,
				},
				Text: &ses.Content{
					Data: &text,
				},
			},
			Subject: &ses.Content{
				Data: &object,
			},
		},
		Source:           &config.EmailProvider.AWSSES.SourceEmail,
		ReplyToAddresses: repto,
	})

	return err
}
