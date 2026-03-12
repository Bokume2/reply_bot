package activitypub

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"reply_bot/internal/infrastructure/config"
	"reply_bot/internal/infrastructure/storage"
	"reply_bot/internal/interface/schema"
	"reply_bot/internal/utils"
	"strings"
	"time"

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

	actorItem, err := storage.DataStore.Load(schema.UsernameToID(config.BOT_PREFERRED_USERNAME))
	if err != nil {
		return nil, err
	}
	instanceActor, err := activitypub.ToActor(actorItem)
	if err != nil {
		return nil, err
	}
	res, err := GetActivityPub(instanceActor, link.String())
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

func PostActivityPub(signingActor *activitypub.Actor, to, body string) (*http.Response, error) {
	req, err := http.NewRequest("POST", to, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(echo.HeaderContentType, "application/activity+json")
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	hash := (sha256.Sum256([]byte(body)))
	digest := "SHA-256=" + base64.StdEncoding.Strict().EncodeToString(hash[:])
	req.Header.Set("Digest", digest)
	key, err := utils.ReadPrivKey(utils.PKeyPath(signingActor.PreferredUsername.String()))
	if err != nil {
		return nil, err
	}
	signer := httpsig.NewRSASHA256Signer(signingActor.PublicKey.ID.String(), key, []string{
		"(request-target)",
		"host",
		"date",
		"content-type",
		"digest",
	})
	err = signer.Sign(req)
	if err != nil {
		return nil, err
	}
	return new(http.Client).Do(req)
}

func GetActivityPub(signingActor *activitypub.Actor, from string) (*http.Response, error) {
	req, err := http.NewRequest("GET", from, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(echo.HeaderAccept, strings.Join([]string{
		jsonld.ContentType,
		"application/activity+json",
	}, ", "))
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	key, err := utils.ReadPrivKey(utils.PKeyPath(signingActor.PreferredUsername.String()))
	if err != nil {
		return nil, err
	}
	signer := httpsig.NewRSASHA256Signer(signingActor.PublicKey.ID.String(), key, []string{
		"(request-target)",
		"host",
		"date",
		"accept",
	})
	err = signer.Sign(req)
	if err != nil {
		return nil, err
	}
	return new(http.Client).Do(req)
}
