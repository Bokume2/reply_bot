package activitypub

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Bokume2/reply_bot/internal/infrastructure/config"
	"github.com/Bokume2/reply_bot/internal/infrastructure/storage"
	"github.com/Bokume2/reply_bot/internal/interface/schema"
	"github.com/Bokume2/reply_bot/pkg/sig_key"

	"github.com/go-ap/activitypub"
	apErrors "github.com/go-ap/errors"
	"github.com/go-ap/httpsig"
	"github.com/go-ap/jsonld"
	"github.com/labstack/echo/v5"
)

func ResolveActivityPubLink(item activitypub.Item) (activitypub.Item, error) {
	if !item.IsLink() {
		return item, nil
	}

	link := item.GetLink()
	it, err := storage.DataStore().Load(link)
	if err == nil {
		return it, nil
	} else if !apErrors.IsNotFound(err) {
		return nil, err
	}

	actorItem, err := storage.DataStore().Load(schema.UsernameToID(config.BotPreferredUsername()))
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
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = res.Body.Close()
	if err != nil {
		return nil, err
	}
	it, err = activitypub.UnmarshalJSON(body)
	return it, err
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
	key, err := sig_key.ReadPrivKey(sig_key.PKeyPath(signingActor.PreferredUsername.String()))
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
	key, err := sig_key.ReadPrivKey(sig_key.PKeyPath(signingActor.PreferredUsername.String()))
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
