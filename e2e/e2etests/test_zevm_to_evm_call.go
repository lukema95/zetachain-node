package e2etests

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/require"
	"github.com/zeta-chain/protocol-contracts/pkg/gatewayzevm.sol"

	"github.com/zeta-chain/node/e2e/runner"
	"github.com/zeta-chain/node/e2e/utils"
	crosschaintypes "github.com/zeta-chain/node/x/crosschain/types"
)

func TestZEVMToEVMCall(r *runner.E2ERunner, args []string) {
	require.Len(r, args, 0)

	payload := randomPayload(r)

	r.AssertTestDAppEVMCalled(false, payload, big.NewInt(0))

	// necessary approval for fee payment
	r.ApproveETHZRC20(r.GatewayZEVMAddr)

	// perform the authenticated call
	tx := r.ZEVMToEMVCall(
		r.TestDAppV2EVMAddr,
		[]byte(payload),
		gatewayzevm.RevertOptions{
			OnRevertGasLimit: big.NewInt(0),
		},
	)

	// wait for the cctx to be mined
	cctx := utils.WaitCctxMinedByInboundHash(r.Ctx, tx.Hash().Hex(), r.CctxClient, r.Logger, r.CctxTimeout)
	r.Logger.CCTX(*cctx, "call")
	require.Equal(r, crosschaintypes.CctxStatus_OutboundMined, cctx.CctxStatus.Status)

	// check the payload was received on the contract
	r.AssertTestDAppEVMCalled(true, payload, big.NewInt(0))

	// check expected sender was used
	senderForMsg, err := r.TestDAppV2EVM.SenderWithMessage(&bind.CallOpts{}, []byte(payload))
	require.NoError(r, err)
	require.Equal(r, r.ZEVMAuth.From, senderForMsg)
}
