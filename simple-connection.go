// This is an example of a ros-like topic connection.
//
// The node struct just has a map of topics.
//
// Each topic is a "connectionList" which has only the list of subscriber channels.
//
// The message transmission uses typed go channels, but they are
// handled generically by using reflect.Value.

// The only thing I'm not that happy with is that it takes two lines
// to subscribe or advertise.  One line is making the typed channel
// and the other is calling the Node function.  The Node function
// can't create the typed channel because there is no runtime type
// registry and no type literals in go.  The way to enable this, if it
// were desired, would be by adding something to the generated
// ros-message code.
	
// However if you do that you need to auto-generate everything that
// returns a typed channel, because go has no generics.  Since I want
// generated messages to live in separate go packages from the node
// type, I can't have Node.subscribe() be generated once for each
// message type.

// So I could have something like:
// var node Node
// point_channel := geometry_msgs.SubscribePoint( node, "topic_name" )
// Not sure that I like that though.

package main

import (
	"fmt"
	"reflect"
	"time"
)

type Point struct {
	x, y, z float32
}

type connectionList struct {
	subscriber_channels []reflect.Value
}

type Node struct {
	topics map[ string ] *connectionList
}

func (node *Node) getOrCreateTopic( topic string ) *connectionList {
	if node.topics == nil {
		node.topics = map[ string ] *connectionList {}
	}
	connections := node.topics[ topic ]
	if connections == nil {
		connections = new( connectionList )
		node.topics[ topic ] = connections
	}
	return connections
}

func (node *Node) advertise( topic string, channel interface{} ) {
	// Nothing really needs to be recorded about the advertisement
	// as yet, instead this just fires up the endless
	// copy-pub-to-subs loop
	connections := node.getOrCreateTopic( topic )

	go func() {
		generic_pub := reflect.ValueOf( channel )
		var msg reflect.Value
		ok := true
		for ; ok; {
			msg, ok = generic_pub.Recv()
			if ok {
				for _, sub_chan := range connections.subscriber_channels {
					sub_chan.Send( msg )
				}
			}
		}
	}()
}

func (node *Node) subscribe( topic string, channel interface{} ) {
	connections := node.getOrCreateTopic( topic )

	if connections.subscriber_channels == nil {
		connections.subscriber_channels = []reflect.Value{ reflect.ValueOf( channel )}
	} else {
		connections.subscriber_channels =
			append( connections.subscriber_channels, reflect.ValueOf( channel ))
	}
}

func chatter( node *Node ) {
	pub := make( chan Point, 1 )
	node.advertise( "topic", pub )

	fmt.Println("chatter")
	var z float32
	z = 1
	for {
		pub <- Point{ 1, 2, z }
		z += 1
		time.Sleep( time.Second )
	}
}

func listen( node *Node ) {
	sub := make( chan Point )
	node.subscribe( "topic", sub )

	fmt.Println("listen")
	var p Point
	for {
		p = <-sub
		fmt.Printf( "%.1f, %.1f, %.1f\n", p.x, p.y, p.z )
	}
}

func main() {
	fmt.Println( "Hello world." )

	var node Node

	go listen( &node )
	go chatter( &node )
	go listen( &node )

	for {
		time.Sleep( time.Second )
	}
}
