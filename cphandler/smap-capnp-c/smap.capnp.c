#include "smap.capnp.h"
/* AUTO GENERATED - DO NOT EDIT */
static const capn_text capn_val0 = {0,""};

Message_ptr new_Message(struct capn_segment *s) {
	Message_ptr p;
	p.p = capn_new_struct(s, 0, 6);
	return p;
}
Message_list new_Message_list(struct capn_segment *s, int len) {
	Message_list p;
	p.p = capn_new_list(s, len, 0, 6);
	return p;
}
void read_Message(struct Message *s, Message_ptr p) {
	capn_resolve(&p.p);
	s->path = capn_get_text(p.p, 0, capn_val0);
	s->uuid = capn_get_data(p.p, 1);
	s->readings.p = capn_getp(p.p, 2, 0);
	s->contents = capn_getp(p.p, 3, 0);
	s->properties.p = capn_getp(p.p, 4, 0);
	s->metadata.p = capn_getp(p.p, 5, 0);
}
void write_Message(const struct Message *s, Message_ptr p) {
	capn_resolve(&p.p);
	capn_set_text(p.p, 0, s->path);
	capn_setp(p.p, 1, s->uuid.p);
	capn_setp(p.p, 2, s->readings.p);
	capn_setp(p.p, 3, s->contents);
	capn_setp(p.p, 4, s->properties.p);
	capn_setp(p.p, 5, s->metadata.p);
}
void get_Message(struct Message *s, Message_list l, int i) {
	Message_ptr p;
	p.p = capn_getp(l.p, i, 0);
	read_Message(s, p);
}
void set_Message(const struct Message *s, Message_list l, int i) {
	Message_ptr p;
	p.p = capn_getp(l.p, i, 0);
	write_Message(s, p);
}

Message_Reading_ptr new_Message_Reading(struct capn_segment *s) {
	Message_Reading_ptr p;
	p.p = capn_new_struct(s, 16, 0);
	return p;
}
Message_Reading_list new_Message_Reading_list(struct capn_segment *s, int len) {
	Message_Reading_list p;
	p.p = capn_new_list(s, len, 16, 0);
	return p;
}
void read_Message_Reading(struct Message_Reading *s, Message_Reading_ptr p) {
	capn_resolve(&p.p);
	s->time = capn_read64(p.p, 0);
	s->data = capn_to_f64(capn_read64(p.p, 8));
}
void write_Message_Reading(const struct Message_Reading *s, Message_Reading_ptr p) {
	capn_resolve(&p.p);
	capn_write64(p.p, 0, s->time);
	capn_write64(p.p, 8, capn_from_f64(s->data));
}
void get_Message_Reading(struct Message_Reading *s, Message_Reading_list l, int i) {
	Message_Reading_ptr p;
	p.p = capn_getp(l.p, i, 0);
	read_Message_Reading(s, p);
}
void set_Message_Reading(const struct Message_Reading *s, Message_Reading_list l, int i) {
	Message_Reading_ptr p;
	p.p = capn_getp(l.p, i, 0);
	write_Message_Reading(s, p);
}

Message_Pair_ptr new_Message_Pair(struct capn_segment *s) {
	Message_Pair_ptr p;
	p.p = capn_new_struct(s, 0, 2);
	return p;
}
Message_Pair_list new_Message_Pair_list(struct capn_segment *s, int len) {
	Message_Pair_list p;
	p.p = capn_new_list(s, len, 0, 2);
	return p;
}
void read_Message_Pair(struct Message_Pair *s, Message_Pair_ptr p) {
	capn_resolve(&p.p);
	s->key = capn_get_text(p.p, 0, capn_val0);
	s->value = capn_get_text(p.p, 1, capn_val0);
}
void write_Message_Pair(const struct Message_Pair *s, Message_Pair_ptr p) {
	capn_resolve(&p.p);
	capn_set_text(p.p, 0, s->key);
	capn_set_text(p.p, 1, s->value);
}
void get_Message_Pair(struct Message_Pair *s, Message_Pair_list l, int i) {
	Message_Pair_ptr p;
	p.p = capn_getp(l.p, i, 0);
	read_Message_Pair(s, p);
}
void set_Message_Pair(const struct Message_Pair *s, Message_Pair_list l, int i) {
	Message_Pair_ptr p;
	p.p = capn_getp(l.p, i, 0);
	write_Message_Pair(s, p);
}
