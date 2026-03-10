package utils

import (
	"github.com/go-ap/activitypub"
	"github.com/go-ap/jsonld"
	"github.com/labstack/echo/v5"
)

func JSONLDMarshal(obj any, ctx ...jsonld.Collapsible) ([]byte, error) {
	if ctx == nil {
		ctx = []jsonld.Collapsible{jsonld.IRI(activitypub.ActivityBaseURI)}
	}
	payload := jsonld.WithContext(ctx...)
	payload.Obj = obj
	b, err := jsonld.Marshal(payload)
	return b, err
}

func JSONLDResponse(c *echo.Context, code int, obj any) error {
	b, err := JSONLDMarshal(obj, nil)
	if err != nil {
		return err
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/activity+json")
	return c.String(code, string(b))
}
