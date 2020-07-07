package reward

import (
	abi "github.com/filecoin-project/specs-actors/actors/abi"
	big "github.com/filecoin-project/specs-actors/actors/abi/big"
)

// A quantity of space * time (in byte-epochs) representing power committed to the network for some duration.
type Spacetime = big.Int

type State struct {
	// CumsumBaseline is a target CumsumRealized needs to reach for EffectiveNetworkTime to increase
	CumsumBaseline Spacetime
	CumsumRealized Spacetime

	// EffectiveNetworkTime is ceiling of EffectiveNetworkTime based on CumsumRealized
	EffectiveNetworkTime abi.ChainEpoch

	// The reward to be paid in per WinCount to block producers.
	// The actual reward total paid out depends on the number of winners in any round.
	ThisEpochReward abi.TokenAmount
	Epoch           abi.ChainEpoch
}

func ConstructState(currRealizedPower abi.StoragePower) *State {
	st := &State{
		CumsumBaseline: big.Zero(),
		CumsumRealized: big.Zero(),

		ThisEpochReward: big.Zero(),
	}

	st.Epoch = -1 // updateToNextEpoch increments it before using it
	st.updateToNextEpochWithReward(currRealizedPower)

	return st
}

// Takes in current realized power and updates internal state
// Used for update of internal state during null rounds
func (st *State) updateToNextEpoch(currRealizedPower abi.StoragePower) {
	st.Epoch++

	cappedRealizedPower := big.Min(BaselinePowerAt(st.Epoch), currRealizedPower)
	st.CumsumRealized = big.Add(st.CumsumRealized, cappedRealizedPower)

	for st.CumsumRealized.GreaterThan(st.CumsumBaseline) {
		st.EffectiveNetworkTime++
		st.CumsumBaseline = big.Add(st.CumsumBaseline, BaselinePowerAt(st.EffectiveNetworkTime))
	}
}

// Takes in a current realized power for a reward epoch and computes
// and updates reward state to track reward for the next epoch
func (st *State) updateToNextEpochWithReward(currRealizedPower abi.StoragePower) {
	prevRewardTheta := computeRTheta(st.EffectiveNetworkTime, st.CumsumRealized, st.CumsumBaseline)
	st.updateToNextEpoch(currRealizedPower)
	currRewardTheta := computeRTheta(st.EffectiveNetworkTime, st.CumsumRealized, st.CumsumBaseline)

	st.ThisEpochReward = computeReward(st.Epoch, prevRewardTheta, currRewardTheta)
}
