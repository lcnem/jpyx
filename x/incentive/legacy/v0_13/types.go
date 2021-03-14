package v0_13

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	tmtime "github.com/tendermint/tendermint/types/time"

	cdptypes "github.com/lcnem/jpyx/x/cdp/types"
	jsmndistTypes "github.com/lcnem/jpyx/x/jsmndist/types"
)

// Valid reward multipliers
const (
	Small                          MultiplierName = "small"
	Medium                         MultiplierName = "medium"
	Large                          MultiplierName = "large"
	JPYXMintingClaimType                          = "jpyx_minting"
	HardLiquidityProviderClaimType                = "hard_liquidity_provider"
	BondDenom                                     = "ujsmn"
)

// Parameter keys and default values
var (
	KeyJPYXMintingRewardPeriods     = []byte("JPYXMintingRewardPeriods")
	KeyHardSupplyRewardPeriods      = []byte("HardSupplyRewardPeriods")
	KeyHardBorrowRewardPeriods      = []byte("HardBorrowRewardPeriods")
	KeyHardDelegatorRewardPeriods   = []byte("HardDelegatorRewardPeriods")
	KeyClaimEnd                     = []byte("ClaimEnd")
	KeyMultipliers                  = []byte("ClaimMultipliers")
	DefaultActive                   = false
	DefaultRewardPeriods            = RewardPeriods{}
	DefaultMultipliers              = Multipliers{}
	DefaultJPYXClaims               = JPYXMintingClaims{}
	DefaultHardClaims               = HardLiquidityProviderClaims{}
	DefaultGenesisAccumulationTimes = GenesisAccumulationTimes{}
	DefaultClaimEnd                 = tmtime.Canonical(time.Unix(0, 0))
	GovDenom                        = cdptypes.DefaultGovDenom
	PrincipalDenom                  = "jpyx"
	IncentiveMacc                   = jsmndistTypes.ModuleName
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	Params                         Params                      `json:"params" yaml:"params"`
	JPYXAccumulationTimes          GenesisAccumulationTimes    `json:"jpyx_accumulation_times" yaml:"jpyx_accumulation_times"`
	HardSupplyAccumulationTimes    GenesisAccumulationTimes    `json:"hard_supply_accumulation_times" yaml:"hard_supply_accumulation_times"`
	HardBorrowAccumulationTimes    GenesisAccumulationTimes    `json:"hard_borrow_accumulation_times" yaml:"hard_borrow_accumulation_times"`
	HardDelegatorAccumulationTimes GenesisAccumulationTimes    `json:"hard_delegator_accumulation_times" yaml:"hard_delegator_accumulation_times"`
	JPYXMintingClaims              JPYXMintingClaims           `json:"jpyx_minting_claims" yaml:"jpyx_minting_claims"`
	HardLiquidityProviderClaims    HardLiquidityProviderClaims `json:"hard_liquidity_provider_claims" yaml:"hard_liquidity_provider_claims"`
}

// NewGenesisState returns a new genesis state
func NewGenesisState(params Params, jpyxAccumTimes, hardSupplyAccumTimes, hardBorrowAccumTimes, hardDelegatorAccumTimes GenesisAccumulationTimes, c JPYXMintingClaims) GenesisState {
	return GenesisState{
		Params:                         params,
		JPYXAccumulationTimes:          jpyxAccumTimes,
		HardSupplyAccumulationTimes:    hardSupplyAccumTimes,
		HardBorrowAccumulationTimes:    hardBorrowAccumTimes,
		HardDelegatorAccumulationTimes: hardDelegatorAccumTimes,
		JPYXMintingClaims:              c,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:                         DefaultParams(),
		JPYXAccumulationTimes:          GenesisAccumulationTimes{},
		HardSupplyAccumulationTimes:    GenesisAccumulationTimes{},
		HardBorrowAccumulationTimes:    GenesisAccumulationTimes{},
		HardDelegatorAccumulationTimes: GenesisAccumulationTimes{},
		JPYXMintingClaims:              DefaultJPYXClaims,
		HardLiquidityProviderClaims:    DefaultHardClaims,
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if err := gs.JPYXAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.HardSupplyAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.HardBorrowAccumulationTimes.Validate(); err != nil {
		return err
	}
	if err := gs.HardDelegatorAccumulationTimes.Validate(); err != nil {
		return err
	}

	if err := gs.HardLiquidityProviderClaims.Validate(); err != nil {
		return err
	}
	return gs.JPYXMintingClaims.Validate()
}

// GenesisAccumulationTime stores the previous reward distribution time and its corresponding collateral type
type GenesisAccumulationTime struct {
	CollateralType           string    `json:"collateral_type" yaml:"collateral_type"`
	PreviousAccumulationTime time.Time `json:"previous_accumulation_time" yaml:"previous_accumulation_time"`
	RewardFactor             sdk.Dec   `json:"reward_factor" yaml:"reward_factor"`
}

// NewGenesisAccumulationTime returns a new GenesisAccumulationTime
func NewGenesisAccumulationTime(ctype string, prevTime time.Time, factor sdk.Dec) GenesisAccumulationTime {
	return GenesisAccumulationTime{
		CollateralType:           ctype,
		PreviousAccumulationTime: prevTime,
		RewardFactor:             factor,
	}
}

// GenesisAccumulationTimes slice of GenesisAccumulationTime
type GenesisAccumulationTimes []GenesisAccumulationTime

// Validate performs validation of GenesisAccumulationTimes
func (gats GenesisAccumulationTimes) Validate() error {
	for _, gat := range gats {
		if err := gat.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate performs validation of GenesisAccumulationTime
func (gat GenesisAccumulationTime) Validate() error {
	if gat.RewardFactor.LT(sdk.ZeroDec()) {
		return fmt.Errorf("reward factor should be ≥ 0.0, is %s for %s", gat.RewardFactor, gat.CollateralType)
	}
	return nil
}

// Params governance parameters for the incentive module
type Params struct {
	JPYXMintingRewardPeriods   RewardPeriods `json:"jpyx_minting_reward_periods" yaml:"jpyx_minting_reward_periods"`
	HardSupplyRewardPeriods    RewardPeriods `json:"hard_supply_reward_periods" yaml:"hard_supply_reward_periods"`
	HardBorrowRewardPeriods    RewardPeriods `json:"hard_borrow_reward_periods" yaml:"hard_borrow_reward_periods"`
	HardDelegatorRewardPeriods RewardPeriods `json:"hard_delegator_reward_periods" yaml:"hard_delegator_reward_periods"`
	ClaimMultipliers           Multipliers   `json:"claim_multipliers" yaml:"claim_multipliers"`
	ClaimEnd                   time.Time     `json:"claim_end" yaml:"claim_end"`
}

// NewParams returns a new params object
func NewParams(jpyxMinting, hardSupply, hardBorrow, hardDelegator RewardPeriods,
	multipliers Multipliers, claimEnd time.Time) Params {
	return Params{
		JPYXMintingRewardPeriods:   jpyxMinting,
		HardSupplyRewardPeriods:    hardSupply,
		HardBorrowRewardPeriods:    hardBorrow,
		HardDelegatorRewardPeriods: hardDelegator,
		ClaimMultipliers:           multipliers,
		ClaimEnd:                   claimEnd,
	}
}

// DefaultParams returns default params for incentive module
func DefaultParams() Params {
	return NewParams(DefaultRewardPeriods, DefaultRewardPeriods,
		DefaultRewardPeriods, DefaultRewardPeriods, DefaultMultipliers, DefaultClaimEnd)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyJPYXMintingRewardPeriods, &p.JPYXMintingRewardPeriods, validateRewardPeriodsParam),
		params.NewParamSetPair(KeyHardSupplyRewardPeriods, &p.HardSupplyRewardPeriods, validateRewardPeriodsParam),
		params.NewParamSetPair(KeyHardBorrowRewardPeriods, &p.HardBorrowRewardPeriods, validateRewardPeriodsParam),
		params.NewParamSetPair(KeyHardDelegatorRewardPeriods, &p.HardDelegatorRewardPeriods, validateRewardPeriodsParam),
		params.NewParamSetPair(KeyClaimEnd, &p.ClaimEnd, validateClaimEndParam),
		params.NewParamSetPair(KeyMultipliers, &p.ClaimMultipliers, validateMultipliersParam),
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {

	if err := validateMultipliersParam(p.ClaimMultipliers); err != nil {
		return err
	}

	if err := validateRewardPeriodsParam(p.JPYXMintingRewardPeriods); err != nil {
		return err
	}

	if err := validateRewardPeriodsParam(p.HardSupplyRewardPeriods); err != nil {
		return err
	}

	if err := validateRewardPeriodsParam(p.HardBorrowRewardPeriods); err != nil {
		return err
	}

	return validateRewardPeriodsParam(p.HardDelegatorRewardPeriods)
}

func validateRewardPeriodsParam(i interface{}) error {
	rewards, ok := i.(RewardPeriods)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return rewards.Validate()
}

func validateMultipliersParam(i interface{}) error {
	multipliers, ok := i.(Multipliers)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return multipliers.Validate()
}

func validateClaimEndParam(i interface{}) error {
	endTime, ok := i.(time.Time)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if endTime.IsZero() {
		return fmt.Errorf("end time should not be zero")
	}
	return nil
}

// RewardPeriod stores the state of an ongoing reward
type RewardPeriod struct {
	Active           bool      `json:"active" yaml:"active"`
	CollateralType   string    `json:"collateral_type" yaml:"collateral_type"`
	Start            time.Time `json:"start" yaml:"start"`
	End              time.Time `json:"end" yaml:"end"`
	RewardsPerSecond sdk.Coin  `json:"rewards_per_second" yaml:"rewards_per_second"` // per second reward payouts
}

// String implements fmt.Stringer
func (rp RewardPeriod) String() string {
	return fmt.Sprintf(`Reward Period:
	Collateral Type: %s,
	Start: %s,
	End: %s,
	Rewards Per Second: %s,
	Active %t,
	`, rp.CollateralType, rp.Start, rp.End, rp.RewardsPerSecond, rp.Active)
}

// NewRewardPeriod returns a new RewardPeriod
func NewRewardPeriod(active bool, collateralType string, start time.Time, end time.Time, reward sdk.Coin) RewardPeriod {
	return RewardPeriod{
		Active:           active,
		CollateralType:   collateralType,
		Start:            start,
		End:              end,
		RewardsPerSecond: reward,
	}
}

// Validate performs a basic check of a RewardPeriod fields.
func (rp RewardPeriod) Validate() error {
	if rp.Start.IsZero() {
		return errors.New("reward period start time cannot be 0")
	}
	if rp.End.IsZero() {
		return errors.New("reward period end time cannot be 0")
	}
	if rp.Start.After(rp.End) {
		return fmt.Errorf("end period time %s cannot be before start time %s", rp.End, rp.Start)
	}
	if !rp.RewardsPerSecond.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", rp.RewardsPerSecond)
	}
	if strings.TrimSpace(rp.CollateralType) == "" {
		return fmt.Errorf("reward period collateral type cannot be blank: %s", rp)
	}
	return nil
}

// RewardPeriods array of RewardPeriod
type RewardPeriods []RewardPeriod

// Validate checks if all the RewardPeriods are valid and there are no duplicated
// entries.
func (rps RewardPeriods) Validate() error {
	seenPeriods := make(map[string]bool)
	for _, rp := range rps {
		if seenPeriods[rp.CollateralType] {
			return fmt.Errorf("duplicated reward period with collateral type %s", rp.CollateralType)
		}

		if err := rp.Validate(); err != nil {
			return err
		}
		seenPeriods[rp.CollateralType] = true
	}

	return nil
}

// Multiplier amount the claim rewards get increased by, along with how long the claim rewards are locked
type Multiplier struct {
	Name         MultiplierName `json:"name" yaml:"name"`
	MonthsLockup int64          `json:"months_lockup" yaml:"months_lockup"`
	Factor       sdk.Dec        `json:"factor" yaml:"factor"`
}

// NewMultiplier returns a new Multiplier
func NewMultiplier(name MultiplierName, lockup int64, factor sdk.Dec) Multiplier {
	return Multiplier{
		Name:         name,
		MonthsLockup: lockup,
		Factor:       factor,
	}
}

// Validate multiplier param
func (m Multiplier) Validate() error {
	if err := m.Name.IsValid(); err != nil {
		return err
	}
	if m.MonthsLockup < 0 {
		return fmt.Errorf("expected non-negative lockup, got %d", m.MonthsLockup)
	}
	if m.Factor.IsNegative() {
		return fmt.Errorf("expected non-negative factor, got %s", m.Factor.String())
	}

	return nil
}

// String implements fmt.Stringer
func (m Multiplier) String() string {
	return fmt.Sprintf(`Claim Multiplier:
	Name: %s
	Months Lockup %d
	Factor %s
	`, m.Name, m.MonthsLockup, m.Factor)
}

// Multipliers slice of Multiplier
type Multipliers []Multiplier

// Validate validates each multiplier
func (ms Multipliers) Validate() error {
	for _, m := range ms {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// String implements fmt.Stringer
func (ms Multipliers) String() string {
	out := "Claim Multipliers\n"
	for _, s := range ms {
		out += fmt.Sprintf("%s\n", s)
	}
	return out
}

// MultiplierName name for valid multiplier
type MultiplierName string

// IsValid checks if the input is one of the expected strings
func (mn MultiplierName) IsValid() error {
	switch mn {
	case Small, Medium, Large:
		return nil
	}
	return fmt.Errorf("invalid multiplier name: %s", mn)
}

// Claim is an interface for handling common claim actions
type Claim interface {
	GetOwner() sdk.AccAddress
	GetReward() sdk.Coin
	GetType() string
}

// Claims is a slice of Claim
type Claims []Claim

// BaseClaim is a common type shared by all Claims
type BaseClaim struct {
	Owner  sdk.AccAddress `json:"owner" yaml:"owner"`
	Reward sdk.Coin       `json:"reward" yaml:"reward"`
}

// GetOwner is a getter for Claim Owner
func (c BaseClaim) GetOwner() sdk.AccAddress { return c.Owner }

// GetReward is a getter for Claim Reward
func (c BaseClaim) GetReward() sdk.Coin { return c.Reward }

// GetType returns the claim type, used to identify auctions in event attributes
func (c BaseClaim) GetType() string { return "base" }

// Validate performs a basic check of a BaseClaim fields
func (c BaseClaim) Validate() error {
	if c.Owner.Empty() {
		return errors.New("claim owner cannot be empty")
	}
	if !c.Reward.IsValid() {
		return fmt.Errorf("invalid reward amount: %s", c.Reward)
	}
	return nil
}

// String implements fmt.Stringer
func (c BaseClaim) String() string {
	return fmt.Sprintf(`Claim:
	Owner: %s,
	Reward: %s,
	`, c.Owner, c.Reward)
}

// -------------- Custom Claim Types --------------

// JPYXMintingClaim is for JPYX minting rewards
type JPYXMintingClaim struct {
	BaseClaim     `json:"base_claim" yaml:"base_claim"`
	RewardIndexes RewardIndexes `json:"reward_indexes" yaml:"reward_indexes"`
}

// NewJPYXMintingClaim returns a new JPYXMintingClaim
func NewJPYXMintingClaim(owner sdk.AccAddress, reward sdk.Coin, rewardIndexes RewardIndexes) JPYXMintingClaim {
	return JPYXMintingClaim{
		BaseClaim: BaseClaim{
			Owner:  owner,
			Reward: reward,
		},
		RewardIndexes: rewardIndexes,
	}
}

// GetType returns the claim's type
func (c JPYXMintingClaim) GetType() string { return JPYXMintingClaimType }

// GetReward returns the claim's reward coin
func (c JPYXMintingClaim) GetReward() sdk.Coin { return c.Reward }

// GetOwner returns the claim's owner
func (c JPYXMintingClaim) GetOwner() sdk.AccAddress { return c.Owner }

// Validate performs a basic check of a Claim fields
func (c JPYXMintingClaim) Validate() error {
	if err := c.RewardIndexes.Validate(); err != nil {
		return err
	}

	return c.BaseClaim.Validate()
}

// String implements fmt.Stringer
func (c JPYXMintingClaim) String() string {
	return fmt.Sprintf(`%s
	Reward Indexes: %s,
	`, c.BaseClaim, c.RewardIndexes)
}

// HasRewardIndex check if a claim has a reward index for the input collateral type
func (c JPYXMintingClaim) HasRewardIndex(collateralType string) (int64, bool) {
	for index, ri := range c.RewardIndexes {
		if ri.CollateralType == collateralType {
			return int64(index), true
		}
	}
	return 0, false
}

// JPYXMintingClaims slice of JPYXMintingClaim
type JPYXMintingClaims []JPYXMintingClaim

// Validate checks if all the claims are valid and there are no duplicated
// entries.
func (cs JPYXMintingClaims) Validate() error {
	for _, c := range cs {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// HardLiquidityProviderClaim stores the hard liquidity provider rewards that can be claimed by owner
type HardLiquidityProviderClaim struct {
	BaseClaim              `json:"base_claim" yaml:"base_claim"`
	SupplyRewardIndexes    RewardIndexes `json:"supply_reward_indexes" yaml:"supply_reward_indexes"`
	BorrowRewardIndexes    RewardIndexes `json:"borrow_reward_indexes" yaml:"borrow_reward_indexes"`
	DelegatorRewardIndexes RewardIndexes `json:"delegator_reward_indexes" yaml:"delegator_reward_indexes"`
}

// NewHardLiquidityProviderClaim returns a new HardLiquidityProviderClaim
func NewHardLiquidityProviderClaim(owner sdk.AccAddress, reward sdk.Coin, supplyRewardIndexes,
	borrowRewardIndexes, delegatorRewardIndexes RewardIndexes) HardLiquidityProviderClaim {
	return HardLiquidityProviderClaim{
		BaseClaim: BaseClaim{
			Owner:  owner,
			Reward: reward,
		},
		SupplyRewardIndexes:    supplyRewardIndexes,
		BorrowRewardIndexes:    borrowRewardIndexes,
		DelegatorRewardIndexes: delegatorRewardIndexes,
	}
}

// GetType returns the claim's type
func (c HardLiquidityProviderClaim) GetType() string { return HardLiquidityProviderClaimType }

// GetReward returns the claim's reward coin
func (c HardLiquidityProviderClaim) GetReward() sdk.Coin { return c.Reward }

// GetOwner returns the claim's owner
func (c HardLiquidityProviderClaim) GetOwner() sdk.AccAddress { return c.Owner }

// Validate performs a basic check of a HardLiquidityProviderClaim fields
func (c HardLiquidityProviderClaim) Validate() error {
	if err := c.SupplyRewardIndexes.Validate(); err != nil {
		return err
	}

	if err := c.BorrowRewardIndexes.Validate(); err != nil {
		return err
	}

	if err := c.DelegatorRewardIndexes.Validate(); err != nil {
		return err
	}

	return c.BaseClaim.Validate()
}

// String implements fmt.Stringer
func (c HardLiquidityProviderClaim) String() string {
	return fmt.Sprintf(`%s
	Supply Reward Indexes: %s,
	Borrow Reward Indexes: %s,
	Delegator Reward Indexes: %s,
	`, c.BaseClaim, c.SupplyRewardIndexes, c.BorrowRewardIndexes, c.DelegatorRewardIndexes)
}

// HasSupplyRewardIndex check if a claim has a supply reward index for the input collateral type
func (c HardLiquidityProviderClaim) HasSupplyRewardIndex(denom string) (int64, bool) {
	for index, ri := range c.SupplyRewardIndexes {
		if ri.CollateralType == denom {
			return int64(index), true
		}
	}
	return 0, false
}

// HasBorrowRewardIndex check if a claim has a borrow reward index for the input collateral type
func (c HardLiquidityProviderClaim) HasBorrowRewardIndex(denom string) (int64, bool) {
	for index, ri := range c.BorrowRewardIndexes {
		if ri.CollateralType == denom {
			return int64(index), true
		}
	}
	return 0, false
}

// HasDelegatorRewardIndex check if a claim has a delegator reward index for the input collateral type
func (c HardLiquidityProviderClaim) HasDelegatorRewardIndex(collateralType string) (int64, bool) {
	for index, ri := range c.DelegatorRewardIndexes {
		if ri.CollateralType == collateralType {
			return int64(index), true
		}
	}
	return 0, false
}

// HardLiquidityProviderClaims slice of HardLiquidityProviderClaim
type HardLiquidityProviderClaims []HardLiquidityProviderClaim

// Validate checks if all the claims are valid and there are no duplicated
// entries.
func (cs HardLiquidityProviderClaims) Validate() error {
	for _, c := range cs {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// -------------- Subcomponents of Custom Claim Types --------------

// TODO: refactor RewardPeriod name from 'collateralType' to 'denom'

// RewardIndex stores reward accumulation information
type RewardIndex struct {
	CollateralType string  `json:"collateral_type" yaml:"collateral_type"`
	RewardFactor   sdk.Dec `json:"reward_factor" yaml:"reward_factor"`
}

// NewRewardIndex returns a new RewardIndex
func NewRewardIndex(collateralType string, factor sdk.Dec) RewardIndex {
	return RewardIndex{
		CollateralType: collateralType,
		RewardFactor:   factor,
	}
}

func (ri RewardIndex) String() string {
	return fmt.Sprintf(`Collateral Type: %s, RewardFactor: %s`, ri.CollateralType, ri.RewardFactor)
}

// Validate validates reward index
func (ri RewardIndex) Validate() error {
	if ri.RewardFactor.IsNegative() {
		return fmt.Errorf("reward factor value should be positive, is %s for %s", ri.RewardFactor, ri.CollateralType)
	}
	if strings.TrimSpace(ri.CollateralType) == "" {
		return fmt.Errorf("collateral type should not be empty")
	}
	return nil
}

// RewardIndexes slice of RewardIndex
type RewardIndexes []RewardIndex

// Validate validation for reward indexes
func (ris RewardIndexes) Validate() error {
	for _, ri := range ris {
		if err := ri.Validate(); err != nil {
			return err
		}
	}
	return nil
}
