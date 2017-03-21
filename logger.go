package bundles

import (
	"log"
	"os"
	"io"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stderr, "bundle: ", log.Ldate | log.Ltime | log.Lshortfile)
}

func SetOutput(w io.Writer) {
	logger.SetOutput(w)
}
