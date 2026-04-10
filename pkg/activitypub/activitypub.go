package activitypub

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"regexp"
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

var (
	requiredHeadersPost = []string{
		"(request-target)",
		"host",
		"date",
		"content-type",
		"digest",
	}
	requiredHeadersGet = []string{
		"(request-target)",
		"host",
		"date",
		"accept",
	}
)

func RequiredHeadersPost() []string {
	return requiredHeadersPost
}

func RequiredHeadersGet() []string {
	return requiredHeadersGet
}

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
	signer := httpsig.NewRSASHA256Signer(signingActor.PublicKey.ID.String(), key, RequiredHeadersPost())
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
	signer := httpsig.NewRSASHA256Signer(signingActor.PublicKey.ID.String(), key, RequiredHeadersGet())
	err = signer.Sign(req)
	if err != nil {
		return nil, err
	}
	return new(http.Client).Do(req)
}

func VerifyActivityRequest(req *http.Request, actor *activitypub.Actor) error {
	kgf := httpsig.KeyGetterFunc(func(id string) (any, error) {
		return GetRemotePubkey(actor, id)
	})
	verifier := httpsig.NewVerifier(kgf)
	switch req.Method {
	case "GET":
		verifier.SetRequiredHeaders(RequiredHeadersGet())
	default:
		verifier.SetRequiredHeaders(RequiredHeadersPost())
	}
	_, err := verifier.Verify(req)
	return err
}

func GetRemotePubkey(actor *activitypub.Actor, id string) (pubkey any, err error) {
	pubkey, err = getRemotePubkeyByActor(actor)
	if pubkey != nil {
		return
	}
	pubkey, err = getRemotePubkeyByID(id)
	return
}

func getRemotePubkeyByActor(actor *activitypub.Actor) (any, error) {
	if actor.PublicKey.PublicKeyPem != "" {
		return parsePubkeyPem(actor.PublicKey.PublicKeyPem)
	} else if actor.PublicKey.ID != activitypub.EmptyID {
		return getRemotePubkeyByID(actor.PublicKey.ID.String())
	}
	return nil, errors.New("Failed to get remote public key by actor.")
}

func getRemotePubkeyByID(id string) (any, error) {
	res, err := http.Get(id)
	if err != nil {
		return nil, err
	}
	cont, err := io.ReadAll(res.Body)
	defer func() {
		_ = res.Body.Close()
	}()
	if err != nil {
		return nil, err
	}
	err = res.Body.Close()
	if err != nil {
		return nil, err
	}
	it, err := activitypub.UnmarshalJSON(cont)
	if err == nil {
		actor, err := activitypub.ToActor(it)
		if actor != nil && err == nil {
			return getRemotePubkeyByActor(actor)
		}
	}
	return parsePubkeyPem(string(cont))
}

func ParseKeyId(req *http.Request) (string, error) {
	signature, ok := req.Header["Signature"]
	if !ok || len(signature) != 1 {
		return "", errors.New("Request has no valid signature")
	}
	idReg := regexp.MustCompile(`keyId="(?P<keyId>.*?)"`)
	matches := idReg.FindAllStringSubmatch(signature[0], 1)
	if matches == nil {
		return "", errors.New("Signature of request is not valid")
	}
	return matches[0][idReg.SubexpIndex("keyId")], nil
}

func parsePubkeyPem(s string) (any, error) {
	block, _ := pem.Decode([]byte(s))
	return x509.ParsePKIXPublicKey(block.Bytes)
}
