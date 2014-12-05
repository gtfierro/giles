using Go = import "go.capnp";
$Go.package("cphandler");
$Go.import("github.com/gtfierro/giles/cphandler");


# make dictionary a union type of None, List(pair) and dictionary?
@0x9f075567e0861f32;

struct Message {
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
