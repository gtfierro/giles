#ifndef CAPN_9F075567E0861F32
#define CAPN_9F075567E0861F32
/* AUTO GENERATED - DO NOT EDIT */
#include <capn.h>

#if CAPN_VERSION != 1
#error "version mismatch between capn.h and generated code"
#endif

#include "go.capnp.h"

#ifdef __cplusplus
extern "C" {
#endif

struct Request;
struct ReqWriteData;
struct Response;
struct SmapMessage;
struct SmapMessage_Reading;
struct SmapMessage_Pair;

typedef struct {capn_ptr p;} Request_ptr;
typedef struct {capn_ptr p;} ReqWriteData_ptr;
typedef struct {capn_ptr p;} Response_ptr;
typedef struct {capn_ptr p;} SmapMessage_ptr;
typedef struct {capn_ptr p;} SmapMessage_Reading_ptr;
typedef struct {capn_ptr p;} SmapMessage_Pair_ptr;

typedef struct {capn_ptr p;} Request_list;
typedef struct {capn_ptr p;} ReqWriteData_list;
typedef struct {capn_ptr p;} Response_list;
typedef struct {capn_ptr p;} SmapMessage_list;
typedef struct {capn_ptr p;} SmapMessage_Reading_list;
typedef struct {capn_ptr p;} SmapMessage_Pair_list;

enum StatusCode {
	StatusCode_ok = 0,
	StatusCode_internalError = 1
};
enum Request_which {
	Request__void = 0,
	Request_writeData = 1
};

struct Request {
	enum Request_which which;
	union {
		ReqWriteData_ptr writeData;
	};
	capn_text apikey;
};

struct ReqWriteData {
	SmapMessage_list messages;
};

struct Response {
	enum StatusCode status;
	SmapMessage_list messages;
};

struct SmapMessage {
	capn_text path;
	capn_data uuid;
	SmapMessage_Reading_list readings;
	capn_ptr contents;
	SmapMessage_Pair_list properties;
	SmapMessage_Pair_list metadata;
};

struct SmapMessage_Reading {
	uint64_t time;
	double data;
};

struct SmapMessage_Pair {
	capn_text key;
	capn_text value;
};

Request_ptr new_Request(struct capn_segment*);
ReqWriteData_ptr new_ReqWriteData(struct capn_segment*);
Response_ptr new_Response(struct capn_segment*);
SmapMessage_ptr new_SmapMessage(struct capn_segment*);
SmapMessage_Reading_ptr new_SmapMessage_Reading(struct capn_segment*);
SmapMessage_Pair_ptr new_SmapMessage_Pair(struct capn_segment*);

Request_list new_Request_list(struct capn_segment*, int len);
ReqWriteData_list new_ReqWriteData_list(struct capn_segment*, int len);
Response_list new_Response_list(struct capn_segment*, int len);
SmapMessage_list new_SmapMessage_list(struct capn_segment*, int len);
SmapMessage_Reading_list new_SmapMessage_Reading_list(struct capn_segment*, int len);
SmapMessage_Pair_list new_SmapMessage_Pair_list(struct capn_segment*, int len);

void read_Request(struct Request*, Request_ptr);
void read_ReqWriteData(struct ReqWriteData*, ReqWriteData_ptr);
void read_Response(struct Response*, Response_ptr);
void read_SmapMessage(struct SmapMessage*, SmapMessage_ptr);
void read_SmapMessage_Reading(struct SmapMessage_Reading*, SmapMessage_Reading_ptr);
void read_SmapMessage_Pair(struct SmapMessage_Pair*, SmapMessage_Pair_ptr);

void write_Request(const struct Request*, Request_ptr);
void write_ReqWriteData(const struct ReqWriteData*, ReqWriteData_ptr);
void write_Response(const struct Response*, Response_ptr);
void write_SmapMessage(const struct SmapMessage*, SmapMessage_ptr);
void write_SmapMessage_Reading(const struct SmapMessage_Reading*, SmapMessage_Reading_ptr);
void write_SmapMessage_Pair(const struct SmapMessage_Pair*, SmapMessage_Pair_ptr);

void get_Request(struct Request*, Request_list, int i);
void get_ReqWriteData(struct ReqWriteData*, ReqWriteData_list, int i);
void get_Response(struct Response*, Response_list, int i);
void get_SmapMessage(struct SmapMessage*, SmapMessage_list, int i);
void get_SmapMessage_Reading(struct SmapMessage_Reading*, SmapMessage_Reading_list, int i);
void get_SmapMessage_Pair(struct SmapMessage_Pair*, SmapMessage_Pair_list, int i);

void set_Request(const struct Request*, Request_list, int i);
void set_ReqWriteData(const struct ReqWriteData*, ReqWriteData_list, int i);
void set_Response(const struct Response*, Response_list, int i);
void set_SmapMessage(const struct SmapMessage*, SmapMessage_list, int i);
void set_SmapMessage_Reading(const struct SmapMessage_Reading*, SmapMessage_Reading_list, int i);
void set_SmapMessage_Pair(const struct SmapMessage_Pair*, SmapMessage_Pair_list, int i);

#ifdef __cplusplus
}
#endif
#endif
