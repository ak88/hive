package suite_blobs_gossip

import (
	"context"

	"github.com/ethereum/hive/hivesim"
	"github.com/ethereum/hive/simulators/eth2/common/clients"
	tn "github.com/ethereum/hive/simulators/eth2/common/testnet"
	"github.com/ethereum/hive/simulators/eth2/dencun/helper"
	suite_base "github.com/ethereum/hive/simulators/eth2/dencun/suites/base"
	blobber_config "github.com/marioevz/blobber/config"
	blobber_slot_actions "github.com/marioevz/blobber/slot_actions"
	beacon "github.com/protolambda/zrnt/eth2/beacon/common"
)

type P2PBlobsGossipTestSpec struct {
	suite_base.BaseTestSpec

	BlobberSlotAction             blobber_slot_actions.SlotAction
	BlobberActionCausesMissedSlot bool

	MaxMissedSlots beacon.Slot
}

const (
	WAIT_EPOCHS_AFTER_FORK = 1
	MAX_MISSED_SLOTS       = 3
)

func (ts P2PBlobsGossipTestSpec) GetDescription() *helper.Description {
	desc := ts.BaseTestSpec.GetDescription()

	// Print the base test spec description plus the blobber action description
	desc.Add("Blobber Behavior", ts.BlobberSlotAction.Description())
	return desc
}

func (ts P2PBlobsGossipTestSpec) GetMaxMissedSlots() beacon.Slot {
	if ts.MaxMissedSlots > 0 {
		return ts.MaxMissedSlots
	}
	return MAX_MISSED_SLOTS
}

func (ts P2PBlobsGossipTestSpec) GetTestnetConfig(
	allNodeDefinitions clients.NodeDefinitions,
) *tn.Config {
	config := ts.BaseTestSpec.GetTestnetConfig(allNodeDefinitions)

	config.EnableBlobber = true
	blobberActionFrequency := uint64(1)
	if ts.BlobberActionCausesMissedSlot {
		// Since we are missing slots due to the blobber action, we need to execute it every 2 slots to guarantee the chain doesn't stall
		blobberActionFrequency = 2
	}
	config.BlobberOptions = []blobber_config.Option{
		blobber_config.WithSlotAction(ts.BlobberSlotAction),
		blobber_config.WithSlotActionFrequency(blobberActionFrequency),
		blobber_config.WithAlwaysErrorValidatorResponse(),
		blobber_config.WithMaxDevP2PSessionReuses(0), // Always reuse the same peer id
	}

	config.DisablePeerScoring = true

	return config
}

func (ts P2PBlobsGossipTestSpec) ExecutePostForkWait(t *hivesim.T,
	ctx context.Context,
	testnet *tn.Testnet,
	env *tn.Environment,
	config *tn.Config,
) {
	// By default all blobber tests simply wait an epoch with a max amount of missed slots to check that the chain doesn't stall
	epochsAfterFork := WAIT_EPOCHS_AFTER_FORK
	maxMissedSlots := ts.GetMaxMissedSlots()

	if err := testnet.WaitSlotsWithMaxMissedSlots(ctx, beacon.Slot(epochsAfterFork)*testnet.Spec().SLOTS_PER_EPOCH, maxMissedSlots); err != nil {
		t.Fatalf("FAIL: error waiting for %d epochs after fork: %v", beacon.Slot(epochsAfterFork), err)
	}
}
