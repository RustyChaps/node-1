package promise

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mysteriumnetwork/node/identity"
	"github.com/mysteriumnetwork/payments/promises"
)

// Issuer interface defines method to sign (issue) provided promise data and return promise with signature
// used by promise issuer (i.e. service consumer or 3d party)
type Issuer interface {
	Issue(promise promises.Promise) (promises.IssuedPromise, error)
}

// State defines current state of promise data (seq number and amount)
type State struct {
	seq    int64
	amount int64
}

// ConsumerTracker tracks and issues promises from consumer perspective, also validates states coming from service provider
type ConsumerTracker struct {
	current  State
	consumer identity.Identity
	receiver identity.Identity
	issuer   Issuer
}

func NewConsumerTracker(initial State, consumer, provider identity.Identity, issuer Issuer) *ConsumerTracker {
	return &ConsumerTracker{
		current:  initial,
		consumer: consumer,
		receiver: provider,
		issuer:   issuer,
	}
}

var UnexpectedAmount = errors.New("unexpected amount")

func (t *ConsumerTracker) AlignStateWithProvider(providerState State) error {
	if providerState.seq > t.current.seq {
		// new promise request
		t.current.seq = providerState.seq
		// ignore provider state value as new promise amount is always zero,
		// if provider tries to trick us to send more than expected it will be caught on next align
		t.current.amount = 0
		return nil
	}
	if providerState.amount > t.current.amount {
		return UnexpectedAmount
	}
	if providerState.amount < t.current.amount {
		return UnexpectedAmount
	}

	return nil
}

func (t *ConsumerTracker) IssuePromiseWithAddedAmount(amountToAdd int64) (promises.IssuedPromise, error) {

	promise := promises.Promise{
		Extra: ExtraData{
			ConsumerAddress: common.HexToAddress(t.consumer.Address),
		},
		Receiver: common.HexToAddress(t.receiver.Address),
		Amount:   t.current.amount + amountToAdd,
		SeqNo:    t.current.seq,
	}
	return t.issuer.Issue(promise)
}
