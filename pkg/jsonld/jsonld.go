package jsonld

import (
	"github.com/go-ap/jsonld"
	"github.com/labstack/echo/v5"
)

const contextKey = "@context"

var defaultContext = map[string]any{
	contextKey: []any{
		"https://www.w3.org/ns/activitystreams",
		"https://w3id.org/security/v1",
	},
}

func DefautContext() map[string]any {
	return defaultContext
}

func defautContextGoAP() []jsonld.Collapsible {
	var ctx []jsonld.Collapsible
	for _, e := range DefautContext()[contextKey].([]any) {
		switch v := e.(type) {
		case string:
			ctx = append(ctx, jsonld.IRI(v))
		}
	}
	return ctx
}

func JSONLDMarshal(obj any, ctx ...jsonld.Collapsible) ([]byte, error) {
	if len(ctx) == 0 {
		ctx = defautContextGoAP()
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
