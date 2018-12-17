package actors

import (
	"fmt"
	"strings"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/plugin"
	"github.com/quorumcontrol/tupelo/gossip3/messages"
	"github.com/quorumcontrol/tupelo/gossip3/middleware"
)

type system interface {
	GetRandomSyncer() *actor.PID
}

const remoteSyncerPrefix = "remoteSyncer"
const currentPusherKey = "currentPusher"

// Gossiper is the root gossiper
type Gossiper struct {
	middleware.LogAwareHolder

	kind             string
	pids             map[string]*actor.PID
	system           system
	syncersAvailable int64
	validatorClear   bool
	round            uint64

	// it is expected the storageActor is actually
	// fronted with a validator
	// and will also report when its busy or not (validatorClear / validatorWorking)
	storageActor *actor.PID
}

const maxSyncers = 3

func NewGossiperProps(kind string, storage *actor.PID, system system) *actor.Props {
	return actor.FromProducer(func() actor.Actor {
		return &Gossiper{
			kind:             kind,
			pids:             make(map[string]*actor.PID),
			syncersAvailable: maxSyncers,
			storageActor:     storage,
			system:           system,
		}
	}).WithMiddleware(
		middleware.LoggingMiddleware,
		plugin.Use(&middleware.LogPlugin{}),
	)
}

func (g *Gossiper) Receive(context actor.Context) {
	// defer func() {
	// 	if re := recover(); re != nil {
	// 		g.Log.Errorw("recover", "re", re)
	// 		panic(re)
	// 	}
	// }()
	switch msg := context.Message().(type) {
	case *actor.Restarting:
		g.Log.Infow("restarting")
	case *actor.Started:
		g.storageActor.Tell(&messages.SubscribeValidatorWorking{
			Actor: context.Self(),
		})
	case *actor.Terminated:
		// this is for when the pushers stop, we can queue up another push
		if msg.Who.Equal(g.pids[currentPusherKey]) {
			g.Log.Debugw("terminate", "isClear", g.validatorClear)
			if g.validatorClear {
				context.Self().Tell(&messages.DoOneGossip{})
			} else {
				delete(g.pids, currentPusherKey)
			}
			return
		}
		if strings.HasPrefix(msg.Who.GetId(), context.Self().GetId()+"/"+remoteSyncerPrefix) {
			g.Log.Debugw("releasing a new remote syncer")
			g.syncersAvailable++
			return
		}
		panic(fmt.Sprintf("unknown actor terminated: %s", msg.Who.GetId()))
	case *messages.StartGossip:
		g.validatorClear = true
		context.Self().Tell(&messages.DoOneGossip{})
	case *messages.DoOneGossip:
		g.Log.Debugw("gossiping again")
		localsyncer, err := context.SpawnNamed(NewPushSyncerProps(g.kind, g.storageActor), "pushSyncer")
		if err != nil {
			panic(fmt.Sprintf("error spawning: %v", err))
		}
		g.pids[currentPusherKey] = localsyncer

		localsyncer.Tell(&messages.DoPush{
			System: g.system,
		})
	case *messages.GetSyncer:
		g.Log.Debugw("GetSyncer", "remote", context.Sender().GetId())
		if g.syncersAvailable > 0 && g.validatorClear {
			receiveSyncer := context.SpawnPrefix(NewPushSyncerProps(g.kind, g.storageActor), remoteSyncerPrefix)
			context.Watch(receiveSyncer)
			g.syncersAvailable--
			available := &messages.SyncerAvailable{}
			available.SetDestination(messages.ToActorPid(receiveSyncer))
			context.Respond(available)
		} else {
			context.Respond(&messages.NoSyncersAvailable{})
		}
	case *messages.Store:
		context.Forward(g.storageActor)
	case *messages.CurrentState:
		context.Forward(g.storageActor)
	case *messages.Get:
		context.Forward(g.storageActor)
	case *messages.BulkRemove:
		context.Forward(g.storageActor)
	case *messages.ValidatorClear:
		g.validatorClear = true
		_, ok := g.pids[currentPusherKey]
		g.Log.Debugw("validator clear", "doGossip", !ok)
		if !ok {
			context.Self().Tell(&messages.DoOneGossip{})
		}
	case *messages.ValidatorWorking:
		g.Log.Debugw("validator working")
		g.validatorClear = false
	case *messages.Debug:
		actor.NewPID("test", "test")
	}
}
