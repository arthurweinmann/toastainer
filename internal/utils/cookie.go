package utils

import (
	"fmt"
	"log"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/http/httpguts"
)

// from response header Set-Cookie
func ReadSetCookies(h http.Header) [][]*http.Cookie {
	lines := h["Set-Cookie"]
	if len(lines) == 0 {
		return nil
	}

	cookies := make([][]*http.Cookie, 0, len(lines))
	var err error

	for i := 0; i < len(lines); i++ {
		spl := strings.Split(lines[i], ";")

		ck := &http.Cookie{}

		for j := 0; j < len(spl); j++ {
			val := strings.Trim(spl[j], " ")
			ind := strings.Index(val, "=")
			switch {
			case ind >= 0 && strings.HasPrefix(val, "Expires"):
				ck.Expires, err = time.Parse(val[ind+1:], "Sat Aug 27 2016 17:07:43 GMT+1000 (AEST)")
				if err != nil {
					fmt.Println("WARNING ReadSetCookies could not parse Expires", val)
				}
			case ind >= 0 && strings.HasPrefix(val, "Max-Age"):
				ck.MaxAge, err = strconv.Atoi(val[ind+1:])
				if err != nil {
					fmt.Println("WARNING ReadSetCookies could not parse Max-Age", val)
				}
			case ind >= 0 && strings.HasPrefix(val, "Domain"):
				ck.Domain = val[ind+1:]
			case ind >= 0 && strings.HasPrefix(val, "Path"):
				ck.Path = val[ind+1:]
			case ind >= 0 && strings.HasPrefix(val, "SameSite"):
				switch val[ind+1:] {
				case "Strict":
					ck.SameSite = http.SameSiteStrictMode
				case "Lax":
					ck.SameSite = http.SameSiteLaxMode
				case "None":
					ck.SameSite = http.SameSiteNoneMode
				default:
					fmt.Println("WARNING ReadSetCookies unknown same site attribute", val)
				}
			case strings.HasPrefix(val, "Secure"):
				ck.Secure = true
			case strings.HasPrefix(val, "HttpOnly"):
				ck.HttpOnly = true
			default:
				switch {
				case ind >= 0:
					ck.Name, ck.Value = val[:ind], val[ind+1:]
				default:
					fmt.Println("WARNING ReadSetCookies unknown value", val)
				}
			}
		}
	}

	return cookies
}

// from request header Cookie
func ReadCookies(h http.Header) [][]*http.Cookie {
	lines := h["Cookie"]
	if len(lines) == 0 {
		return [][]*http.Cookie{}
	}

	cookies := make([][]*http.Cookie, 0, len(lines))
	for _, line := range lines {
		line = textproto.TrimString(line)

		var tmp []*http.Cookie

		var part string
		for len(line) > 0 { // continue since we have rest
			if splitIndex := strings.Index(line, ";"); splitIndex > 0 {
				part, line = line[:splitIndex], line[splitIndex+1:]
			} else {
				part, line = line, ""
			}
			part = textproto.TrimString(part)
			if len(part) == 0 {
				continue
			}
			name, val := part, ""
			if j := strings.Index(part, "="); j >= 0 {
				name, val = name[:j], name[j+1:]
			}
			if !isCookieNameValid(name) {
				continue
			}
			val, ok := parseCookieValue(val, true)
			if !ok {
				continue
			}
			tmp = append(tmp, &http.Cookie{Name: name, Value: val})
		}

		cookies = append(cookies, tmp)
	}
	return cookies
}

func DumpCookie(cks [][]*http.Cookie) []string {
	var ret []string

	for i := 0; i < len(cks); i++ {
		var tmp string
		for j := 0; j < len(cks[i]); j++ {
			if j > 0 {
				tmp += "; "
			}
			tmp += cks[i][j].String()
		}
		ret = append(ret, tmp)
	}

	return ret
}

func parseCookieValue(raw string, allowDoubleQuote bool) (string, bool) {
	// Strip the quotes, if present.
	if allowDoubleQuote && len(raw) > 1 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	for i := 0; i < len(raw); i++ {
		if !validCookieValueByte(raw[i]) {
			return "", false
		}
	}
	return raw, true
}

func isCookieNameValid(raw string) bool {
	if raw == "" {
		return false
	}
	return strings.IndexFunc(raw, isNotToken) < 0
}

func isNotToken(r rune) bool {
	return !httpguts.IsTokenRune(r)
}

var cookieNameSanitizer = strings.NewReplacer("\n", "-", "\r", "-")

func sanitizeCookieName(n string) string {
	return cookieNameSanitizer.Replace(n)
}

// sanitizeCookieValue produces a suitable cookie-value from v.
// https://tools.ietf.org/html/rfc6265#section-4.1.1
// cookie-value      = *cookie-octet / ( DQUOTE *cookie-octet DQUOTE )
// cookie-octet      = %x21 / %x23-2B / %x2D-3A / %x3C-5B / %x5D-7E
//           ; US-ASCII characters excluding CTLs,
//           ; whitespace DQUOTE, comma, semicolon,
//           ; and backslash
// We loosen this as spaces and commas are common in cookie values
// but we produce a quoted cookie-value if and only if v contains
// commas or spaces.
// See https://golang.org/issue/7243 for the discussion.
func sanitizeCookieValue(v string) string {
	v = sanitizeOrWarn("Cookie.Value", validCookieValueByte, v)
	if len(v) == 0 {
		return v
	}
	if strings.IndexByte(v, ' ') >= 0 || strings.IndexByte(v, ',') >= 0 {
		return `"` + v + `"`
	}
	return v
}

func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}

func sanitizeOrWarn(fieldName string, valid func(byte) bool, v string) string {
	ok := true
	for i := 0; i < len(v); i++ {
		if valid(v[i]) {
			continue
		}
		log.Printf("net/http: invalid byte %q in %s; dropping invalid bytes", v[i], fieldName)
		ok = false
		break
	}
	if ok {
		return v
	}
	buf := make([]byte, 0, len(v))
	for i := 0; i < len(v); i++ {
		if b := v[i]; valid(b) {
			buf = append(buf, b)
		}
	}
	return string(buf)
}
