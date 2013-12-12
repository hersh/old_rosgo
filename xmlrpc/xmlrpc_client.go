package xmlrpc

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

type XmlRpcClient struct {
	addr             string
	post_data_holder *bytes.Buffer
	enc              *xml.Encoder
}

// NewClientCodec returns a ClientCodec for communicating with the server
// on the other end of the conn.
func NewClient(addr string) *XmlRpcClient {
	var c XmlRpcClient
	c.addr = addr
	c.post_data_holder = new(bytes.Buffer)
	c.enc = xml.NewEncoder(c.post_data_holder)
	c.enc.Indent("", " ")
	return &c
}

// WriteRequest writes the appropriate header and obj encoded as XML
// to the connection.
func (c *XmlRpcClient) CallStrings(service_method string, string_params []string, response_data interface{}) error {

	params := make([]Param, len(string_params))
	for i := range string_params {
		params[i] = Param{Val: Value{String: &string_params[i]}}
	}
	return c.Call(service_method, params, response_data)
}

func (c *XmlRpcClient) Call(service_method string, params []Param, response_data interface{}) error {
	mc := &MethodCall{Name: service_method, Params: params}
	c.post_data_holder.Reset()
	err := c.enc.Encode(mc)
	if err != nil {
		return err
	}
	response, err := http.Post(c.addr, "text/xml", c.post_data_holder)
	if err != nil {
		return err
	}
	response_bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	fmt.Printf("Call(%s) returned xml:\n%s\n\n", service_method, string(response_bytes))
	return xml.Unmarshal(response_bytes, response_data)
}
