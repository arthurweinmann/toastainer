package awssns

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
)

var certCache map[string]*x509.Certificate
var certCacheMu sync.Mutex

func Init() error {
	certCache = map[string]*x509.Certificate{}

	return nil
}

// Payload contains a single POST from SNS
type Payload struct {
	Message          string `json:"Message"`
	MessageId        string `json:"MessageId"`
	Signature        string `json:"Signature"`
	SignatureVersion string `json:"SignatureVersion"`
	SigningCertURL   string `json:"SigningCertURL"`
	SubscribeURL     string `json:"SubscribeURL"`
	Subject          string `json:"Subject"`
	Timestamp        string `json:"Timestamp"`
	Token            string `json:"Token"`
	TopicArn         string `json:"TopicArn"`
	Type             string `json:"Type"`
	UnsubscribeURL   string `json:"UnsubscribeURL"`
}

// ConfirmSubscriptionResponse contains the XML response of accessing a SubscribeURL
type ConfirmSubscriptionResponse struct {
	XMLName         xml.Name `xml:"ConfirmSubscriptionResponse"`
	SubscriptionArn string   `xml:"ConfirmSubscriptionResult>SubscriptionArn"`
	RequestId       string   `xml:"ResponseMetadata>RequestId"`
}

// UnsubscribeResponse contains the XML response of accessing an UnsubscribeURL
type UnsubscribeResponse struct {
	XMLName   xml.Name `xml:"UnsubscribeResponse"`
	RequestId string   `xml:"ResponseMetadata>RequestId"`
}

// BuildSignature returns a byte array containing a signature usable for SNS verification
func (payload *Payload) BuildSignedContent() ([]byte, error) {
	var builtSignature bytes.Buffer
	signableKeys, err := payload.fieldForSignature()
	if err != nil {
		return nil, err
	}

	reflectedStruct := reflect.Indirect(reflect.ValueOf(payload))
	for i := 0; i < len(signableKeys); i++ {
		field := reflectedStruct.FieldByName(signableKeys[i])
		value := field.String()
		if field.IsValid() && value != "" {
			builtSignature.WriteString(signableKeys[i] + "\n")
			builtSignature.WriteString(value + "\n")
		}
	}

	return builtSignature.Bytes(), nil
}

// VerifyPayload will verify that a payload came from SNS
func (payload *Payload) VerifyPayload() error {
	certURL, err := url.Parse(payload.SigningCertURL)
	if err != nil {
		return err
	}

	if certURL.Scheme != "https" {
		return fmt.Errorf("url should be using https")
	}

	if !strings.HasSuffix(certURL.Host, "amazonaws.com") {
		return fmt.Errorf("certificate is located on an invalid domain: %s", certURL.Host)
	}

	builtSignedContent, err := payload.BuildSignedContent()
	if err != nil {
		return err
	}

	payloadSignature, err := base64.StdEncoding.DecodeString(payload.Signature)
	if err != nil {
		return err
	}

	certCacheMu.Lock()
	cert, ok := certCache[payload.SigningCertURL]
	if !ok {
		var resp *http.Response
		resp, err = http.Get(payload.SigningCertURL)
		if err == nil {
			var body []byte
			body, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err == nil {
				decodedPem, _ := pem.Decode(body)
				if decodedPem == nil {
					err = errors.New("The decoded PEM file was empty!")
				} else {
					cert, err = x509.ParseCertificate(decodedPem.Bytes)
				}
			}
		}
	}
	certCacheMu.Unlock()
	if err != nil {
		return err
	}

	err = cert.CheckSignature(x509.SHA1WithRSA, builtSignedContent, payloadSignature)
	if err != nil {
		return err
	}

	pub := cert.PublicKey.(*rsa.PublicKey)

	h := sha1.New()
	h.Write(builtSignedContent)
	digest := h.Sum(nil)

	err = rsa.VerifyPKCS1v15(pub, crypto.SHA1, digest, payloadSignature)
	if err != nil {
		return err
	}

	return nil
}

// Subscribe will use the SubscribeURL in a payload to confirm a subscription and return a ConfirmSubscriptionResponse
func (payload *Payload) Subscribe() (ConfirmSubscriptionResponse, error) {
	var response ConfirmSubscriptionResponse
	if payload.SubscribeURL == "" {
		return response, errors.New("Payload does not have a SubscribeURL!")
	}

	resp, err := http.Get(payload.SubscribeURL)
	if err != nil {
		return response, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	xmlErr := xml.Unmarshal(body, &response)
	if xmlErr != nil {
		return response, xmlErr
	}
	return response, nil
}

// Unsubscribe will use the UnsubscribeURL in a payload to confirm a subscription and return a UnsubscribeResponse
func (payload *Payload) Unsubscribe() (UnsubscribeResponse, error) {
	var response UnsubscribeResponse
	resp, err := http.Get(payload.UnsubscribeURL)
	if err != nil {
		return response, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	xmlErr := xml.Unmarshal(body, &response)
	if xmlErr != nil {
		return response, xmlErr
	}
	return response, nil
}

func (payload *Payload) fieldForSignature() ([]string, error) {
	if payload.Type == "SubscriptionConfirmation" || payload.Type == "UnsubscribeConfirmation" {
		return []string{"Message", "MessageId", "SubscribeURL", "Timestamp", "Token", "TopicArn", "Type"}, nil
	} else if payload.Type == "Notification" {
		return []string{"Message", "MessageId", "Subject", "Timestamp", "TopicArn", "Type"}, nil
	}

	return nil, fmt.Errorf("unrecognized sns message of type %s", payload.Type)
}
