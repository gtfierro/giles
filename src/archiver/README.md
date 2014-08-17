## TODOs

* need to start storing metadata: perhaps send them in a channel to an open mongo connection?
* how do we implement queries? What is the subset of queries I need to support?
	* 'where' queries for metadata
	* data queries

* flesh out ReadingDB module:
	* `latest(where, limit=1)`
	* `prev(where, ref, limit=1)`: ref is the timestamp
	* `next(where, ref, limit=1)`
	* `data(where, start, end, limit)`
	* `data_uuid(uuids, start, end)`
