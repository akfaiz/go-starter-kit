package lang

import (
	"embed"
	"log"

	"github.com/akfaiz/go-starter-kit/pkg/i18n"
)

//go:embed *.yml
var fs embed.FS

// Init initializes the i18n system by loading language files from the embedded filesystem.
func Init() {
	if err := i18n.LoadWithDefault(fs, "en"); err != nil {
		log.Fatalf("failed to load language files: %v", err)
	}
}
