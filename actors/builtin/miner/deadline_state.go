package miner

import (
	"fmt"
	"io"

	"github.com/filecoin-project/go-bitfield"
	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/util/adt"
)

// A bitfield of sector numbers due at each deadline.
// The sectors for each deadline are logically grouped into sequential partitions for proving.
type Deadlines struct {
	// TODO: Consider inlining part of the deadline struct (e.g.,
	// active/assigned sectors) to make sector assignment cheaper. At the
	// moment, assigning a sector requires loading all deadlines to figure
	// out where best to assign new sectors.
	Due [WPoStPeriodDeadlines]cid.Cid // []Deadline
}

type Deadline struct {
	// Partitions in this deadline, in order.
	// The keys of this AMT are always sequential integers beginning with zero.
	Partitions cid.Cid // AMT[PartitionNumber]Partition

	// Partitions that will be scheduled at the start of the next proving period.
	// Pending partitions will be _merged_ with existing partitions.
	// The keys of this AMT are always sequential integers beginning with zero, but the partitions will be
	// re-numbered when activated.
	PendingPartitions cid.Cid // AMT[PartitionNumber]Partition

	// Partitions numbers with PoSt submissions since the proving period started.
	PostSubmissions *abi.BitField

	// Number active sectors in the deadline. This number does not include
	// terminated or pending sectors.
	ActiveSectors uint64

	// Maps epochs to partitions with sectors that became faulty during that epoch.
	FaultsEpochs cid.Cid // AMT[ChainEpoch]BitField

	// Maps epochs to partitions with sectors that expire in that epoch.
	ExpirationsEpochs cid.Cid // AMT[ChainEpoch]BitField
}

//
// Deadlines (plural)
//

func ConstructDeadlines(emptyDeadlineCid cid.Cid) *Deadlines {
	d := new(Deadlines)
	for i := range d.Due {
		d.Due[i] = emptyDeadlineCid
	}
	return d
}

func (d *Deadlines) LoadDeadline(store adt.Store, dlIdx uint64) (*Deadline, error) {
	if dlIdx >= uint64(len(d.Due)) {
		return nil, xerrors.Errorf("invalid deadline %d", dlIdx)
	}
	deadline := new(Deadline)
	err := store.Get(store.Context(), d.Due[dlIdx], deadline)
	if err != nil {
		return nil, err
	}
	return deadline, nil
}

func (d *Deadlines) ForEach(store adt.Store, cb func(dlIdx uint64, dl *Deadline) error) error {
	for dlIdx := range d.Due {
		dl, err := d.LoadDeadline(store, uint64(dlIdx))
		if err != nil {
			return err
		}
		err = cb(uint64(dlIdx), dl)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Deadlines) UpdateDeadline(store adt.Store, dlIdx uint64, deadline *Deadline) error {
	if dlIdx >= uint64(len(d.Due)) {
		return xerrors.Errorf("invalid deadline %d", dlIdx)
	}
	dlCid, err := store.Put(store.Context(), deadline)
	if err != nil {
		return err
	}
	d.Due[dlIdx] = dlCid
	return nil
}

//
// Deadline (singular)
//

func ConstructDeadline(emptyArrayCid cid.Cid) *Deadline {
	return &Deadline{
		Partitions:        emptyArrayCid,
		PendingPartitions: emptyArrayCid,
		PostSubmissions:   abi.NewBitField(),
		FaultsEpochs:      emptyArrayCid,
		ExpirationsEpochs: emptyArrayCid,
	}
}

func (d *Deadline) PartitionsArray(store adt.Store) (*adt.Array, error) {
	return adt.AsArray(store, d.Partitions)
}

func (dl *Deadline) PopExpiredPartitions(store adt.Store, until abi.ChainEpoch) (*bitfield.BitField, error) {
	stopErr := fmt.Errorf("stop")

	partitionExpirationQ, err := adt.AsArray(store, dl.ExpirationsEpochs)
	if err != nil {
		return nil, err
	}

	partitionsWithExpiredSectors := abi.NewBitField()
	var expiredEpochs []uint64
	var bf bitfield.BitField
	err = partitionExpirationQ.ForEach(&bf, func(i int64) error {
		if abi.ChainEpoch(i) > until {
			return stopErr
		}
		expiredEpochs = append(expiredEpochs, uint64(i))
		partitionsWithExpiredSectors, err = bitfield.MergeBitFields(partitionsWithExpiredSectors, &bf)
		if err != nil {
			return err
		}
		return nil
	})
	switch err {
	case nil, stopErr:
	default:
		return nil, err
	}

	err = partitionExpirationQ.BatchDelete(expiredEpochs)
	if err == nil {
		return nil, err
	}

	dl.ExpirationsEpochs, err = partitionExpirationQ.Root()
	if err != nil {
		return nil, err
	}

	return partitionsWithExpiredSectors, nil
}

func (dl *Deadline) MarshalCBOR(w io.Writer) error {
	panic("implement me")
}

func (dl *Deadline) UnmarshalCBOR(r io.Reader) error {
	panic("implement me")
}
