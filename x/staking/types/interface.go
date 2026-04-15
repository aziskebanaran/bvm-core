package types

type StakingKeeper interface {
    Stake(addr string, amount uint64) error
    Unstake(addr string, amount uint64) error
    Delegate(delegator, validator string, amount uint64) error
    AutoDelegate(addr string, amount uint64) error

    ModifyValidatorPower(address string, amount uint64, isAdding bool) error
    GetValidatorPower(address string) uint64

    // ✅ Bersih dari awalan package
    GetValidators() []Validator
    GetTopValidators(n int) []Validator

    GetValidatorStake(addr string) uint64
    ProcessIncentive(addr string, amount uint64) error

    // 🚩 PERBAIKAN DI SINI:
    GetValidatorObjects() ([]Validator, error)
    QueryTopValidators(n int) []*Validator
}
