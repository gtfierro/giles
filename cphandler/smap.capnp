using Go = import "go.capnp";
$Go.package("cphandler");
$Go.import("github.com/gtfierro/giles/cphandler");

@0x9f075567e0861f32;

# base struct for all requests sent to the Giles archiver
struct Request {
	union {
		void	@0 :Void;
		writeData @1 :ReqWriteData;
	}
	apikey	@2 :Text;
}

# struct for writing data to the archiver
struct ReqWriteData {
	messages	@0 :List(SmapMessage);
}

# base struct for all responses received from the Giles archiver
struct Response {
	status	@0	:StatusCode;
	messages	@1 :List(SmapMessage);

}

# based off of error codes by MPA for Quasar
enum StatusCode {
	ok	@0;
	internalError	@1;
}

# translation of sMAP JSON message into capnproto
struct SmapMessage {
    path @0 :Text;
    uuid @1 :Data;
    readings @2 :List(Reading);
    contents @3 :List(Text);
    properties @4 :List(Pair);
    metadata @5 :List(Pair);

    struct Reading {
        time @0 :UInt64;
        data @1 :Float64;
    }

    struct Pair {
        key @0 :Text;
        value @1 :Text;
    }
}
