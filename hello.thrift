service Hello {
    EchoResponse echo(1:EchoRequest echo)
    CallHomeResponse callHome(1:CallHomeRequest callHome)
}

struct EchoRequest {
    1: required string message;
    2: required i16 count;
}

struct EchoResponse {
    1: required string message;
    2: required i16 count;
}

struct CallHomeRequest {
    1: required EchoRequest echo;
}

struct CallHomeResponse {
    1: required EchoResponse echo;
}
