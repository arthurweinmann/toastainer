package utils

import (
	"fmt"
	"net/mail"
	"unicode"
)

func ValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IsAlnumOrHyphen(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' {
			return false
		}
	}
	return true
}

func IsValidPassword(pass string) error {
	var (
		upp, low, num, sym bool
		tot                uint8
	)

	for _, char := range pass {
		switch {
		case unicode.IsUpper(char):
			upp = true
			tot++
		case unicode.IsLower(char):
			low = true
			tot++
		case unicode.IsNumber(char):
			num = true
			tot++
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			sym = true
			tot++
		default:
		}
	}

	if !upp {
		// return fmt.Errorf("your password must contain at least one uppercase letter")
	}

	if !low {
		// return fmt.Errorf("your password must contain at least one lowercase letter")
	}

	if !sym {
		return fmt.Errorf("your password must contain at least one unicode symbol or punctuation")
	}

	if !num {
		return fmt.Errorf("your password must contain at least one number")
	}

	if tot < 8 {
		return fmt.Errorf("your password must contain at least 8 characters")
	}

	return nil
}
