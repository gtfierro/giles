package cphandler

import (
    "github.com/gtfierro/giles/giles"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("cphandler")
var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
var logBackend = logging.NewLogBackend(os.Stderr, "", 0)

func Handle(a *giles.Archiver) {
    log.Notice("Handling Capn Proto")
}
