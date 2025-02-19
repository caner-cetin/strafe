package cli

import (
	"runtime"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func check(err error) {
	if err != nil {
		var entry = log.NewEntry(log.StandardLogger())
		fields := make(log.Fields)
		var pc, file, no, ok = runtime.Caller(1)
		var details = runtime.FuncForPC(pc)
		if ok {
			fields["function"] = details.Name()
			fields["file"] = file
			fields["line"] = no
		}
		cause := errors.Cause(err)
		if !errors.Is(cause, err) {
			fields["cause"] = cause
		}
		entry = entry.WithFields(fields)
		entry.Fatal(err.Error())
	}

}
