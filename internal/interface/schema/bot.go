package schema

import (
	"reply_bot/internal/infrastructure/config"

	"github.com/go-ap/activitypub"
)

func UsernameToId(username string) activitypub.IRI {
	return config.LOCAL_ORIGIN.AddPath("/bots", username)
}
