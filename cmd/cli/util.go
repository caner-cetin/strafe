package cli

import (
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

// go does not have error handling, so our applications can run faster!
// seriously, fuck this languages error handling
// todo: something more sophisticated
func check(err error) {
	if err != nil {
		log.Fatal(color.RedString(err.Error()))
	}
}
