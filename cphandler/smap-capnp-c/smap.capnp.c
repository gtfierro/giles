#include "smap.capnp.h"
/* AUTO GENERATED - DO NOT EDIT */
static const capn_text capn_val0 = {0,""};

Request_ptr new_Request(struct capn_segment *s) {
	Request_ptr p;
	p.p = capn_new_struct(s, 8, 2);
	return p;
}
Request_list new_Request_list(struct capn_segment *s, int len) {
	Request_list p;
	p.p = capn_new_list(s, len, 8, 2);
	return p;
}
void read_Request(struct Request *s, Request_ptr p) {
	capn_resolve(&p.p);
	s->which = (enum Request_which) capn_read16(p.p, 0);
	switch (s->which) {
	case Request_writeData:
		s->writeData.p = capn_getp(p.p, 0, 0);
		break;
	default:
		break;
	}
	s->apikey = capn_get_text(p.p, 1, capn_val0);
}
void write_Request(const struct Request *s, Request_ptr p) {
	capn_resolve(&p.p);
	capn_write16(p.p, 0, s->which);
	switch (s->which) {
	case Request_writeData:
		capn_setp(p.p, 0, s->writeData.p);
		break;
	default:
		break;
	}
	capn_set_text(p.p, 1, s->apikey);
}
void get_Request(struct Request *s, Request_list l, int i) {
	Request_ptr p;
	p.p = capn_getp(l.p, i, 0);
	read_Request(s, p);
}
void set_Request(const struct Request *s, Request_list l, int i) {
	Request_ptr p;
	p.p = capn_getp(l.p, i, 0);
	write_Request(s, p);
}

ReqWriteData_ptr new_ReqWriteData(struct capn_segment *s) {
	ReqWriteData_ptr p;
	p.p = capn_new_struct(s, 0, 1);
	return p;
}
ReqWriteData_list new_ReqWriteData_list(struct capn_segment *s, int len) {
	ReqWriteData_list p;
	p.p = capn_new_list(s, len, 0, 1);
	return p;
}
void read_ReqWriteData(struct ReqWriteData *s, ReqWriteData_ptr p) {
	capn_resolve(&p.p);
	s->messages.p = capn_getp(p.p, 0, 0);
}
void write_ReqWriteData(const struct ReqWriteData *s, ReqWriteData_ptr p) {
	capn_resolve(&p.p);
	capn_setp(p.p, 0, s->messages.p);
}
void get_ReqWriteData(struct ReqWriteData *s, ReqWriteData_list l, int i) {
	ReqWriteData_ptr p;
	p.p = capn_getp(l.p, i, 0);
	read_ReqWriteData(s, p);
}
void set_ReqWriteData(const struct ReqWriteData *s, ReqWriteData_list l, int i) {
	ReqWriteData_ptr p;
	p.p = capn_getp(l.p, i, 0);
	write_ReqWriteData(s, p);
}

Response_ptr new_Response(struct capn_segment *s) {
	Response_ptr p;
	p.p = capn_new_struct(s, 8, 1);
	return p;
}
Response_list new_Response_list(struct capn_segment *s, int len) {
	Response_list p;
	p.p = capn_new_list(s, len, 8, 1);
	return p;
}
void read_Response(struct Response *s, Response_ptr p) {
	capn_resolve(&p.p);
	s->status = (enum StatusCode) capn_read16(p.p, 0);
	s->messages.p = capn_getp(p.p, 0, 0);
}
void write_Response(const struct Response *s, Response_ptr p) {
	capn_resolve(&p.p);
	capn_write16(p.p, 0, (uint16_t) s->status);
	capn_setp(p.p, 0, s->messages.p);
}
void get_Response(struct Response *s, Response_list l, int i) {
	Response_ptr p;
	p.p = capn_getp(l.p, i, 0);
	read_Response(s, p);
}
void set_Response(const struct Response *s, Response_list l, int i) {
	Response_ptr p;
	p.p = capn_getp(l.p, i, 0);
	write_Response(s, p);
}

SmapMessage_ptr new_SmapMessage(struct capn_segment *s) {
	SmapMessage_ptr p;
	p.p = capn_new_struct(s, 0, 6);
	return p;
}
SmapMessage_list new_SmapMessage_list(struct capn_segment *s, int len) {
	SmapMessage_list p;
	p.p = capn_new_list(s, len, 0, 6);
	return p;
}
void read_SmapMessage(struct SmapMessage *s, SmapMessage_ptr p) {
	capn_resolve(&p.p);
	s->path = capn_get_text(p.p, 0, capn_val0);
	s->uuid = capn_get_data(p.p, 1);
	s->readings.p = capn_getp(p.p, 2, 0);
	s->contents = capn_getp(p.p, 3, 0);
	s->properties.p = capn_getp(p.p, 4, 0);
	s->metadata.p = capn_getp(p.p, 5, 0);
}
void write_SmapMessage(const struct SmapMessage *s, SmapMessage_ptr p) {
	capn_resolve(&p.p);
	capn_set_text(p.p, 0, s->path);
	capn_setp(p.p, 1, s->uuid.p);
	capn_setp(p.p, 2, s->readings.p);
	capn_setp(p.p, 3, s->contents);
	capn_setp(p.p, 4, s->properties.p);
	capn_setp(p.p, 5, s->metadata.p);
}
void get_SmapMessage(struct SmapMessage *s, SmapMessage_list l, int i) {
	SmapMessage_ptr p;
	p.p = capn_getp(l.p, i, 0);
	read_SmapMessage(s, p);
}
void set_SmapMessage(const struct SmapMessage *s, SmapMessage_list l, int i) {
	SmapMessage_ptr p;
	p.p = capn_getp(l.p, i, 0);
	write_SmapMessage(s, p);
}

SmapMessage_Reading_ptr new_SmapMessage_Reading(struct capn_segment *s) {
	SmapMessage_Reading_ptr p;
	p.p = capn_new_struct(s, 16, 0);
	return p;
}
SmapMessage_Reading_list new_SmapMessage_Reading_list(struct capn_segment *s, int len) {
	SmapMessage_Reading_list p;
	p.p = capn_new_list(s, len, 16, 0);
	return p;
}
void read_SmapMessage_Reading(struct SmapMessage_Reading *s, SmapMessage_Reading_ptr p) {
	capn_resolve(&p.p);
	s->time = capn_read64(p.p, 0);
	s->data = capn_to_f64(capn_read64(p.p, 8));
}
void write_SmapMessage_Reading(const struct SmapMessage_Reading *s, SmapMessage_Reading_ptr p) {
	capn_resolve(&p.p);
	capn_write64(p.p, 0, s->time);
	capn_write64(p.p, 8, capn_from_f64(s->data));
}
void get_SmapMessage_Reading(struct SmapMessage_Reading *s, SmapMessage_Reading_list l, int i) {
	SmapMessage_Reading_ptr p;
	p.p = capn_getp(l.p, i, 0);
	read_SmapMessage_Reading(s, p);
}
void set_SmapMessage_Reading(const struct SmapMessage_Reading *s, SmapMessage_Reading_list l, int i) {
	SmapMessage_Reading_ptr p;
	p.p = capn_getp(l.p, i, 0);
	write_SmapMessage_Reading(s, p);
}

SmapMessage_Pair_ptr new_SmapMessage_Pair(struct capn_segment *s) {
	SmapMessage_Pair_ptr p;
	p.p = capn_new_struct(s, 0, 2);
	return p;
}
SmapMessage_Pair_list new_SmapMessage_Pair_list(struct capn_segment *s, int len) {
	SmapMessage_Pair_list p;
	p.p = capn_new_list(s, len, 0, 2);
	return p;
}
void read_SmapMessage_Pair(struct SmapMessage_Pair *s, SmapMessage_Pair_ptr p) {
	capn_resolve(&p.p);
	s->key = capn_get_text(p.p, 0, capn_val0);
	s->value = capn_get_text(p.p, 1, capn_val0);
}
void write_SmapMessage_Pair(const struct SmapMessage_Pair *s, SmapMessage_Pair_ptr p) {
	capn_resolve(&p.p);
	capn_set_text(p.p, 0, s->key);
	capn_set_text(p.p, 1, s->value);
}
void get_SmapMessage_Pair(struct SmapMessage_Pair *s, SmapMessage_Pair_list l, int i) {
	SmapMessage_Pair_ptr p;
	p.p = capn_getp(l.p, i, 0);
	read_SmapMessage_Pair(s, p);
}
void set_SmapMessage_Pair(const struct SmapMessage_Pair *s, SmapMessage_Pair_list l, int i) {
	SmapMessage_Pair_ptr p;
	p.p = capn_getp(l.p, i, 0);
	write_SmapMessage_Pair(s, p);
}
