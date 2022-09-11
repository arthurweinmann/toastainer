package utils

import "strings"

func SplitSlash(url string) []string {
	if url == "" || url == "/" {
		return []string{"/"}
	}

	splitedURL := strings.Split(url, "/")

	if splitedURL[0] == "" {
		splitedURL = splitedURL[1:]
	}

	if splitedURL[len(splitedURL)-1] == "" {
		splitedURL = splitedURL[:len(splitedURL)-1]
	}

	return splitedURL
}

func StripIPPort(ip string) string {
	colon := strings.IndexByte(ip, ':')
	if colon == -1 {
		return ip
	}
	return ip[:colon]
}

func StripPort(hostport string) string {
	colon := strings.IndexByte(hostport, ':')
	if colon == -1 {
		return hostport
	}
	if i := strings.IndexByte(hostport, ']'); i != -1 {
		return strings.TrimPrefix(hostport[:i], "[")
	}
	return hostport[:colon]
}
