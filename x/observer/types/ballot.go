package types

import (
	"fmt"

	cosmoserrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
)

func (m Ballot) AddVote(address string, vote VoteType) (Ballot, error) {
	if m.HasVoted(address) {
		return m, cosmoserrors.Wrap(
			ErrUnableToAddVote,
			fmt.Sprintf(" Voter : %s | Ballot :%s | Already Voted", address, m.String()),
		)
	}
	// `index` is the index of the `address` in the `VoterList`
	// `index` is used to set the vote in the `Votes` array
	index := m.GetVoterIndex(address)
	if index == -1 {
		return m, cosmoserrors.Wrap(ErrUnableToAddVote, fmt.Sprintf("Voter %s not in voter list", address))
	}
	m.Votes[index] = vote
	return m, nil
}

func (m Ballot) HasVoted(address string) bool {
	index := m.GetVoterIndex(address)
	if index == -1 {
		return false
	}
	return m.Votes[index] != VoteType_NotYetVoted
}

// GetVoterIndex returns the index of the `address` in the `VoterList`
func (m Ballot) GetVoterIndex(address string) int {
	index := -1
	for i, addr := range m.VoterList {
		if addr == address {
			return i
		}
	}
	return index
}

// IsFinalizingVote checks sets the ballot to a final status if enough votes have been added
// If it has already been finalized it returns false
// It enough votes have not been added it returns false
func (m Ballot) IsFinalizingVote() (Ballot, bool) {
	if m.BallotStatus != BallotStatus_BallotInProgress {
		return m, false
	}
	success, failure := sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()
	total := sdkmath.LegacyNewDec(int64(len(m.VoterList)))
	if total.IsZero() {
		return m, false
	}
	for _, vote := range m.Votes {
		if vote == VoteType_SuccessObservation {
			success = success.Add(sdkmath.LegacyOneDec())
		}
		if vote == VoteType_FailureObservation {
			failure = failure.Add(sdkmath.LegacyOneDec())
		}
	}
	if failure.IsPositive() {
		if failure.Quo(total).GTE(m.BallotThreshold) {
			m.BallotStatus = BallotStatus_BallotFinalized_FailureObservation
			return m, true
		}
	}

	if success.IsPositive() {
		if success.Quo(total).GTE(m.BallotThreshold) {
			m.BallotStatus = BallotStatus_BallotFinalized_SuccessObservation
			return m, true
		}
	}
	return m, false
}

func CreateVotes(listSize int) []VoteType {
	voterList := make([]VoteType, listSize)
	for i := range voterList {
		voterList[i] = VoteType_NotYetVoted
	}
	return voterList
}

// BuildRewardsDistribution builds the rewards distribution map for the ballot
// It returns the total rewards units which account for the observer block rewards
func (m Ballot) BuildRewardsDistribution(rewardsMap map[string]int64) int64 {
	totalRewardUnits := int64(0)
	switch m.BallotStatus {
	case BallotStatus_BallotFinalized_SuccessObservation:
		for _, address := range m.VoterList {
			vote := m.Votes[m.GetVoterIndex(address)]
			if vote == VoteType_SuccessObservation {
				rewardsMap[address] = rewardsMap[address] + 1
				totalRewardUnits++
				continue
			}
			rewardsMap[address] = rewardsMap[address] - 1
		}
	case BallotStatus_BallotFinalized_FailureObservation:
		for _, address := range m.VoterList {
			vote := m.Votes[m.GetVoterIndex(address)]
			if vote == VoteType_FailureObservation {
				rewardsMap[address] = rewardsMap[address] + 1
				totalRewardUnits++
				continue
			}
			rewardsMap[address] = rewardsMap[address] - 1
		}
	}
	return totalRewardUnits
}
