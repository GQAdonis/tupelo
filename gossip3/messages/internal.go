//go:generate msgp

package messages

import (
	"github.com/quorumcontrol/differencedigest/ibf"
	extmsgs "github.com/quorumcontrol/tupelo-go-client/gossip3/messages"
)

func init() {
	extmsgs.RegisterMessage(&GetSyncer{})
	extmsgs.RegisterMessage(&SyncDone{})
	extmsgs.RegisterMessage(&NoSyncersAvailable{})
	extmsgs.RegisterMessage(&SyncerAvailable{})
	extmsgs.RegisterMessage(&ProvideStrata{})
	extmsgs.RegisterMessage(&ProvideBloomFilter{})
	extmsgs.RegisterMessage(&RequestKeys{})
	extmsgs.RegisterMessage(&RequestIBF{})
}

type DestinationHolder struct {
	Destination *extmsgs.ActorPID
}

func (dh *DestinationHolder) SetDestination(newDestination *extmsgs.ActorPID) {
	dh.Destination = newDestination
}

func (dh *DestinationHolder) GetDestination() *extmsgs.ActorPID {
	return dh.Destination
}

type GetSyncer struct {
	Kind string
}

func (GetSyncer) TypeCode() int8 {
	return -100
}

type RequestIBF struct {
	Count int
}

func (RequestIBF) TypeCode() int8 {
	return -101
}

type SyncDone struct{}

func (SyncDone) TypeCode() int8 {
	return -102
}

type NoSyncersAvailable struct{}

func (NoSyncersAvailable) TypeCode() int8 {
	return -103
}

type SyncerAvailable struct {
	DestinationHolder
}

func (SyncerAvailable) TypeCode() int8 {
	return -104
}

type RequestKeys struct {
	Keys []uint64
}

func (RequestKeys) TypeCode() int8 {
	return -105
}

type ProvideStrata struct {
	DestinationHolder

	Strata *ibf.DifferenceStrata
}

func (ProvideStrata) TypeCode() int8 {
	return -106
}

type ProvideBloomFilter struct {
	DestinationHolder

	Filter *ibf.InvertibleBloomFilter
}

func (ProvideBloomFilter) TypeCode() int8 {
	return -107
}