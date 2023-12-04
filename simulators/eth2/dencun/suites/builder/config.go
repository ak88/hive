package suite_builder

import (
	"fmt"
	"math/big"

	"github.com/ethereum/hive/simulators/eth2/common/clients"
	"github.com/ethereum/hive/simulators/eth2/common/testnet"
	"github.com/ethereum/hive/simulators/eth2/common/utils"
	suite_base "github.com/ethereum/hive/simulators/eth2/dencun/suites/base"
	mock_builder "github.com/marioevz/mock-builder/mock"
	beacon "github.com/protolambda/zrnt/eth2/beacon/common"
)

var REQUIRES_FINALIZATION_TO_ACTIVATE_BUILDER = []string{
	"lighthouse",
	"teku",
}

type BuilderTestSpec struct {
	suite_base.BaseTestSpec
	ErrorOnHeaderRequest        bool
	ErrorOnPayloadReveal        bool
	InvalidPayloadVersion       bool
	InvalidatePayload           mock_builder.PayloadInvalidation
	InvalidatePayloadAttributes mock_builder.PayloadAttributesInvalidation
}

func (ts BuilderTestSpec) GetTestnetConfig(
	allNodeDefinitions clients.NodeDefinitions,
) *testnet.Config {
	tc := ts.BaseTestSpec.GetTestnetConfig(allNodeDefinitions)

	tc.DenebForkEpoch = big.NewInt(1)

	if len(
		allNodeDefinitions.FilterByCL(
			REQUIRES_FINALIZATION_TO_ACTIVATE_BUILDER,
		),
	) > 0 {
		// At least one of the CLs require finalization to start requesting
		// headers from the builder
		tc.DenebForkEpoch = big.NewInt(5)
	}

	// Builders are always enabled for these tests
	tc.EnableBuilders = true

	// Builder config
	// Configure the builder according to the error
	tc.BuilderOptions = make([]mock_builder.Option, 0)

	// Bump the built payloads value
	tc.BuilderOptions = append(
		tc.BuilderOptions,
		mock_builder.WithPayloadWeiValueMultiplier(big.NewInt(10)),
		mock_builder.WithExtraDataWatermark("builder payload tst"),
	)

	// Inject test error
	denebEpoch := beacon.Epoch(tc.DenebForkEpoch.Uint64())
	if ts.ErrorOnHeaderRequest {
		tc.BuilderOptions = append(
			tc.BuilderOptions,
			mock_builder.WithErrorOnHeaderRequestAtEpoch(denebEpoch),
		)
	}
	if ts.ErrorOnPayloadReveal {
		tc.BuilderOptions = append(
			tc.BuilderOptions,
			mock_builder.WithErrorOnPayloadRevealAtEpoch(denebEpoch),
		)
	}
	if ts.InvalidatePayload != "" {
		tc.BuilderOptions = append(
			tc.BuilderOptions,
			mock_builder.WithPayloadInvalidatorAtEpoch(
				denebEpoch,
				ts.InvalidatePayload,
			),
		)
	}
	if ts.InvalidatePayloadAttributes != "" {
		tc.BuilderOptions = append(
			tc.BuilderOptions,
			mock_builder.WithPayloadAttributesInvalidatorAtEpoch(
				denebEpoch,
				ts.InvalidatePayloadAttributes,
			),
		)
	}
	if ts.InvalidPayloadVersion {
		tc.BuilderOptions = append(
			tc.BuilderOptions,
			mock_builder.WithInvalidBuilderBidVersionAtEpoch(denebEpoch),
		)
	}

	return tc
}

func (ts BuilderTestSpec) GetDescription() *utils.Description {
	desc := ts.BaseTestSpec.GetDescription()
	desc.Add(utils.CategoryTestnetConfiguration, `
	- Deneb/Cancun transition occurs on Epoch 1 or 5
		- Epoch depends on whether builder workflow activation requires finalization [on the CL client](#clients-that-require-finalization-to-enable-builder).
	- Builder is enabled for all nodes
	- Builder action is only enabled after fork
	- Nodes have the mock-builder configured as builder endpoint`)
	desc.Add(utils.CategoryVerificationsConsensusClient, `
	- Verify that the builder, up to before Deneb fork, has been able to produce blocks and they have been included in the canonical chain`)
	if ts.InvalidPayloadCaughtBeforeReveal() {
		desc.Add(utils.CategoryVerificationsConsensusClient, fmt.Sprintf(`
	- After Deneb fork, the builder must be able to include blocks with blobs in the canonical chain, which implicitly verifies:
		- Consensus client is able to properly format header requests to the builder
		- Consensus client is able to properly format blinded signed requests to the builder
		- No signed block contained an invalid format or signature
	- There are no more than %d missed slots on the latest epoch`, MAX_MISSED_SLOTS_NO_CIRCUIT_BREAKER))
	} else {
		desc.Add(utils.CategoryVerificationsConsensusClient, fmt.Sprintf(`
		- Circuit breaker correctly kicks in and disables the builder workflow
			- Implicitly verified by missed-slot count
		- Builder workflow is re-enabled after the circuit breaker timeout
		- There are no more than %d missed slots on the latest epoch`, MAX_MISSED_SLOTS_BEFORE_CIRCUIT_BREAKER))
	}
	return desc
}
