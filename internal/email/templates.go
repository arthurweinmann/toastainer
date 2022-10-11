package email

import (
	"bytes"
	"fmt"
	"html/template"

	_ "embed"

	"github.com/toastate/toastainer/internal/utils"
)

//go:embed signup.template
var signuptemp []byte

//go:embed resetpassword.template
var resetpasswordtemp []byte

func initTemplates() error {
	_, err := template.New("t").Parse(utils.ByteSlice2String(signuptemp))
	if err != nil {
		return fmt.Errorf("signup Email template: %v", err)
	}

	_, err = template.New("t").Parse(utils.ByteSlice2String(resetpasswordtemp))
	if err != nil {
		return fmt.Errorf("ResetPassword Email template: %v", err)
	}

	return nil
}

func SignupTemplate() string {
	return utils.ByteSlice2String(signuptemp)
}

func ResetPasswordTemplate(link string) string {
	t, _ := template.New("t").Parse(utils.ByteSlice2String(resetpasswordtemp))

	buf := new(bytes.Buffer)
	err := t.Execute(buf, struct {
		Link string
	}{
		Link: link,
	})
	if err != nil {
		utils.Error("msg", "ResetPasswordTemplate", err)
	}

	return buf.String()
}
