package clients

import (
	// "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/panyam/slicer/service"
	"google.golang.org/grpc"
	"log"
)

type RpcClient[T any] struct {
	// Note that Addr can be uniquely used to identify a client
	Addr   string
	Conn   *grpc.ClientConn
	Client T
}

func (rc *RpcClient[T]) Close() {
	rc.Conn.Close()
}

func NewRpcClient[T any](addr string, ClientCreator func(grpc.ClientConnInterface) T) (*RpcClient[T], error) {
	// var opts []grpc.DialOption
	// conn, err := grpc.Dial(addr, opts...)
	conn, err := grpc.Dial(addr, grpc.WithInsecure()) // grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
	// grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),

	if err != nil {
		log.Println("Error creating client on address: ", addr, err)
		return nil, err
	}
	out := &RpcClient[T]{
		Addr:   addr,
		Conn:   conn,
		Client: ClientCreator(conn),
	}
	return out, err
}

// What is the flow here?
// ShardClietns need to know where the controller is so they can ask the the controller to konw where service clients are and to get updated on shard changes.
// ServiceClients need to know where the controller is so they can (a) get commanded by the controller to lock down which shards they can serve (otherwise they would
// return 404 etc).  They also need to know the controller to ack any shards that are in "transitioning" state.  - Provide a "GetShardState" method that can be called
// or they subscribe to the controller's to ShardServiceUpdates stream and send heart beats
//
// How does one find the address of a controller?
// This can be either a static LB endpoint - or a sharded endpoint where all shards are known by connecting to one known shard (and using gossip)
// All clients also need to send heartbeats to record into our DB or to the controller
// The controller needs to be a cetnral point - this can change based on gossip
//
// Interfaces:
/**
Scenarios:

A ctrler client exists in every host that is deployed and participates in the gossip ring.  This way leader can change but leader changes have to be notified to everybody and clients have to change.  Here the leaders themselves are going from active to secondary state and idle will reject reuqests to them forcing clients to query for the new leader.  Leaders can be found by looking in the DB - and the client can do this.  The client needs an access to the DB to query state.

A single ctrler can exist which means they can never fail and can always be accessed by static address.

Step 1: Find the leader (DB or static IP)
Step 2: for each shard - leader.GetClient() - do this when needed -
Step 3: (optional) be notified when shard assignment changes - only needed if (2) is infrequent and helps to avoid DB
Step 4: (by shard nly) connect and register as a Shard handler - so controller can assign shards or key ranges to it
				(or move things around)..  How do we handle disconnects/reconnects?  Or what happens if shards connect on different
				addresses - looks like shards must be key aware?  Shards can get into SYNCING state for particular shards so
				controller can either keep that in mind and allow them to go on or put a stop to them.  How long should this be
				tolerated?  If a shard goes offline and ctrl cleansup (say after 10 min timeout) and then shard comes back up?
				shard must get permission first to sync?  If a shard went offline (from ctrl's perspective), but was visible to
				client - what should client expect?   ctrl should be source of truth - if ctrl tells clients a shard is out
				clients should respect this
*/
type CtrlClient interface {
}

type ClientMgr[T any] interface {
	GetClient(entityId string) (*RpcClient[T], error)
}

// This is the ControlService proto
type ShardController interface {
}

type ShardObserver interface {
	// Fisrt step of a shard observer is to observe where the controller is
	// so it can communicate with it.  This should always return the
	// latest primary - there can only be one primary
	DiscoverController() RpcClient[service.ControlServiceClient]
}

type ShardTarget interface {
	ShardObserver
	// data plane service implements this to add/remove shards
	AddShards(shardkey ...string)
	RemoveShards(shardkey ...string)
	GetShardStates(shardkey ...string) map[string]ShardState
}

type ShardClient interface {
	ShardObserver

	// Notifies of shard assignment changes
	ShardsChanged(shardkey ...string)
}
