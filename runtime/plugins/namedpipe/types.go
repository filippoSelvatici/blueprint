package namedpipe

import "os"

type Invocation struct {
	Method string
	Args   []byte
}

type InvocationResp struct {
	Err  string
	Resp []byte
}

type InvocationOverChannel struct {
	Req        Invocation
	ResultPipe *os.File
}

type InvocationRespOverChannel struct {
	Resp       InvocationResp
	ResultPipe *os.File
}
