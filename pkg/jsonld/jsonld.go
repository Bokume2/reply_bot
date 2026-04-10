package jsonld

import (
	"github.com/go-ap/activitypub"
	"github.com/go-ap/jsonld"
	"github.com/labstack/echo/v5"
)

func JSONLDMarshal(obj any, ctx ...jsonld.Collapsible) ([]byte, error) {
	if len(ctx) == 0 {
		ctx = []jsonld.Collapsible{
			jsonld.IRI(activitypub.ActivityBaseURI),
			jsonld.IRI("https://w3id.org/security/v1"),
		}
	}
	payload := jsonld.WithContext(ctx...)
	return payload.Marshal(obj)
}

func JSONLDResponse(c *echo.Context, code int, obj any) error {
	b, err := JSONLDMarshal(obj)
	if err != nil {
		return err
	}
	return c.Blob(code, "application/activity+json", b)
}
