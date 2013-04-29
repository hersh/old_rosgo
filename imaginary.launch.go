
type Talker struct {
	Node
	pub chan string
}

// inside ros_tutorials.NewTalker( master ) we have something like this...
func NewTalker(master Master) Talker {
	talker := new( Talker )
	talker.node = master.NewNode( "talker" )

	talker.pub = make( chan string )
	talker.node.Advertise( "chatter", talker.pub )

	return talker
}

type Listener struct {
	Node
	sub chan string
}

func NewListener(master Master) Talker {
	listener := new( Listener )
	listener.node = master.NewNode( "listener" )

	listener.sub = make( chan string )
	listener.node.subscribe( "chatter", listener.sub )

	return listener
}

func main() {
	master := ros.NewMaster()

	talker := ros_tutorials.NewTalker( master )
	talker.setParam( "thing_to_say", "blah blah blah" )
	talker.remapTopic( "chatter", "blah" )

	listener := ros_tutorials.NewListener( master )
	listener.remapTopic( "chatter", "blah" )

	master.Run()
}
