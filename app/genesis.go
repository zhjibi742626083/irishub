package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/irisnet/irishub/modules/gov"
	"github.com/irisnet/irishub/modules/service"
	"github.com/irisnet/irishub/modules/upgrade"
	"github.com/irisnet/irishub/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	Denom             = "iris"
	StakeDenom        = Denom + "-" + types.Atto
	FeeAmt            = int64(100)
	IrisCt            = types.NewDefaultCoinType(Denom)
	FreeFermionVal, _ = IrisCt.ConvertToMinCoin(fmt.Sprintf("%d%s", FeeAmt, Denom))
	FreeFermionAcc, _ = IrisCt.ConvertToMinCoin(fmt.Sprintf("%d%s", int64(150), Denom))
)

const (
	defaultUnbondingTime time.Duration = 60 * 10 * time.Second
	// DefaultKeyPass contains the default key password for genesis transactions
	DefaultKeyPass = "1234567890"
)

// State to Unmarshal
type GenesisState struct {
	Accounts     []GenesisAccount      `json:"accounts"`
	AuthData     auth.GenesisState     `json:"auth"`
	StakeData    stake.GenesisState    `json:"stake"`
	MintData     mint.GenesisState     `json:"mint"`
	DistrData    distr.GenesisState    `json:"distr"`
	GovData      gov.GenesisState      `json:"gov"`
	UpgradeData  upgrade.GenesisState  `json:"upgrade"`
	SlashingData slashing.GenesisState `json:"slashing"`
	ServiceData  service.GenesisState  `json:"service"`
	GenTxs       []json.RawMessage     `json:"gentxs"`
}

func NewGenesisState(accounts []GenesisAccount, authData auth.GenesisState, stakeData stake.GenesisState, mintData mint.GenesisState,
	distrData distr.GenesisState, govData gov.GenesisState, upgradeData upgrade.GenesisState, serviceData service.GenesisState, slashingData slashing.GenesisState) GenesisState {

	return GenesisState{
		Accounts:     accounts,
		AuthData:     authData,
		StakeData:    stakeData,
		MintData:     mintData,
		DistrData:    distrData,
		GovData:      govData,
		UpgradeData:  upgradeData,
		ServiceData:  serviceData,
		SlashingData: slashingData,
	}
}

// GenesisAccount doesn't need pubkey or sequence
type GenesisAccount struct {
	Address       sdk.AccAddress `json:"address"`
	Coins         sdk.Coins      `json:"coins"`
	Sequence      int64          `json:"sequence_number"`
	AccountNumber int64          `json:"account_number"`
}

func NewGenesisAccount(acc *auth.BaseAccount) GenesisAccount {
	return GenesisAccount{
		Address:       acc.Address,
		Coins:         acc.Coins,
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
	}
}

func NewGenesisAccountI(acc auth.Account) GenesisAccount {
	return GenesisAccount{
		Address:       acc.GetAddress(),
		Coins:         acc.GetCoins(),
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence(),
	}
}

// convert GenesisAccount to auth.BaseAccount
func (ga *GenesisAccount) ToAccount() (acc *auth.BaseAccount) {
	return &auth.BaseAccount{
		Address:       ga.Address,
		Coins:         ga.Coins.Sort(),
		AccountNumber: ga.AccountNumber,
		Sequence:      ga.Sequence,
	}
}

// NewDefaultGenesisState generates the default state for gaia.
func NewDefaultGenesisState() GenesisState {
	return GenesisState{
		Accounts:     nil,
		StakeData:    createStakeGenesisState(),
		MintData:     createMintGenesisState(),
		DistrData:    distr.DefaultGenesisState(),
		GovData:      gov.DefaultGenesisState(),
		UpgradeData:  upgrade.DefaultGenesisState(),
		ServiceData:  service.DefaultGenesisState(),
		SlashingData: slashing.DefaultGenesisState(),
		GenTxs:       nil,
	}
}

// get app init parameters for server init command
func IrisAppInit() server.AppInit {
	return server.AppInit{
		AppGenState: IrisAppGenStateJSON,
	}
}

// Create the core parameters for genesis initialization for iris
// note that the pubkey input is this machines pubkey
func IrisAppGenState(cdc *codec.Codec, genDoc tmtypes.GenesisDoc, appGenTxs []json.RawMessage) (
	genesisState GenesisState, err error) {
	if err = cdc.UnmarshalJSON(genDoc.AppState, &genesisState); err != nil {
		return genesisState, err
	}

	// if there are no gen txs to be processed, return the default empty state
	if len(appGenTxs) == 0 {
		return genesisState, errors.New("there must be at least one genesis tx")
	}

	stakeData := genesisState.StakeData
	for i, genTx := range appGenTxs {
		var tx auth.StdTx
		if err := cdc.UnmarshalJSON(genTx, &tx); err != nil {
			return genesisState, err
		}
		msgs := tx.GetMsgs()
		if len(msgs) != 1 {
			return genesisState, errors.New(
				"must provide genesis StdTx with exactly 1 CreateValidator message")
		}
		if _, ok := msgs[0].(stake.MsgCreateValidator); !ok {
			return genesisState, fmt.Errorf(
				"Genesis transaction %v does not contain a MsgCreateValidator", i)
		}
	}

	for _, acc := range genesisState.Accounts {
		// create the genesis account, give'm few iris-atto and a buncha token with there name
		for _, coin := range acc.Coins {
			if coin.Denom == StakeDenom {
				stakeData.Pool.LooseTokens = stakeData.Pool.LooseTokens.
					Add(sdk.NewDecFromInt(coin.Amount)) // increase the supply
			}
		}
	}
	genesisState.StakeData = stakeData
	genesisState.GenTxs = appGenTxs
	genesisState.UpgradeData = genesisState.UpgradeData
	return genesisState, nil
}

func genesisAccountFromMsgCreateValidator(msg stake.MsgCreateValidator, amount sdk.Int) GenesisAccount {
	accAuth := auth.NewBaseAccountWithAddress(sdk.AccAddress(msg.ValidatorAddr))
	accAuth.Coins = []sdk.Coin{
		{StakeDenom, amount},
	}
	return NewGenesisAccount(&accAuth)
}

// IrisValidateGenesisState ensures that the genesis state obeys the expected invariants
// TODO: No validators are both bonded and jailed (#2088)
// TODO: Error if there is a duplicate validator (#1708)
// TODO: Ensure all state machine parameters are in genesis (#1704)
func IrisValidateGenesisState(genesisState GenesisState) (err error) {
	err = validateGenesisStateAccounts(genesisState.Accounts)
	if err != nil {
		return
	}
	// skip stakeData validation as genesis is created from txs
	if len(genesisState.GenTxs) > 0 {
		return nil
	}
	return stake.ValidateGenesis(genesisState.StakeData)
}

// Ensures that there are no duplicate accounts in the genesis state,
func validateGenesisStateAccounts(accs []GenesisAccount) (err error) {
	addrMap := make(map[string]bool, len(accs))
	for i := 0; i < len(accs); i++ {
		acc := accs[i]
		strAddr := string(acc.Address)
		if _, ok := addrMap[strAddr]; ok {
			return fmt.Errorf("Duplicate account in genesis state: Address %v", acc.Address)
		}
		addrMap[strAddr] = true
	}
	return
}

// IrisAppGenState but with JSON
func IrisAppGenStateJSON(cdc *codec.Codec, genDoc tmtypes.GenesisDoc, appGenTxs []json.RawMessage) (
	appState json.RawMessage, err error) {

	// create the final app state
	genesisState, err := IrisAppGenState(cdc, genDoc, appGenTxs)
	if err != nil {
		return nil, err
	}
	appState, err = codec.MarshalJSONIndent(cdc, genesisState)
	return
}

// CollectStdTxs processes and validates application's genesis StdTxs and returns
// the list of appGenTxs, and persistent peers required to generate genesis.json.
func CollectStdTxs(cdc *codec.Codec, moniker string, genTxsDir string, genDoc tmtypes.GenesisDoc) (
	appGenTxs []auth.StdTx, persistentPeers string, err error) {

	var fos []os.FileInfo
	fos, err = ioutil.ReadDir(genTxsDir)
	if err != nil {
		return appGenTxs, persistentPeers, err
	}

	// prepare a map of all accounts in genesis state to then validate
	// against the validators addresses
	var appState GenesisState
	if err := cdc.UnmarshalJSON(genDoc.AppState, &appState); err != nil {
		return appGenTxs, persistentPeers, err
	}
	addrMap := make(map[string]GenesisAccount, len(appState.Accounts))
	for i := 0; i < len(appState.Accounts); i++ {
		acc := appState.Accounts[i]
		strAddr := string(acc.Address)
		addrMap[strAddr] = acc
	}

	// addresses and IPs (and port) validator server info
	var addressesIPs []string

	for _, fo := range fos {
		filename := filepath.Join(genTxsDir, fo.Name())
		if !fo.IsDir() && (filepath.Ext(filename) != ".json") {
			continue
		}

		// get the genStdTx
		var jsonRawTx []byte
		if jsonRawTx, err = ioutil.ReadFile(filename); err != nil {
			return appGenTxs, persistentPeers, err
		}
		var genStdTx auth.StdTx
		if err = cdc.UnmarshalJSON(jsonRawTx, &genStdTx); err != nil {
			return appGenTxs, persistentPeers, err
		}
		appGenTxs = append(appGenTxs, genStdTx)

		// the memo flag is used to store
		// the ip and node-id, for example this may be:
		// "528fd3df22b31f4969b05652bfe8f0fe921321d5@192.168.2.37:26656"
		nodeAddrIP := genStdTx.GetMemo()
		if len(nodeAddrIP) == 0 {
			return appGenTxs, persistentPeers, fmt.Errorf(
				"couldn't find node's address and IP in %s", fo.Name())
		}

		// genesis transactions must be single-message
		msgs := genStdTx.GetMsgs()
		if len(msgs) != 1 {

			return appGenTxs, persistentPeers, errors.New(
				"each genesis transaction must provide a single genesis message")
		}

		// validate the validator address and funds against the accounts in the state
		msg := msgs[0].(stake.MsgCreateValidator)
		addr := string(sdk.AccAddress(msg.ValidatorAddr))
		acc, ok := addrMap[addr]
		if !ok {
			return appGenTxs, persistentPeers, fmt.Errorf(
				"account %v not in genesis.json: %+v", addr, addrMap)
		}
		if acc.Coins.AmountOf(msg.Delegation.Denom).LT(msg.Delegation.Amount) {
			err = fmt.Errorf("insufficient fund for the delegation: %s < %s",
				acc.Coins.AmountOf(msg.Delegation.Denom), msg.Delegation.Amount)
		}

		// exclude itself from persistent peers
		if msg.Description.Moniker != moniker {
			addressesIPs = append(addressesIPs, nodeAddrIP)
		}
	}

	sort.Strings(addressesIPs)
	persistentPeers = strings.Join(addressesIPs, ",")

	return appGenTxs, persistentPeers, nil
}

func NewDefaultGenesisAccount(addr sdk.AccAddress) GenesisAccount {
	accAuth := auth.NewBaseAccountWithAddress(addr)
	accAuth.Coins = []sdk.Coin{
		FreeFermionAcc,
	}
	return NewGenesisAccount(&accAuth)
}

func createStakeGenesisState() stake.GenesisState {
	return stake.GenesisState{
		Pool: stake.Pool{
			LooseTokens:  sdk.ZeroDec(),
			BondedTokens: sdk.ZeroDec(),
		},
		Params: stake.Params{
			UnbondingTime: defaultUnbondingTime,
			MaxValidators: 100,
			BondDenom:     StakeDenom,
		},
	}
}

func createMintGenesisState() mint.GenesisState {
	return mint.GenesisState{
		Minter: mint.InitialMinter(),
		Params: mint.Params{
			MintDenom:           StakeDenom,
			InflationRateChange: sdk.NewDecWithPrec(13, 2),
			InflationMax:        sdk.NewDecWithPrec(20, 2),
			InflationMin:        sdk.NewDecWithPrec(7, 2),
			GoalBonded:          sdk.NewDecWithPrec(67, 2),
		},
	}
}
