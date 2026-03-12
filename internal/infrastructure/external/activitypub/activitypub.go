package activitypub

import (
	"net/http"
	"reply_bot/internal/infrastructure/storage"
	"reply_bot/internal/utils"
	"strings"

	"github.com/go-ap/activitypub"
	apErrors "github.com/go-ap/errors"
	"github.com/go-ap/httpsig"
	"github.com/go-ap/jsonld"
	"github.com/labstack/echo/v5"
)

func ResolveActivityPubLink(item *activitypub.Item) (*activitypub.Item, error) {
	if !(*item).IsLink() {
		return item, nil
	}

	link := (*item).GetLink()
	it, err := storage.DataStore.Load(link)
	if err == nil {
		return &it, nil
	} else if !apErrors.IsNotFound(err) {
		return nil, err
	}

	res, err := GetActivityPub(link.String())
	if err != nil {
		return nil, err
	}
	buf := make([]byte, res.ContentLength)
	var len int64 = 0
	for len < res.ContentLength {
		l, _ := res.Body.Read(buf[len:])
		len += int64(l)
	}
	err = res.Body.Close()
	if err != nil {
		return nil, err
	}
	it, err = activitypub.UnmarshalJSON(buf[:len])
	return &it, err
}

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
