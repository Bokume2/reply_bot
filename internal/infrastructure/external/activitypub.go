package external

import (
	"net/http"
	"reply_bot/internal/utils"
	"strings"

	"github.com/go-ap/httpsig"
	"github.com/go-ap/jsonld"
	"github.com/labstack/echo/v5"
)

func PostActivityPub(pkeyPath, to, body string) (*http.Response, error) {
	req, err := http.NewRequest("POST", to, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(echo.HeaderContentType, "application/activity+json")
	key, err := utils.ReadPrivKey(pkeyPath)
	if err != nil {
		return nil, err
	}
	signer := httpsig.NewRSASHA256Signer("signer", key, nil)
	err = signer.Sign(req)
	if err != nil {
		return nil, err
	}
	return new(http.Client).Do(req)
}

func GetActivityPub(from string) (*http.Response, error) {
	req, err := http.NewRequest("GET", from, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(echo.HeaderAccept, jsonld.ContentType)
	return new(http.Client).Do(req)
}
