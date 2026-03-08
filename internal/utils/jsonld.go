package utils

import (
	"github.com/go-ap/activitypub"
	"github.com/go-ap/jsonld"
	"github.com/labstack/echo/v5"
)

func JSONLDResponse(c *echo.Context, code int, obj any) error {
	payload := jsonld.WithContext(jsonld.IRI(activitypub.ActivityBaseURI))
	payload.Obj = obj
	b, err := jsonld.Marshal(payload)
	if err != nil {
		return err
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/activity+json")
	return c.String(code, string(b))
}
