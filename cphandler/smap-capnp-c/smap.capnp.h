#ifndef CAPN_9F075567E0861F32
#define CAPN_9F075567E0861F32
/* AUTO GENERATED - DO NOT EDIT */
#include <capn.h>

#if CAPN_VERSION != 1
#error "version mismatch between capn.h and generated code"
#endif


#ifdef __cplusplus
extern "C" {
#endif

struct Message;
struct Message_Reading;
struct Message_Pair;

typedef struct {capn_ptr p;} Message_ptr;
typedef struct {capn_ptr p;} Message_Reading_ptr;
typedef struct {capn_ptr p;} Message_Pair_ptr;

typedef struct {capn_ptr p;} Message_list;
typedef struct {capn_ptr p;} Message_Reading_list;
typedef struct {capn_ptr p;} Message_Pair_list;

struct Message {
	capn_text path;
	capn_data uuid;
	Message_Reading_list readings;
	capn_ptr contents;
	Message_Pair_list properties;
	Message_Pair_list metadata;
};

struct Message_Reading {
	uint64_t time;
	double data;
};

struct Message_Pair {
	capn_text key;
	capn_text value;
};

Message_ptr new_Message(struct capn_segment*);
Message_Reading_ptr new_Message_Reading(struct capn_segment*);
Message_Pair_ptr new_Message_Pair(struct capn_segment*);

Message_list new_Message_list(struct capn_segment*, int len);
Message_Reading_list new_Message_Reading_list(struct capn_segment*, int len);
Message_Pair_list new_Message_Pair_list(struct capn_segment*, int len);

void read_Message(struct Message*, Message_ptr);
void read_Message_Reading(struct Message_Reading*, Message_Reading_ptr);
void read_Message_Pair(struct Message_Pair*, Message_Pair_ptr);

void write_Message(const struct Message*, Message_ptr);
void write_Message_Reading(const struct Message_Reading*, Message_Reading_ptr);
void write_Message_Pair(const struct Message_Pair*, Message_Pair_ptr);

void get_Message(struct Message*, Message_list, int i);
void get_Message_Reading(struct Message_Reading*, Message_Reading_list, int i);
void get_Message_Pair(struct Message_Pair*, Message_Pair_list, int i);

void set_Message(const struct Message*, Message_list, int i);
void set_Message_Reading(const struct Message_Reading*, Message_Reading_list, int i);
void set_Message_Pair(const struct Message_Pair*, Message_Pair_list, int i);

#ifdef __cplusplus
}
#endif
#endif
