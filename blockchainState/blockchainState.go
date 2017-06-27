// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"bytes"
	"encoding/gob"
	//"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

const COMMITEXPIRATIONM1 uint32 = 500
const COMMITEXPIRATIONM2 uint32 = 20 //TODO: set properly

const M2SWITCHHEIGHT uint32 = 70411 //TODO: double-check

type BlockchainState struct {
	NetworkID uint32

	DBlockHead      *HashPair
	DBlockHeight    uint32
	DBlockTimestamp *primitives.Timestamp
	DBlockHeader    []byte //For DBSignatureEntry

	ECBlockHead *HashPair

	FBlockHead *HashPair

	ABlockHeadRefHash *primitives.Hash

	EBlockHeads map[string]*HashPair

	ECBalances   map[string]int64
	FBalances    map[string]int64
	ExchangeRate uint64

	PendingCommits map[string]*PendingCommit //entry hash: current DBlock height

	IdentityManager identity.IdentityManager
}

func NewBSMainNet() *BlockchainState {
	bs := new(BlockchainState)
	bs.NetworkID = constants.MAIN_NETWORK_ID
	bs.IdentityManager.SetSkeletonKeyMainNet()
	bs.Init()
	return bs
}

func NewBSTestNet() *BlockchainState {
	bs := new(BlockchainState)
	bs.NetworkID = constants.TEST_NETWORK_ID
	bs.Init()
	return bs
}

func NewBSLocalNet() *BlockchainState {
	bs := new(BlockchainState)
	bs.NetworkID = constants.LOCAL_NETWORK_ID
	bs.Init()
	return bs
}

func (bs *BlockchainState) IsMainNet() bool {
	return bs.NetworkID == constants.MAIN_NETWORK_ID
}

func (bs *BlockchainState) Init() {
	if bs.EBlockHeads == nil {
		bs.EBlockHeads = map[string]*HashPair{}
	}
	if bs.ECBalances == nil {
		bs.ECBalances = map[string]int64{}
	}
	if bs.FBalances == nil {
		bs.FBalances = map[string]int64{}
	}
	if bs.PendingCommits == nil {
		bs.PendingCommits = map[string]*PendingCommit{}
	}

	if bs.DBlockHead == nil {
		bs.DBlockHead = NewHashPair()
	}
	if bs.ECBlockHead == nil {
		bs.ECBlockHead = NewHashPair()
	}
	if bs.FBlockHead == nil {
		bs.FBlockHead = NewHashPair()
	}
	if bs.ABlockHeadRefHash == nil {
		bs.ABlockHeadRefHash = primitives.NewZeroHash().(*primitives.Hash)
	}
}

func (bs *BlockchainState) ProcessBlockSet(dBlock interfaces.IDirectoryBlock, aBlock interfaces.IAdminBlock, fBlock interfaces.IFBlock, ecBlock interfaces.IEntryCreditBlock,
	eBlocks []interfaces.IEntryBlock, entries []interfaces.IEBEntry) error {
	bs.Init()
	err := bs.HandlePreBlockErrors(dBlock.DatabasePrimaryIndex())
	if err != nil {
		return err
	}

	prevHeader := bs.DBlockHeader

	err = bs.ProcessDBlock(dBlock)
	if err != nil {
		return err
	}
	err = bs.ProcessABlock(aBlock, dBlock, prevHeader)
	if err != nil {
		return err
	}
	err = bs.ProcessFBlock(fBlock)
	if err != nil {
		return err
	}
	err = bs.ProcessECBlock(ecBlock)
	if err != nil {
		return err
	}
	err = bs.ProcessEBlocks(eBlocks, entries)
	if err != nil {
		return err
	}

	err = bs.HandlePostBlockErrors(dBlock.DatabasePrimaryIndex())
	if err != nil {
		return err
	}

	return nil
}

func (bs *BlockchainState) Clone() (*BlockchainState, error) {
	data, err := bs.MarshalBinaryData()
	if err != nil {
		return nil, err
	}
	b := new(BlockchainState)
	b.Init()
	err = b.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (bs *BlockchainState) MarshalBinaryData() ([]byte, error) {
	b := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(b)
	err := enc.Encode(bs)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (bs *BlockchainState) UnmarshalBinaryData(data []byte) error {
	bs.Init()
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	return dec.Decode(bs)
}

func (e *BlockchainState) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *BlockchainState) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *BlockchainState) String() string {
	str, _ := e.JSONString()
	return str
}