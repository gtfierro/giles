package httphandler

import (
	"encoding/json"
	//"errors"
	//simplejson "github.com/bitly/go-simplejson"
	"github.com/gtfierro/giles/archiver"
	"github.com/pquerna/ffjson/ffjson"
	//"gopkg.in/mgo.v2/bson"
	"io"
	//"strconv"
)

var decoder = ffjson.NewDecoder()

func ffhandleJSON(r io.Reader) (archiver.TieredSmapMessage, error) {
	var res archiver.TieredSmapMessage
	err := decoder.DecodeReader(r, &res)
	for path, msg := range res {
		msg.Path = path
	}
	return res, err
}

func handleJSON(r io.Reader) (decoded archiver.TieredSmapMessage, err error) {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	err = decoder.Decode(&decoded)
	for path, msg := range decoded {
		if msg == nil {
			continue
		}
		msg.Path = path
	}
	return
}
