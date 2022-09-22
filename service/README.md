
This module is a simplified version of our sharding manager.

Goals:

* Easy to manage and operate at low scale (upto 100 nodes)
* Cheap/Efficient/Low overhead

Actors:

Target
  - A host or service target that serves requests and where shards reside.

Shard
  - A key (or key range) that is the partition being assigned to targets for serving.

Seeker
  - Clients that need to access services on certain keys and need to be updated when shard assignments change

Controller
  - The controller/shard-manager that keeps track of shard <-> target assignment

Interface:

Seeker:

Target:
  CanHandleShard(shardkey)

Controller:
  // Called by target to acknowledge it is ready to accept requests
  // on a particular shard.  Note this is only needed on an add and not
  // on a removal
  UpdateShard()
  ShardReady()
  GetTargetsForShard(shard)
  Connect() -> [ShardChanged, AddShard, RemoveShard]


State Machine:

