package schema

import (
	"strings"

	"github.com/Bokume2/reply_bot/internal/infrastructure/config"

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
