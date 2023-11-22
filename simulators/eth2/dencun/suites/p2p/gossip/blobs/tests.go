package suite_blobs_gossip

import (
	"github.com/ethereum/hive/hivesim"
	"github.com/ethereum/hive/simulators/eth2/common/clients"
	"github.com/ethereum/hive/simulators/eth2/dencun/suites"
	suite_base "github.com/ethereum/hive/simulators/eth2/dencun/suites/base"
	blobber_slot_actions "github.com/marioevz/blobber/slot_actions"
)

var testSuite = hivesim.Suite{
	Name:        "eth2-deneb-p2p-blobs-gossip",
	DisplayName: "Deneb P2P Blobs Gossip",
	Description: `Collection of test vectors that verify client behavior under different blob gossiping scenarios.`,
	Location:    "suites/p2p/gossip/blobs",
}

var Tests = make([]suites.TestSpec, 0)

func init() {
	Tests = append(Tests,
		P2PBlobsGossipTestSpec{
			BaseTestSpec: suite_base.BaseTestSpec{
				Name: "test-blob-gossiping-sanity",
				Description: `
		Sanity test where the blobber is verified to be working correctly
		`,
				DenebGenesis: true,
				GenesisExecutionWithdrawalCredentialsShares: 1,
			},
		},
		P2PBlobsGossipTestSpec{
			BlobberSlotAction: blobber_slot_actions.BroadcastBlobsBeforeBlock{},
			BaseTestSpec: suite_base.BaseTestSpec{
				Name: "test-blob-gossiping-before-block",
				Description: `
		Test chain health where the blobs are gossiped before the block
		`,
				DenebGenesis: true,
				GenesisExecutionWithdrawalCredentialsShares: 1,
			},
		},
		P2PBlobsGossipTestSpec{
			BlobberSlotAction: blobber_slot_actions.BlobGossipDelay{
				DelayMilliseconds: 500,
			},
			BaseTestSpec: suite_base.BaseTestSpec{
				Name: "test-blob-gossiping-delay",
				Description: `
		Test chain health where the blobs are gossiped after the block with a 500ms delay
		`,
				DenebGenesis: true,
				GenesisExecutionWithdrawalCredentialsShares: 1,
			},
		},
		P2PBlobsGossipTestSpec{
			BlobberSlotAction: blobber_slot_actions.BlobGossipDelay{
				DelayMilliseconds: 6000,
			},
			BaseTestSpec: suite_base.BaseTestSpec{
				Name: "test-blob-gossiping-one-slot-delay",
				Description: `
		Test chain health where the blobs are gossiped after the block with a 6s delay
		`,
				DenebGenesis: true,
				GenesisExecutionWithdrawalCredentialsShares: 1,
			},
			// A slot might be missed due to blobs arriving late
			BlobberActionCausesMissedSlot: true,
		},
		P2PBlobsGossipTestSpec{
			BaseTestSpec: suite_base.BaseTestSpec{
				Name: "test-blob-gossiping-extra-blob",
				Description: `
		Test chain health where there is always an extra blob with:
		 - Correct KZG commitment
		 - Correct block root
		 - Correct proposer signature
		 - Broadcasted after the block
		 - Broadcasted before the rest of the blobs (results in correct blob being ignored per spec)
		`,
				DenebGenesis: true,
				GenesisExecutionWithdrawalCredentialsShares: 1,
			},
			BlobberSlotAction: blobber_slot_actions.ExtraBlobs{
				BroadcastBlockFirst:     true,
				BroadcastExtraBlobFirst: true,
			},
			// Since the extra blob has a correct signature, and comes before the correct blob, the correct blob is ignored
			BlobberActionCausesMissedSlot: true,
		},
		P2PBlobsGossipTestSpec{
			BaseTestSpec: suite_base.BaseTestSpec{
				Name: "test-blob-gossiping-extra-blob-with-incorrect-kzg-commitment",
				Description: `
		Test chain health where there is always an extra blob with:
		 - Incorrect KZG commitment
		 - Correct block root
		 - Correct proposer signature
		 - Broadcasted after the block
		 - Broadcasted before the rest of the blobs (results in correct blob being ignored per spec)
		`,
				DenebGenesis: true,
				GenesisExecutionWithdrawalCredentialsShares: 1,
			},
			BlobberSlotAction: blobber_slot_actions.ExtraBlobs{
				BroadcastBlockFirst:     true,
				BroadcastExtraBlobFirst: true,
				IncorrectKZGCommitment:  true,
			},
			// Since the extra blob has a correct signature, and comes before the correct blob, the correct blob is ignored
			BlobberActionCausesMissedSlot: true,
		},
		P2PBlobsGossipTestSpec{
			BaseTestSpec: suite_base.BaseTestSpec{
				Name: "test-blob-gossiping-extra-blob-with-incorrect-signature",
				Description: `
		Test chain health where there is always an extra blob with:
		 - Correct KZG commitment
		 - Correct block root
		 - Incorrect proposer signature
		 - Broadcasted after the block
		 - Broadcasted before the rest of the blobs (results in correct blob being ignored per spec)
		`,
				DenebGenesis: true,
				GenesisExecutionWithdrawalCredentialsShares: 1,
			},
			BlobberSlotAction: blobber_slot_actions.ExtraBlobs{
				BroadcastBlockFirst:     true,
				BroadcastExtraBlobFirst: true,
				IncorrectSignature:      true,
			},
			// TODO: The extra blob has an incorrect signature, so we might get disconnected+banned and unable to send the rest of the blobs
			BlobberActionCausesMissedSlot: false,
		},
		P2PBlobsGossipTestSpec{
			BaseTestSpec: suite_base.BaseTestSpec{
				Name: "test-blob-gossiping-conflicting-blobs",
				Description: `
		Test chain health where there is a single conflicting blob (same blob index) broadcasted to different clients at the same time,
		all with correct signatures and pointing to the correct block root.
		`,
				DenebGenesis: true,
				GenesisExecutionWithdrawalCredentialsShares: 1,
			},
			BlobberSlotAction: blobber_slot_actions.ConflictingBlobs{},
			// The blobs do not break any rejection rules
			BlobberActionCausesMissedSlot: false,
		},
		P2PBlobsGossipTestSpec{
			BaseTestSpec: suite_base.BaseTestSpec{
				Name: "test-blob-gossiping-max-conflicting-blobs",
				Description: `
		Test chain health where there are conflicting blobs (same blob index) broadcasted to different clients at the same time,
		all with correct signatures and pointing to the correct block root.
		`,
				DenebGenesis: true,
				GenesisExecutionWithdrawalCredentialsShares: 1,
			},
			BlobberSlotAction: blobber_slot_actions.ConflictingBlobs{
				ConflictingBlobsCount: 6,
			},
			// The blobs do not break any rejection rules
			BlobberActionCausesMissedSlot: false,
		},
		P2PBlobsGossipTestSpec{
			BaseTestSpec: suite_base.BaseTestSpec{
				Name: "test-gossiping-incorrectly-indexed-blobs",
				Description: `
		Test chain health where there are blobs with incorrect indexes broadcasted to the network,
		all with correct signatures and pointing to the correct block root.
		`,
				DenebGenesis: true,
				GenesisExecutionWithdrawalCredentialsShares: 1,
			},
			BlobberSlotAction: blobber_slot_actions.SwapBlobs{
				SplitNetwork: false,
			},
			// The blobs must be rejected as they are incorrectly indexed
			BlobberActionCausesMissedSlot: true,
		},
	)
}

func Suite(c *clients.ClientDefinitionsByRole) hivesim.Suite {
	suites.SuiteHydrate(&testSuite, c, Tests)
	return testSuite
}
