package e2etests

import (
	"github.com/stretchr/testify/require"

	testcontract "github.com/zeta-chain/node/e2e/contracts/reverter"
	"github.com/zeta-chain/node/e2e/runner"
	"github.com/zeta-chain/node/e2e/utils"
	cctypes "github.com/zeta-chain/node/x/crosschain/types"
)

func TestTONDepositAndCallRefund(r *runner.E2ERunner, args []string) {
	require.Len(r, args, 1)

	// Given amount and arbitrary call data
	var (
		amount = utils.ParseUint(r, args[0])
		data   = []byte("hello reverter")
	)

	// Given deployer mock revert contract
	// deploy a reverter contract in ZEVM
	reverterAddr, _, _, err := testcontract.DeployReverter(r.ZEVMAuth, r.ZEVMClient)
	require.NoError(r, err)
	r.Logger.Info("Reverter contract deployed at: %s", reverterAddr.String())

	// ACT
	// Send a deposit and call transaction from the deployer (faucet)
	// to the reverter contract
	cctx, err := r.TONDepositAndCall(
		&r.TONDeployer.Wallet,
		amount,
		reverterAddr,
		data,
		runner.TONExpectStatus(cctypes.CctxStatus_Reverted),
	)

	// ASSERT
	require.NoError(r, err)
	r.Logger.CCTX(*cctx, "ton_deposit_and_refund")

	require.Contains(r, cctx.CctxStatus.ErrorMessage, utils.ErrHashRevertFoo)
}
