package comm

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"rosgo/xmlrpc"
	"time"
)

type publisherConnection struct {
	host string
	port int
}

type TcpRosSubscriber struct {
	publishers []*publisherConnection
}

func requestTopic(publisher_uri, caller_id, topic string, response_data interface{}) error {
	fmt.Printf("  talking to publisher '%s'\n", publisher_uri)

	pub_rpc := xmlrpc.NewClient(publisher_uri)

	tcpros_string := "TCPROS"
	tcpros := make([]xmlrpc.Value, 1)
	tcpros[0] = xmlrpc.Value{String: &tcpros_string}

	protocols := make([]xmlrpc.Value, 1)
	protocols[0] = xmlrpc.Value{Array: &tcpros}

	params := make([]xmlrpc.Param, 3)
	params[0] = xmlrpc.Param{xmlrpc.Value{String: &caller_id}}
	params[1] = xmlrpc.Param{xmlrpc.Value{String: &topic}}
	params[2] = xmlrpc.Param{xmlrpc.Value{Array: &protocols}}

	return pub_rpc.Call("requestTopic", params, response_data)
}

func Subscribe(topic string, channel interface{}) (*TcpRosSubscriber, error) {
	rpc_client := xmlrpc.NewClient("http://localhost:11311")

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	msg_port := 11317
	msg_uri := fmt.Sprintf("http://%s:%d", hostname, msg_port)

	var response2 xmlrpc.MethodResponse
	err = rpc_client.CallStrings("registerSubscriber",
		[]string{"/chubbles", "/chatter", "rosgraph_msgs/Log", msg_uri},
		&response2)
	if err != nil {
		return nil, err
	}

	var sub TcpRosSubscriber

	array := *response2.Params[0].Val.Array
	code := *array[0].Int
	status := *array[1].String
	fmt.Printf("got response code: %d, status '%s'\n", code, status)
	pub_ptr := array[2].Array
	if pub_ptr == nil {
		// No publishers yet.  TODO: implement updatePublishers() xmlrpc call.
		return &sub, nil
	}
	publishers := *pub_ptr
	caller_id := "/chubbles"
	for i := range publishers {
		publisher_uri := *publishers[i].String
		var response xmlrpc.MethodResponse
		err := requestTopic(publisher_uri, caller_id, "/chatter", &response)
		if err != nil {
			fmt.Printf(" error from requestTopic: %v\n", err)
			continue
		}
		fmt.Printf("response to requestTopic():\n")
		response.Params[0].Val.Print()
		fmt.Printf("woot\n")
		array := response.Params[0].Val.GetArray()
		code, _ := array[0].GetInt()
		status, _ := array[1].GetString()
		fmt.Printf(" requestTopic got response code: %d, status '%s'\n", code, status)
		protocol := array[2].GetArray()
		protocol_name, _ := protocol[0].GetString()
		host, _ := protocol[1].GetString()
		port, _ := protocol[2].GetInt()
		fmt.Printf("  protocol %s %s %d\n", protocol_name, host, port)
		publisher := new(publisherConnection)
		publisher.host = host
		publisher.port = port
		sub.publishers = append(sub.publishers, publisher)
		go publisher.read(channel)
	}
	return &sub, nil
}

func appendInt(slice []byte, val int32) []byte {
	return append(slice, byte(val), byte(val>>8), byte(val>>16), byte(val>>24))
}

func encodeHeader(header map[string]string) []byte {
	buf := make([]byte, 0, 500)
	buf = appendInt(buf, 0) // reserve space for the initial length
	for key, val := range header {
		key_bytes := []byte(key)
		val_bytes := []byte(val)
		length := len(key_bytes) + 1 + len(val_bytes) // 1 is for equals sign
		buf = appendInt(buf, int32(length))
		buf = append(buf, key_bytes...)
		buf = append(buf, byte('='))
		buf = append(buf, val_bytes...)
	}
	length := len(buf) - 4 // Subtract 4 to avoid including size of initial length field in length
	buf[0] = byte(length)
	buf[1] = byte(length >> 8)
	buf[2] = byte(length >> 16)
	buf[3] = byte(length >> 24)
	return buf
}

// fills recv_buffer with a chunk of data starting with a 32-bit
// integer representing the size of the remainder.  Reads the size out
// of recv_buffer before returning, so all the data you see in
// recv_buffer is the payload, not including the size.
func readSizePrefixedChunk(conn net.Conn, recv_buffer *bytes.Buffer) error {
	recv_slice := make([]byte, 500)
	n, err := conn.Read(recv_slice)
	if err != nil {
		return err
	}
	fmt.Printf("read %d bytes, err is %v\n", n, err)
	recv_buffer.Write(recv_slice[:n])
	for recv_buffer.Len() < 4 {
		time.Sleep(time.Millisecond)
		n, err := conn.Read(recv_slice)
		if err != nil {
			return err
		}
		recv_buffer.Write(recv_slice[:n])
		fmt.Printf("read2 %d bytes, err is %v\n", n, err)
	}
	size_slice := recv_buffer.Next(4)
	response_size := int(size_slice[0]) | int(size_slice[1])<<8 | int(size_slice[2])<<16 | int(size_slice[3])<<24
	fmt.Printf("response size is %d\n", response_size)

	for recv_buffer.Len() < response_size {
		time.Sleep(time.Millisecond)
		n, err := conn.Read(recv_slice)
		if err != nil {
			return err
		}
		recv_buffer.Write(recv_slice[:n])
		fmt.Printf("read3 %d bytes, err is %v\n", n, err)
	}
	return nil
}

func (pub *publisherConnection) read(channel interface{}) error {
	message_output_channel := channel.(chan string)
	if message_output_channel == nil {
		return errors.New("channel passed to read() is not a 'chan string'")
	}

	fmt.Printf("read1\n")
	// write subscription header
	// TODO: all these values should be read from appropriate places.
	header := map[string]string{
		"message_definition": "string data",
		"callerid":           "/chubbles",
		"topic":              "/chatter",
		"tcp_nodelay":        "1",
		"md5sum":             "992ce8a1687cec8c8bd883ec73ca41d1",
		"type":               "std_msgs/String",
	}
	serialized_header := encodeHeader(header)
	header_buffer := bytes.NewBuffer(serialized_header)
	fmt.Printf("read() sending %d byte header\n", header_buffer.Len())

	addr := fmt.Sprintf("%s:%d", pub.host, pub.port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("dial(%s) failed: %v\n", addr, err)
		return err
	}
	fmt.Printf("read3\n")
	header_buffer.WriteTo(conn)
	fmt.Printf("read4\n")

	// read subscription reply from publisher
	recv_buffer := new(bytes.Buffer)
	err = readSizePrefixedChunk(conn, recv_buffer)
	if err != nil {
		return err
	}
	for recv_buffer.Len() > 0 {
		size_slice := recv_buffer.Next(4)
		item_size := int(size_slice[0]) | int(size_slice[1])<<8 | int(size_slice[2])<<16 | int(size_slice[3])<<24
		item_slice := recv_buffer.Next(item_size)
		fmt.Printf("  response item: '%s'\n", string(item_slice))
	}
	// loop forever, reading messages from publisher
	for {
		err := readSizePrefixedChunk(conn, recv_buffer)
		if err != nil {
			return err
		}
		if recv_buffer.Len() > 0 {
			size_slice := recv_buffer.Next(4)
			item_size := int(size_slice[0]) | int(size_slice[1])<<8 | int(size_slice[2])<<16 | int(size_slice[3])<<24
			item_slice := recv_buffer.Next(item_size)
			message_output_channel <- string(item_slice)
		}
	}
	return nil
}
