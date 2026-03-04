package schema

import (
	"fmt"
	"reply_bot/internal/infrastructure/config"

	"github.com/go-ap/activitypub"
)

func UsernameToId(username string) activitypub.IRI {
	return activitypub.IRI(fmt.Sprintf("https://%s/bots/%s", config.LOCAL_DOMAIN, username))
}
