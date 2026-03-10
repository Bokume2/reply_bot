package schema

import (
	"reply_bot/internal/infrastructure/config"
	"strings"

	"github.com/go-ap/activitypub"
)

func UsernameToID(username string) activitypub.IRI {
	return config.LOCAL_ORIGIN.AddPath("/bots", username)
}

func IDToUsername(id activitypub.IRI) string {
	idStr := id.String()
	idStr = strings.Replace(idStr, config.LOCAL_ORIGIN.AddPath("/bots").String(), "", 1)
	return strings.Trim(idStr, "/")
}
