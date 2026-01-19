package sc

import (
	"fmt"
	"io"
	"os"

	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/snapshots/types"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	rootmulti2 "github.com/cosmos/cosmos-sdk/storev2/rootmulti"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sei-protocol/sei-chain/tools/utils"
	"github.com/sei-protocol/sei-db/config"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

type Migrator struct {
	homeDir string
	logger  log.Logger
	storeV1 store.CommitMultiStore
	storeV2 store.CommitMultiStore
}

func NewMigrator(homeDir string, db dbm.DB) *Migrator {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	// Creating CMS for store V1
	cmsV1 := rootmulti.NewStore(db, logger)
	for _, key := range utils.ModuleKeys {
		cmsV1.MountStoreWithDB(key, sdk.StoreTypeIAVL, nil)
	}
	err := cmsV1.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	// Creating CMS for store V2
	scConfig := config.DefaultStateCommitConfig()
	scConfig.Enable = true
	ssConfig := config.DefaultStateStoreConfig()
	ssConfig.Enable = true
	ssConfig.KeepRecent = 0
	cmsV2 := rootmulti2.NewStore(homeDir, logger, scConfig, ssConfig, true)
	for _, key := range utils.ModuleKeys {
		cmsV2.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	}
	err = cmsV2.LoadLatestVersion()
	if err != nil {
		panic(err)
	}
	return &Migrator{
		homeDir: homeDir,
		logger:  logger,
		storeV1: cmsV1,
		storeV2: cmsV2,
	}
}

func (m *Migrator) Migrate(version int64) error {
	// Create a snapshot
	chunks := make(chan io.ReadCloser)
	go func() {
		err := m.createSnapshot(uint64(version), chunks)
		if err != nil {
			panic(err)
		}
	}()
	streamReader, err := snapshots.NewStreamReader(chunks)
	if err != nil {
		return err
	}
	fmt.Printf("Start restoring SC store for height: %d\n", version)
	next, _ := m.storeV2.Restore(uint64(version), types.CurrentFormat, streamReader)
	for {
		if next.Item == nil {
			// end of stream
			break
		}
		// Skip any extension items
		break
	}
	fmt.Printf("Finished restoring SC store for height: %d\n", version)
	return nil
}

func (m *Migrator) createSnapshot(height uint64, chunks chan<- io.ReadCloser) error {
	streamWriter := snapshots.NewStreamWriter(chunks)
	defer streamWriter.Close()
	fmt.Printf("Start creating snapshot for height: %d\n", height)
	if err := m.storeV1.Snapshot(height, streamWriter); err != nil {
		m.logger.Error("Snapshot creation failed", "err", err)
		streamWriter.CloseWithError(err)
		return err
	}
	return nil
}
