// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package ocr_2

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// SetPayees is the `setPayees` instruction.
type SetPayees struct {
	Payees *[]ag_solanago.PublicKey

	// [0] = [WRITE] state
	//
	// [1] = [SIGNER] authority
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewSetPayeesInstructionBuilder creates a new `SetPayees` instruction builder.
func NewSetPayeesInstructionBuilder() *SetPayees {
	nd := &SetPayees{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 2),
	}
	return nd
}

// SetPayees sets the "payees" parameter.
func (inst *SetPayees) SetPayees(payees []ag_solanago.PublicKey) *SetPayees {
	inst.Payees = &payees
	return inst
}

// SetStateAccount sets the "state" account.
func (inst *SetPayees) SetStateAccount(state ag_solanago.PublicKey) *SetPayees {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(state).WRITE()
	return inst
}

// GetStateAccount gets the "state" account.
func (inst *SetPayees) GetStateAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

// SetAuthorityAccount sets the "authority" account.
func (inst *SetPayees) SetAuthorityAccount(authority ag_solanago.PublicKey) *SetPayees {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(authority).SIGNER()
	return inst
}

// GetAuthorityAccount gets the "authority" account.
func (inst *SetPayees) GetAuthorityAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[1]
}

func (inst SetPayees) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_SetPayees,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst SetPayees) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *SetPayees) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Payees == nil {
			return errors.New("Payees parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.State is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Authority is not set")
		}
	}
	return nil
}

func (inst *SetPayees) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("SetPayees")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=1]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Payees", *inst.Payees))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=2]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("    state", inst.AccountMetaSlice[0]))
						accountsBranch.Child(ag_format.Meta("authority", inst.AccountMetaSlice[1]))
					})
				})
		})
}

func (obj SetPayees) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Payees` param:
	err = encoder.Encode(obj.Payees)
	if err != nil {
		return err
	}
	return nil
}
func (obj *SetPayees) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Payees`:
	err = decoder.Decode(&obj.Payees)
	if err != nil {
		return err
	}
	return nil
}

// NewSetPayeesInstruction declares a new SetPayees instruction with the provided parameters and accounts.
func NewSetPayeesInstruction(
	// Parameters:
	payees []ag_solanago.PublicKey,
	// Accounts:
	state ag_solanago.PublicKey,
	authority ag_solanago.PublicKey) *SetPayees {
	return NewSetPayeesInstructionBuilder().
		SetPayees(payees).
		SetStateAccount(state).
		SetAuthorityAccount(authority)
}
