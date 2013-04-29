package ros

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"io"
	"net"
	"os"
	"net/rpc"
	"net/url"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() os.Error { return nil }

type RPCClient struct {
	conn net.Conn
	url  *url.URL
}

func (client *RPCClient) RPCCall(methodName string,
	args ...interface{}) (interface{}, *rpc.Fault, *rpc.Error) {
	buf := bytes.NewBufferString("")
	berr := Marshal(buf, methodName, args)
	if berr != nil {
		return nil, nil, berr
	}

	var req http.Request
	req.URL = client.url
	req.Method = "POST"
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.Close = false
	req.Body = nopCloser{buf}
	req.Header = map[string][]string{
		"Content-Type": {"text/xml"},
	}
	req.RawURL = "/RPC"
	req.ContentLength = int64(buf.Len())

	if client.conn == nil {
		var cerr *rpc.Error
		if client.conn, cerr = rpc.Open(client.url); cerr != nil {
			return nil, nil, cerr
		}
	}

	if werr := req.Write(client.conn); werr != nil {
		client.conn.Close()
		return nil, nil, &rpc.Error{Msg: werr.String()}
	}

	reader := bufio.NewReader(client.conn)
	resp, rerr := http.ReadResponse(reader, &req)
	if rerr != nil {
		client.conn.Close()
		return nil, nil, &rpc.Error{Msg: rerr.String()}
	} else if resp == nil {
		rrerr := fmt.Sprintf("ReadResponse for %s returned nil response\n",
			methodName)
		return nil, nil, &rpc.Error{Msg: rrerr}
	}

	_, pval, perr, pfault := Unmarshal(resp.Body)

	if resp.Close {
		resp.Body.Close()
		client.conn = nil
	}

	return pval, pfault, perr
}

func (client *RPCClient) Close() {
	client.conn.Close()
}

func NewClient(host string, port int) (c *RPCClient, err *rpc.Error) {
	conn, url, cerr := rpc.OpenConnURL(host, port)
	if cerr != nil {
		return nil, &rpc.Error{Msg: cerr.String()}
	}

	return &RPCClient{conn: conn, url: url}, nil
}
