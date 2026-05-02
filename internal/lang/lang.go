package lang

import (
	"embed"
	"log"

	"github.com/invopop/ctxi18n"
)

//go:embed *.yml
var fs embed.FS

func Init() {
	if err := ctxi18n.LoadWithDefault(fs, "en"); err != nil {
		log.Fatalf("failed to load language files: %v", err)
	}
}
