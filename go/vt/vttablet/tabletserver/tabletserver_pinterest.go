package tabletserver

import "flag"

var streamHealthBufferSize = flag.Uint("stream_health_buffer_size", 20, "max streaming health entries to buffer per StreamHealth rpc")
