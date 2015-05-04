package stats

import (
	"encoding/json"
	"github.com/gtfierro/giles/archiver"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

/*type gilesStats struct {
	NumRepubClients  int    `json:"num_repub_clients"`
	IncomingMessages uint64 `json:"incoming_counter"`
	PendingWrites    uint64 `json:"pending_writes"`
	TSDBConnections  int    `json:"tsdb_connections"`
}*/

func gilesStatsHandler(a *archiver.Archiver, w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	jsonGilesStats, err := json.Marshal(a.Stats())
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonGilesStats)
}
