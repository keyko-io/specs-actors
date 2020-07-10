package verifreg_test

import (
	"context"
	"fmt"
	"testing"

	addr "github.com/filecoin-project/go-address"
	big "github.com/filecoin-project/specs-actors/actors/abi/big"
	builtin "github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/builtin/market"
	"github.com/filecoin-project/specs-actors/actors/builtin/verifreg"
	exitcode "github.com/filecoin-project/specs-actors/actors/runtime/exitcode"
	adt "github.com/filecoin-project/specs-actors/actors/util/adt"
	"github.com/filecoin-project/specs-actors/support/mock"
	tutil "github.com/filecoin-project/specs-actors/support/testing"
	"github.com/stretchr/testify/assert"
)

func TestExports(t *testing.T) {
	mock.CheckActorExports(t, verifreg.Actor{})
}

func TestRemoveAllError(t *testing.T) {
	verifiedRegistryActor := tutil.NewIDAddr(t, 100)
	builder := mock.NewBuilder(context.Background(), verifiedRegistryActor)
	rt := builder.Build(t)
	store := adt.AsStore(rt)

	smm := market.MakeEmptySetMultimap(store)

	if err := smm.RemoveAll(42); err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
}

func TestVerifiedRegistryActor(t *testing.T) {

	// var st verifreg.State
	actor := verifreg.Actor{}
	receiver := tutil.NewIDAddr(t, 100)
	builder := mock.NewBuilder(context.Background(), receiver).
		WithCaller(builtin.SystemActorAddr, builtin.SystemActorCodeID)

	holder := tutil.NewIDAddr(t, 101)
	holderAddress := addr.Address(holder)

	verifier := tutil.NewIDAddr(t, 102)
	verifierAddress := addr.Address(verifier)

	client := tutil.NewIDAddr(t, 103)
	clientAddress := addr.Address(client)

	t.Run("simple construction", func(t *testing.T) {

		rt := builder.Build(t)

		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		store := adt.AsStore(rt)

		emptyMap, err := adt.MakeEmptyMap(store).Root()
		assert.NoError(t, err)

		var state verifreg.State
		rt.GetState(&state)

		assert.Equal(t, emptyMap, state.Verifiers)
		assert.Equal(t, emptyMap, state.VerifiedClients)
	})

	t.Run("Add Verifier", func(t *testing.T) {

		rt := builder.Build(t)
		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		store := adt.AsStore(rt)
		var state verifreg.State

		params := verifreg.AddVerifierParams{
			Address:   verifierAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 30)),
		}

		rt.SetCaller(holderAddress, builtin.MultisigActorCodeID)
		rt.ExpectValidateCallerAddr(holderAddress)

		ret = rt.Call(actor.AddVerifier, &params).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)

		datacap, found, err := state.GetVerifier(store, verifierAddress)
		if err != nil {
			fmt.Print("fail getting verifiers registry state")
		}
		assert.Equal(t, true, found)
		assert.Equal(t, big.NewInt(1<<30), *datacap)

	})

	t.Run("Remove Verifier", func(t *testing.T) {

		rt := builder.Build(t)
		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		store := adt.AsStore(rt)
		var state verifreg.State

		params := verifreg.AddVerifierParams{
			Address:   verifierAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 30)),
		}

		rt.SetCaller(holderAddress, builtin.MultisigActorCodeID)
		rt.ExpectValidateCallerAddr(holderAddress)

		ret = rt.Call(actor.AddVerifier, &params).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)

		_, found, err := state.GetVerifier(store, verifierAddress)
		if err != nil {
			fmt.Print("fail getting verifiers registry state")
		}
		assert.Equal(t, true, found)

		rt.SetCaller(holderAddress, builtin.MultisigActorCodeID)
		rt.ExpectValidateCallerAddr(holderAddress)
		ret = rt.Call(actor.RemoveVerifier, &verifierAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)

		_, found, err = state.GetVerifier(store, verifierAddress)
		if err != nil {
			fmt.Print("fail getting verifiers registry state")
		}
		assert.Equal(t, false, found)

	})

	t.Run("Failing Removing Verifier", func(t *testing.T) {

		rt := builder.Build(t)
		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.SetCaller(holderAddress, builtin.MultisigActorCodeID)
		rt.ExpectValidateCallerAddr(holderAddress)

		rt.ExpectAbort(exitcode.ErrIllegalState, func() {
			rt.Call(actor.RemoveVerifier, &verifierAddress)
		})
		rt.Verify()

	})

	/*
		t.Run("Failing Add Verifier Client with datacap allowance too low", func(t *testing.T) {

			rt := builder.Build(t)
			rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

			ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
			assert.Nil(t, ret)
			rt.Verify()

			store := adt.AsStore(rt)
			var state verifreg.State
			params := verifreg.AddVerifierParams{
				Address:   verifierAddress,
				Allowance: verifreg.DataCap(big.NewInt(1 << 30)),
			}

			rt.SetCaller(holderAddress, builtin.MultisigActorCodeID)
			rt.ExpectValidateCallerAddr(holderAddress)

			ret = rt.Call(actor.AddVerifier, &params).(*adt.EmptyValue)
			assert.Nil(t, ret)
			rt.Verify()

			rt.GetState(&state)
			_, found, err := state.GetVerifier(store, verifierAddress)
			if err != nil {
				fmt.Print("fail getting verifiers registry state")
			}
			assert.Equal(t, true, found)

			client := tutil.NewIDAddr(t, 103)
			clientAddress := addr.Address(client)
			addClientParams := verifreg.AddVerifiedClientParams{
				Address:   clientAddress,
				Allowance: verifreg.DataCap(big.NewInt(100)),
			}

			rt.SetCaller(verifierAddress, builtin.AccountActorCodeID)
			rt.ExpectValidateCallerAddr(verifierAddress)

			rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
				rt.Call(actor.AddVerifiedClient, &addClientParams)
			})
			rt.Verify()

		})
	*/

	t.Run("Add Verifier Client", func(t *testing.T) {

		rt := builder.Build(t)
		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		store := adt.AsStore(rt)
		var state verifreg.State
		params := verifreg.AddVerifierParams{
			Address:   verifierAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 30)),
		}

		rt.SetCaller(holderAddress, builtin.MultisigActorCodeID)
		rt.ExpectValidateCallerAddr(holderAddress)

		ret = rt.Call(actor.AddVerifier, &params).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)

		_, found, err := state.GetVerifier(store, verifierAddress)
		if err != nil {
			fmt.Print("fail getting verifiers registry state")
		}
		assert.Equal(t, true, found)

		addClientParams := verifreg.AddVerifiedClientParams{
			Address:   clientAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 21)),
		}

		rt.SetCaller(verifierAddress, builtin.AccountActorCodeID)
		rt.ExpectValidateCallerAny()

		ret = rt.Call(actor.AddVerifiedClient, &addClientParams).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)
		_, found, err = state.GetVerifiedClient(store, clientAddress)
		if err != nil {
			fmt.Print("fail getting verified clients registry state")
		}
		assert.Equal(t, true, found)

	})

	t.Run("Failing Add Verifier Client, Verifier not found", func(t *testing.T) {

		rt := builder.Build(t)
		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		addClientParams := verifreg.AddVerifiedClientParams{
			Address:   clientAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 21)),
		}

		rt.SetCaller(verifierAddress, builtin.AccountActorCodeID)
		rt.ExpectValidateCallerAny()

		rt.ExpectAbort(exitcode.ErrNotFound, func() {
			rt.Call(actor.AddVerifiedClient, &addClientParams)
		})

		rt.Verify()

	})

	t.Run("Failing Add Verifier Client, Datacap Allowance Error", func(t *testing.T) {

		rt := builder.Build(t)
		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		store := adt.AsStore(rt)
		var state verifreg.State
		params := verifreg.AddVerifierParams{
			Address:   verifierAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 30)),
		}

		rt.SetCaller(holderAddress, builtin.MultisigActorCodeID)
		rt.ExpectValidateCallerAddr(holderAddress)

		ret = rt.Call(actor.AddVerifier, &params).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)
		_, found, err := state.GetVerifier(store, verifierAddress)
		if err != nil {
			fmt.Print("fail getting verifiers registry state")
		}
		assert.Equal(t, true, found)

		addClientParams := verifreg.AddVerifiedClientParams{
			Address:   clientAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 31)),
		}

		rt.SetCaller(verifierAddress, builtin.AccountActorCodeID)
		rt.ExpectValidateCallerAny()

		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			rt.Call(actor.AddVerifiedClient, &addClientParams)
		})

		rt.Verify()

	})

	t.Run("Failing Add Verifier Client, Client already exist", func(t *testing.T) {

		rt := builder.Build(t)
		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		store := adt.AsStore(rt)
		var state verifreg.State
		params := verifreg.AddVerifierParams{
			Address:   verifierAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 30)),
		}

		rt.SetCaller(holderAddress, builtin.MultisigActorCodeID)
		rt.ExpectValidateCallerAddr(holderAddress)

		ret = rt.Call(actor.AddVerifier, &params).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)
		_, found, err := state.GetVerifier(store, verifierAddress)
		if err != nil {
			fmt.Print("fail getting verifiers registry state")
		}
		assert.Equal(t, true, found)

		addClientParams := verifreg.AddVerifiedClientParams{
			Address:   clientAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 21)),
		}

		rt.SetCaller(verifierAddress, builtin.AccountActorCodeID)
		rt.ExpectValidateCallerAny()

		ret = rt.Call(actor.AddVerifiedClient, &addClientParams).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.SetCaller(verifierAddress, builtin.AccountActorCodeID)
		rt.ExpectValidateCallerAny()

		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			rt.Call(actor.AddVerifiedClient, &addClientParams)
		})

		rt.Verify()

	})

	t.Run("Failing Use Bytes, low deal", func(t *testing.T) {

		rt := builder.Build(t)
		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		useBytesParams := verifreg.UseBytesParams{
			Address:  clientAddress,
			DealSize: big.NewInt(1 << 2),
		}

		rt.SetCaller(builtin.StorageMarketActorAddr, builtin.StorageMarketActorCodeID)
		rt.ExpectValidateCallerAddr(builtin.StorageMarketActorAddr)

		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			rt.Call(actor.UseBytes, &useBytesParams)
		})

		rt.Verify()

	})

	t.Run("Failing Use Bytes, Verified Client not found", func(t *testing.T) {

		rt := builder.Build(t)
		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		useBytesParams := verifreg.UseBytesParams{
			Address:  clientAddress,
			DealSize: big.NewInt(1 << 21),
		}

		rt.SetCaller(builtin.StorageMarketActorAddr, builtin.StorageMarketActorCodeID)
		rt.ExpectValidateCallerAddr(builtin.StorageMarketActorAddr)

		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			rt.Call(actor.UseBytes, &useBytesParams)
		})

		rt.Verify()

	})

	t.Run("Failing Use Bytes, Verified Client not enough datacap allowance", func(t *testing.T) {

		rt := builder.Build(t)
		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		store := adt.AsStore(rt)
		var state verifreg.State
		params := verifreg.AddVerifierParams{
			Address:   verifierAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 30)),
		}

		rt.SetCaller(holderAddress, builtin.MultisigActorCodeID)
		rt.ExpectValidateCallerAddr(holderAddress)

		ret = rt.Call(actor.AddVerifier, &params).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)
		_, found, err := state.GetVerifier(store, verifierAddress)
		if err != nil {
			fmt.Print("fail getting verifiers registry state")
		}
		assert.Equal(t, true, found)

		addClientParams := verifreg.AddVerifiedClientParams{
			Address:   clientAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 21)),
		}

		rt.SetCaller(verifierAddress, builtin.AccountActorCodeID)
		rt.ExpectValidateCallerAny()

		ret = rt.Call(actor.AddVerifiedClient, &addClientParams).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)
		_, found, err = state.GetVerifiedClient(store, clientAddress)
		if err != nil {
			fmt.Print("fail getting verified clients registry state")
		}
		assert.Equal(t, true, found)

		useBytesParams := verifreg.UseBytesParams{
			Address:  clientAddress,
			DealSize: big.NewInt(1 << 22),
		}

		rt.SetCaller(builtin.StorageMarketActorAddr, builtin.StorageMarketActorCodeID)
		rt.ExpectValidateCallerAddr(builtin.StorageMarketActorAddr)

		rt.ExpectAbort(exitcode.ErrIllegalArgument, func() {
			rt.Call(actor.UseBytes, &useBytesParams)
		})

		rt.Verify()

	})

	t.Run("Use Bytes OK", func(t *testing.T) {

		rt := builder.Build(t)
		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		store := adt.AsStore(rt)
		var state verifreg.State
		params := verifreg.AddVerifierParams{
			Address:   verifierAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 60)),
		}

		rt.SetCaller(holderAddress, builtin.MultisigActorCodeID)
		rt.ExpectValidateCallerAddr(holderAddress)

		ret = rt.Call(actor.AddVerifier, &params).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)
		_, found, err := state.GetVerifier(store, verifierAddress)
		if err != nil {
			fmt.Print("fail getting verifiers registry state")
		}
		assert.Equal(t, true, found)

		addClientParams := verifreg.AddVerifiedClientParams{
			Address:   clientAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 30)),
		}

		rt.SetCaller(verifierAddress, builtin.AccountActorCodeID)
		rt.ExpectValidateCallerAny()

		ret = rt.Call(actor.AddVerifiedClient, &addClientParams).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)

		_, found, err = state.GetVerifiedClient(store, clientAddress)
		if err != nil {
			fmt.Print("fail getting verified clients registry state")
		}
		assert.Equal(t, true, found)

		useBytesParams := verifreg.UseBytesParams{
			Address:  clientAddress,
			DealSize: big.NewInt(1 << 22),
		}

		rt.SetCaller(builtin.StorageMarketActorAddr, builtin.StorageMarketActorCodeID)
		rt.ExpectValidateCallerAddr(builtin.StorageMarketActorAddr)

		ret = rt.Call(actor.UseBytes, &useBytesParams).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)

		_, found, err = state.GetVerifiedClient(store, clientAddress)
		if err != nil {
			fmt.Print("fail getting verified clients registry state")
		}
		assert.Equal(t, true, found)

	})

	t.Run("Use Bytes and then delete verified Client", func(t *testing.T) {

		rt := builder.Build(t)
		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		store := adt.AsStore(rt)
		var state verifreg.State
		params := verifreg.AddVerifierParams{
			Address:   verifierAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 30)),
		}

		rt.SetCaller(holderAddress, builtin.MultisigActorCodeID)
		rt.ExpectValidateCallerAddr(holderAddress)

		ret = rt.Call(actor.AddVerifier, &params).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)
		_, found, err := state.GetVerifier(store, verifierAddress)
		if err != nil {
			fmt.Print("fail getting verifiers registry state")
		}
		assert.Equal(t, true, found)

		addClientParams := verifreg.AddVerifiedClientParams{
			Address:   clientAddress,
			Allowance: verifreg.DataCap(big.NewInt(1 << 22)),
		}

		rt.SetCaller(verifierAddress, builtin.AccountActorCodeID)
		rt.ExpectValidateCallerAny()

		ret = rt.Call(actor.AddVerifiedClient, &addClientParams).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)

		_, found, err = state.GetVerifiedClient(store, clientAddress)
		if err != nil {
			fmt.Print("fail getting verified clients registry state")
		}
		assert.Equal(t, true, found)

		useBytesParams := verifreg.UseBytesParams{
			Address:  clientAddress,
			DealSize: big.NewInt(1 << 22),
		}

		rt.SetCaller(builtin.StorageMarketActorAddr, builtin.StorageMarketActorCodeID)
		rt.ExpectValidateCallerAddr(builtin.StorageMarketActorAddr)

		ret = rt.Call(actor.UseBytes, &useBytesParams).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)

		_, found, err = state.GetVerifiedClient(store, clientAddress)
		if err != nil {
			fmt.Print("fail getting verified clients registry state")
		}
		assert.Equal(t, false, found)

	})

	t.Run("Restore Bytes OK, adding new Verified Client with datacap allowance", func(t *testing.T) {

		rt := builder.Build(t)
		rt.ExpectValidateCallerAddr(builtin.SystemActorAddr)

		ret := rt.Call(actor.Constructor, &holderAddress).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		store := adt.AsStore(rt)
		var state verifreg.State

		restoreBytesParams := verifreg.RestoreBytesParams{
			Address:  clientAddress,
			DealSize: big.NewInt(1 << 21),
		}

		rt.SetCaller(builtin.StorageMarketActorAddr, builtin.StorageMarketActorCodeID)
		rt.ExpectValidateCallerAddr(builtin.StorageMarketActorAddr)

		ret = rt.Call(actor.RestoreBytes, &restoreBytesParams).(*adt.EmptyValue)
		assert.Nil(t, ret)
		rt.Verify()

		rt.GetState(&state)

		vcCap, found, err := state.GetVerifiedClient(store, clientAddress)
		if err != nil || !found {
			fmt.Print("fail getting verified clients registry state")
		}

		assert.Equal(t, big.NewInt(1<<21), vcCap)

	})

}
