package types

import (
	"bytes"
	"fmt"

	"github.com/bcbchain/bclib/tendermint/go-amino"
	"github.com/bcbchain/bclib/tendermint/go-crypto"
	"github.com/bcbchain/bclib/tendermint/tmlibs/merkle"
)

// ErrEvidenceInvalid wraps a piece of evidence and the error denoting how or why it is invalid.
type ErrEvidenceInvalid struct {
	Evidence   Evidence
	ErrorValue error
}

func NewEvidenceInvalidErr(ev Evidence, err error) *ErrEvidenceInvalid {
	return &ErrEvidenceInvalid{ev, err}
}

// Error returns a string representation of the error.
func (err *ErrEvidenceInvalid) Error() string {
	return fmt.Sprintf("Invalid evidence: %v. Evidence: %v", err.ErrorValue, err.Evidence)
}

//-------------------------------------------

// Evidence represents any provable malicious activity by a validator
type Evidence interface {
	Height() int64               // height of the equivocation
	Address() string             // address of the equivocating validator
	Index() int                  // index of the validator in the validator set
	Hash() []byte                // hash of the evidence
	Verify(chainID string) error // verify the evidence
	Equal(Evidence) bool         // check equality of evidence

	String() string
}

func RegisterEvidences(cdc *amino.Codec) {
	cdc.RegisterInterface((*Evidence)(nil), nil)
	cdc.RegisterConcrete(&DuplicateVoteEvidence{}, "tendermint/DuplicateVoteEvidence", nil)
	cdc.RegisterConcrete(&TooManyProposalEvidence{}, "tendermint/TooManyProposalEvidence", nil)
}

//-------------------------------------------

// DuplicateVoteEvidence contains evidence a validator signed two conflicting votes.
type DuplicateVoteEvidence struct {
	PubKey crypto.PubKey
	VoteA  *Vote
	VoteB  *Vote
}

// String returns a string representation of the evidence.
func (dve *DuplicateVoteEvidence) String() string {
	return fmt.Sprintf("VoteA: %v; VoteB: %v", dve.VoteA, dve.VoteB)

}

// Height returns the height this evidence refers to.
func (dve *DuplicateVoteEvidence) Height() int64 {
	return dve.VoteA.Height
}

// Address returns the address of the validator.
func (dve *DuplicateVoteEvidence) Address() string {
	return dve.PubKey.Address(crypto.GetChainId())
}

// Index returns the index of the validator.
func (dve *DuplicateVoteEvidence) Index() int {
	return dve.VoteA.ValidatorIndex
}

// Hash returns the hash of the evidence.
func (dve *DuplicateVoteEvidence) Hash() []byte {
	return aminoHasher(dve).Hash()
}

// Verify returns an error if the two votes aren't conflicting.
// To be conflicting, they must be from the same validator, for the same H/R/S, but for different blocks.
func (dve *DuplicateVoteEvidence) Verify(chainID string) error {
	// H/R/S must be the same
	if dve.VoteA.Height != dve.VoteB.Height ||
		dve.VoteA.Round != dve.VoteB.Round ||
		dve.VoteA.Type != dve.VoteB.Type {
		return fmt.Errorf("DuplicateVoteEvidence Error: H/R/S does not match. Got %v and %v", dve.VoteA, dve.VoteB)
	}

	// Address must be the same
	if dve.VoteA.ValidatorAddress == dve.VoteB.ValidatorAddress {
		return fmt.Errorf("DuplicateVoteEvidence Error: Validator addresses do not match. Got %X and %X", dve.VoteA.ValidatorAddress, dve.VoteB.ValidatorAddress)
	}
	// XXX: Should we enforce index is the same ?
	if dve.VoteA.ValidatorIndex != dve.VoteB.ValidatorIndex {
		return fmt.Errorf("DuplicateVoteEvidence Error: Validator indices do not match. Got %d and %d", dve.VoteA.ValidatorIndex, dve.VoteB.ValidatorIndex)
	}

	// BlockIDs must be different
	if dve.VoteA.BlockID.Equals(dve.VoteB.BlockID) {
		return fmt.Errorf("DuplicateVoteEvidence Error: BlockIDs are the same (%v) - not a real duplicate vote", dve.VoteA.BlockID)
	}

	// Signatures must be valid
	if !dve.PubKey.VerifyBytes(dve.VoteA.SignBytes(chainID), dve.VoteA.Signature) {
		return fmt.Errorf("DuplicateVoteEvidence Error verifying VoteA: %v", ErrVoteInvalidSignature)
	}
	if !dve.PubKey.VerifyBytes(dve.VoteB.SignBytes(chainID), dve.VoteB.Signature) {
		return fmt.Errorf("DuplicateVoteEvidence Error verifying VoteB: %v", ErrVoteInvalidSignature)
	}

	return nil
}

// Equal checks if two pieces of evidence are equal.
func (dve *DuplicateVoteEvidence) Equal(ev Evidence) bool {
	if _, ok := ev.(*DuplicateVoteEvidence); !ok {
		return false
	}

	// just check their hashes
	dveHash := aminoHasher(dve).Hash()
	evHash := aminoHasher(ev).Hash()
	return bytes.Equal(dveHash, evHash)
}

//-----------------------------------------------------------------

// TooManyProposalEvidence contains evidence: a validator propose too many times, disproportional to its voting power.
type TooManyProposalEvidence struct {
	PubKey crypto.PubKey
	H      int64
}

func (tmpe *TooManyProposalEvidence) Height() int64 {
	return tmpe.H
}
func (tmpe *TooManyProposalEvidence) Address() string {
	return tmpe.PubKey.Address(crypto.GetChainId())
}
func (tmpe *TooManyProposalEvidence) Index() int {
	return 0
}
func (tmpe *TooManyProposalEvidence) Hash() []byte {
	return aminoHasher(tmpe).Hash()
}
func (tmpe *TooManyProposalEvidence) Verify(chainID string) error {
	return nil
}

func (tmpe *TooManyProposalEvidence) Equal(ev Evidence) bool {
	if _, ok := ev.(*TooManyProposalEvidence); !ok {
		return false
	}

	dveHash := aminoHasher(tmpe).Hash()
	evHash := aminoHasher(ev).Hash()
	return bytes.Equal(dveHash, evHash)
}

func (tmpe *TooManyProposalEvidence) String() string {
	return fmt.Sprintf("TooManyProposal: %v at height: %d", tmpe.PubKey.Address(crypto.GetChainId()), tmpe.H)
}

//-----------------------------------------------------------------

// UNSTABLE
type MockGoodEvidence struct {
	Height_  int64
	Address_ string
	Index_   int
}

// UNSTABLE
func NewMockGoodEvidence(height int64, index int, address string) MockGoodEvidence {
	return MockGoodEvidence{height, address, index}
}

func (e MockGoodEvidence) Height() int64   { return e.Height_ }
func (e MockGoodEvidence) Address() string { return e.Address_ }
func (e MockGoodEvidence) Index() int      { return e.Index_ }
func (e MockGoodEvidence) Hash() []byte {
	return []byte(fmt.Sprintf("%d-%d", e.Height_, e.Index_))
}
func (e MockGoodEvidence) Verify(chainID string) error { return nil }
func (e MockGoodEvidence) Equal(ev Evidence) bool {
	e2 := ev.(MockGoodEvidence)
	return e.Height_ == e2.Height_ &&
		e.Address_ == e2.Address_ &&
		e.Index_ == e2.Index_
}
func (e MockGoodEvidence) String() string {
	return fmt.Sprintf("GoodEvidence: %d/%s/%d", e.Height_, e.Address_, e.Index_)
}

// UNSTABLE
type MockBadEvidence struct {
	MockGoodEvidence
}

func (e MockBadEvidence) Verify(chainID string) error { return fmt.Errorf("MockBadEvidence") }
func (e MockBadEvidence) Equal(ev Evidence) bool {
	e2 := ev.(MockBadEvidence)
	return e.Height_ == e2.Height_ &&
		e.Address_ == e2.Address_ &&
		e.Index_ == e2.Index_
}
func (e MockBadEvidence) String() string {
	return fmt.Sprintf("BadEvidence: %d/%s/%d", e.Height_, e.Address_, e.Index_)
}

//-------------------------------------------

// EvidenceList is a list of Evidence. Evidences is not a word.
type EvidenceList []Evidence

// Hash returns the simple merkle root hash of the EvidenceList.
func (evl EvidenceList) Hash() []byte {
	// Recursive impl.
	// Copied from tmlibs/merkle to avoid allocations
	switch len(evl) {
	case 0:
		return nil
	case 1:
		return evl[0].Hash()
	default:
		left := EvidenceList(evl[:(len(evl)+1)/2]).Hash()
		right := EvidenceList(evl[(len(evl)+1)/2:]).Hash()
		return merkle.SimpleHashFromTwoHashes(left, right)
	}
}

func (evl EvidenceList) String() string {
	s := ""
	for _, e := range evl {
		s += fmt.Sprintf("%s\t\t", e)
	}
	return s
}

// Has returns true if the evidence is in the EvidenceList.
func (evl EvidenceList) Has(evidence Evidence) bool {
	for _, ev := range evl {
		if ev.Equal(evidence) {
			return true
		}
	}
	return false
}
