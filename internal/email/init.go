package email

import (
	"fmt"

	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/email/awsses"
)

var Client interface {
	Send(recipients []string, object, text, html string) error
}

func Init() error {
	err := initTemplates()
	if err != nil {
		return err
	}

	switch config.EmailProvider.Name {
	case "awsses":
		Client, err = awsses.NewHandler()
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("not yet supported email provider: %s", config.EmailProvider.Name)
	}

	return nil
}
