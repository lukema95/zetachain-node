package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/zeta-chain/node/testutil/sample"
	zetasimulation "github.com/zeta-chain/node/testutil/simulation"
	"github.com/zeta-chain/node/x/observer/keeper"
	"github.com/zeta-chain/node/x/observer/types"
)

// SimulateUpdateKeygen generates a MsgUpdateKeygen and delivers it.
func SimulateUpdateKeygen(k keeper.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accounts []simtypes.Account, _ string,
	) (OperationMsg simtypes.OperationMsg, futureOps []simtypes.FutureOperation, err error) {
		policyAccount, err := zetasimulation.GetPolicyAccount(ctx, k.GetAuthorityKeeper(), accounts)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgUpdateKeygen, err.Error()), nil, nil
		}

		authAccount := k.GetAuthKeeper().GetAccount(ctx, policyAccount.Address)
		spendable := k.GetBankKeeper().SpendableCoins(ctx, authAccount.GetAddress())

		_, found := k.GetKeygen(ctx)
		if !found {
			kg := sample.KeygenFromRand(r)
			k.SetKeygen(ctx, kg)
		}

		blockHeightMin := ctx.BlockHeight() + 11
		blockHeightMax := ctx.BlockHeight() + 1000
		keygenBlockHeight := int64(r.Intn(int(blockHeightMax-blockHeightMin))) + blockHeightMin

		msg := types.MsgUpdateKeygen{
			Creator: policyAccount.Address.String(),
			Block:   keygenBlockHeight,
		}

		err = msg.ValidateBasic()
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgUpdateKeygen, err.Error()), nil, err
		}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           moduletestutil.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             &msg,
			Context:         ctx,
			SimAccount:      policyAccount,
			AccountKeeper:   k.GetAuthKeeper(),
			Bankkeeper:      k.GetBankKeeper(),
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return zetasimulation.GenAndDeliverTxWithRandFees(txCtx, true)
	}
}
