package template

import (
	"log"
	"path/filepath"
	"reply_bot/internal/infrastructure/config"
	"reply_bot/internal/utils"
	"strings"
	"text/template"
)

var (
	WellKnownNodeInfo string
	NodeInfo          string
)

const templatesDir string = "templates"

func LoadTemplate() {
	// .well-known/nodeifno
	wkniPath := filepath.Join(templatesDir, ".well-known/nodeinfo_tmpl.json")
	wknit, err := template.New("nodeinfo_tmpl.json").ParseFiles(wkniPath)
	if err != nil {
		log.Fatal(err)
	}
	wknibuf := new(strings.Builder)
	err = wknit.Execute(wknibuf, wellKnownNodeInfoTmplData{LocalDomain: config.LOCAL_DOMAIN})
	if err != nil {
		log.Fatal(err)
	}
	WellKnownNodeInfo = wknibuf.String()

	// nodeinfo/2.1
	niPath := filepath.Join(templatesDir, "nodeinfo/2.1_tmpl.json")
	nit, err := template.New("2.1_tmpl.json").ParseFiles(niPath)
	if err != nil {
		log.Fatal(err)
	}
	nibuf := new(strings.Builder)
	err = nit.Execute(nibuf, nodeInfoTmplData{Version: utils.Version()})
	if err != nil {
		log.Fatal(err)
	}
	NodeInfo = nibuf.String()
}

type wellKnownNodeInfoTmplData struct {
	LocalDomain string
}

type nodeInfoTmplData struct {
	Version string
}
