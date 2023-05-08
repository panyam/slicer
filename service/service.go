package service

import (
	"context"
	"fmt"
	gut "github.com/panyam/goutils/utils"
	"github.com/panyam/slicer/db"
	"github.com/panyam/slicer/protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"sort"
	"time"
)

type ControlService struct {
	protos.UnimplementedControlServiceServer
	ControlDB db.ControlDB
	Logger    *log.Logger
}

func NewControlService(ctrldb db.ControlDB, logger *log.Logger) (out *ControlService) {
	out = &ControlService{
		ControlDB: ctrldb,
		Logger:    logger,
	}
	return
}

func (s *ControlService) GetTargets(ctx context.Context, request *protos.GetTargetsRequest) (resp *protos.GetTargetsResponse, err error) {
	resp = nil
	err = nil
	targets, err := s.ControlDB.GetTargets(true, request.Address...)
	if targets == nil {
		resp = nil
		err = status.Error(codes.NotFound, fmt.Sprintf("Target not found: %s", request.Address))
	} else {
		resp = &protos.GetTargetsResponse{
			Targets: gut.Map(targets, TargetToProto),
		}
	}
	return
}

func (s *ControlService) PingTarget(ctx context.Context, request *protos.PingTargetRequest) (resp *protos.Target, err error) {
	s.Logger.Println("Received ping from: ", request.Address)
	targets, err := s.ControlDB.GetTargets(false, request.Address)
	if err == nil {
		if len(targets) == 0 {
			// create it
			newtarget := &db.Target{
				Address:  request.Address,
				PingedAt: time.Now(),
				Status:   "ACTIVE",
			}
			err = s.ControlDB.SaveTarget(newtarget)
			s.Logger.Println("Trying to save: ", newtarget, err)
			if err == nil {
				resp = TargetToProto(newtarget)
			}
		} else {
			targets[0].PingedAt = time.Now()
			err = s.ControlDB.SaveTarget(targets[0])
			if err == nil {
				resp = TargetToProto(targets[0])
			} else {
				s.Logger.Println("SaveTarget error: ", err)
			}
		}
	}
	return resp, err
}

func (s *ControlService) SaveTarget(ctx context.Context, request *protos.SaveTargetRequest) (resp *protos.Target, err error) {
	dbt := TargetFromProto(request.Target)
	err = s.ControlDB.SaveTarget(dbt)
	if err != nil {
		resp = TargetToProto(dbt)
	}
	return resp, err
}

func (s *ControlService) DeleteTargets(ctx context.Context, request *protos.DeleteTargetsRequest) (resp *protos.DeleteTargetsResponse, err error) {
	err = s.ControlDB.DeleteTargets(request.Addresses...)
	return nil, err
}

// Shard specific APIs

func (s *ControlService) GetShard(ctx context.Context, request *protos.GetShardRequest) (resp *protos.GetShardResponse, err error) {
	resp = nil
	err = nil
	shard_targets, err := s.ControlDB.GetShardTargets(request.Shard.Key)
	if err == nil {
		addresses := gut.Map(shard_targets, func(st *db.ShardTarget) string {
			return st.TargetAddress
		})
		var targets []*db.Target
		targets, err = s.ControlDB.GetTargets(false, addresses...)
		if err == nil {
			target_info := make(map[string]*protos.Target)
			for _, target := range targets {
				target_info[target.Address] = TargetToProto(target)
			}
			// Order Shard targets by most recently pinged
			sort.Slice(shard_targets, func(t1 int, t2 int) bool {
				return targets[t2].PingedAt.Sub(targets[t1].PingedAt) < 0
			})
			resp = &protos.GetShardResponse{
				Key:        &protos.ShardKey{Key: request.Shard.Key},
				Targets:    gut.Map(shard_targets, ShardTargetToProto),
				TargetInfo: target_info,
			}
		}
	}
	return
}

func (s *ControlService) SaveShard(ctx context.Context, request *protos.SaveShardRequest) (resp *protos.SaveShardResponse, err error) {
	// TODO - Transactionalize this
	if request.RemoveAll {
		err = s.ControlDB.DeleteShardTargets(request.Key.Key)
	} else {
		err = s.ControlDB.DeleteShardTargets(request.Key.Key, request.Removed...)
	}
	if err == nil {
		for _, address := range request.Added {
			err = s.ControlDB.SaveShardTarget(&db.ShardTarget{
				ShardKey:      request.Key.Key,
				TargetAddress: address,
				Status:        "ACTIVE",
			})
		}
	}
	return
}

func (s *ControlService) DeleteShard(ctx context.Context, request *protos.DeleteShardRequest) (resp *protos.DeleteShardResponse, err error) {
	err = s.ControlDB.DeleteShardTargets(request.Key.Key, request.Addresses...)
	return nil, err
}

func (s *ControlService) Connect(stream protos.ControlService_ConnectServer) (err error) {
	/**
	 * We have two types of connections - either from clients or from producers
	 */
	/**
	 * A new connection has been made that is interested in listening to
	 * updates for one or more screens (which will be sent as sub/unsub
	 * requests)
	 *
	 * Each time a Send is called (say for one or more screens), it needs to
	 * be sent to a bunch of these connections that may be interested in it.
	 *
	 * we have:
	 * Connection (interested in X Screens)
	 * Screen powered by a different topic.
	 * This is truly a hub architecture instead of fan-in or fan-out
	 */

	// 1. First start a stream with a reader so we can read sub/unsub
	// messages on this
	log.Println("Received a new connection....")
	reader := conc.NewReader[*protos.ControlRequest, any](stream.Recv)
	defer reader.Stop()

	// 2. Listen to sub/unsub messages and register/deregister channels
	// where we can listen to events to be sent on the connection
	screenIds := make(map[string]<-chan *protos.ScreenEvent)
	fanIn := conc.NewFanIn[*protos.ScreenEvent](nil)
	var wg sync.WaitGroup
	wg.Add(1)
	connClosedChan := make(chan bool)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-connClosedChan:
				return
			case msg := <-reader.ResultChannel():
				log.Println("Received msg on topic: ", msg)
				if msg.Error != nil {
					// TODO - Handle and return etc
					panic(msg.Error)
				}
				if msg.Value.GetSubscribeRequest() != nil {
					screenId := msg.Value.GetSubscribeRequest().ScreenId
					topic := s.screenClients.Ensure(screenId)
					if _, ok := screenIds[screenId]; !ok {
						// find topic corresponding to this msg's screen add this to the
						// fanout
						clientChan := topic.New()
						// TODO - ensure no duplicates
						screenIds[screenId] = clientChan
						fanIn.Add(clientChan)
					}
				} else if msg.Value.GetUnsubscribeRequest() != nil {
					screenId := msg.Value.GetUnsubscribeRequest().ScreenId
					topic := s.screenClients.Ensure(screenId)
					if clientChan, ok := screenIds[screenId]; ok {
						// TODO - ensure no duplicates
						delete(screenIds, screenId)
						fanIn.Remove(clientChan)
						topic.Remove(clientChan)
					}
				}
			}
			break
		}
	}()

	// Pause till it is closed - in the mean time the publisher will be sending
	// messages with queued up events on this channel
	closed := false
	for !closed {
		select {
		case <-stream.Context().Done():
			closed = true
			connClosedChan <- true
			log.Println("Closing connection to streamer...")
			break
		case event := <-fanIn.Channel():
			log.Println("Received fanIn: ", event)
			msgproto := protos.Message{Events: []*protos.ScreenEvent{event}}
			stream.Send(&msgproto)
			break
		}
	}

	wg.Wait()
	log.Println("All connected readers closed")
	return
}
