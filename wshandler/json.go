package wshandler

import (
	"encoding/json"
	"github.com/gtfierro/giles/archiver"
	"github.com/op/go-logging"
	"io"
	"os"
)

var log = logging.MustGetLogger("wshandler")
var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
var logBackend = logging.NewLogBackend(os.Stderr, "", 0)

func handleJSON(r io.Reader) (decoded archiver.TieredSmapMessage, err error) {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	err = decoder.Decode(&decoded)
	for path, msg := range decoded {
		msg.Path = path
	}
	return
}
