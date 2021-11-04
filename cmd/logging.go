package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var debug bool
var nativeCmd bool

type PlainFormatter struct {
}

func (f *PlainFormatter) Format(entry *log.Entry) ([]byte, error) {
	if nativeCmd {
		return []byte(fmt.Sprintf("%s\n", entry.Message)), nil
	} else {
		return []byte(fmt.Sprintf("KUBE-LOCK: %s\n", entry.Message)), nil
	}
}
func toggleDebug(cmd *cobra.Command, args []string) {
	if debug {
		log.Info("Debug logs enabled")
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{})
	} else {
		plainFormatter := new(PlainFormatter)
		log.SetFormatter(plainFormatter)
	}
}
