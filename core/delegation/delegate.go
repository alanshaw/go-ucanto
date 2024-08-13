package delegation

import (
	"fmt"

	"github.com/storacha-network/go-ucanto/core/dag/blockstore"
	"github.com/storacha-network/go-ucanto/core/ipld"
	"github.com/storacha-network/go-ucanto/core/ipld/block"
	"github.com/storacha-network/go-ucanto/core/ipld/codec/cbor"
	"github.com/storacha-network/go-ucanto/core/ipld/hash/sha256"
	"github.com/storacha-network/go-ucanto/ucan"
	udm "github.com/storacha-network/go-ucanto/ucan/datamodel/ucan"
)

// Option is an option configuring a UCAN delegation.
type Option func(cfg *delegationConfig) error

type delegationConfig struct {
	exp uint64
	nbf uint64
	nnc string
	fct []ucan.FactBuilder
	prf Proofs
}

// WithExpiration configures the expiration time in UTC seconds since Unix
// epoch. Set this to -1 for no expiration.
func WithExpiration(exp uint64) Option {
	return func(cfg *delegationConfig) error {
		cfg.exp = exp
		return nil
	}
}

// WithNotBefore configures the time in UTC seconds since Unix epoch when the
// UCAN will become valid.
func WithNotBefore(nbf uint64) Option {
	return func(cfg *delegationConfig) error {
		cfg.nbf = nbf
		return nil
	}
}

// WithNonce configures the nonce value for the UCAN.
func WithNonce(nnc string) Option {
	return func(cfg *delegationConfig) error {
		cfg.nnc = nnc
		return nil
	}
}

// WithFacts configures the facts for the UCAN.
func WithFacts(fct []ucan.FactBuilder) Option {
	return func(cfg *delegationConfig) error {
		cfg.fct = fct
		return nil
	}
}

// WithProofs configures the proofs for the UCAN. If the `issuer` of this
// `Delegation` is not the resource owner / service provider, for the delegated
// capabilities, the `proofs` must contain valid `Proof`s containing
// delegations to the `issuer`.
func WithProofs(prf Proofs) Option {
	return func(cfg *delegationConfig) error {
		cfg.prf = prf
		return nil
	}
}

// Delegate creates a new signed token with a given `options.issuer`. If
// expiration is not set it defaults to 30 seconds from now. Returns UCAN in
// primary IPLD representation.
func Delegate(issuer ucan.Signer, audience ucan.Principal, capabilities []ucan.Capability[ipld.Builder], options ...Option) (Delegation, error) {
	cfg := delegationConfig{}
	for _, opt := range options {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	bs, err := blockstore.NewBlockStore()
	if err != nil {
		return nil, err
	}

	links, err := cfg.prf.WriteInto(bs)
	if err != nil {
		return nil, err
	}

	data, err := ucan.Issue(
		issuer,
		audience,
		capabilities,
		ucan.WithExpiration(cfg.exp),
		ucan.WithFacts(cfg.fct),
		ucan.WithNonce(cfg.nnc),
		ucan.WithNotBefore(cfg.nbf),
		ucan.WithProofs(links),
	)
	if err != nil {
		return nil, fmt.Errorf("issuing UCAN: %s", err)
	}

	rt, err := block.Encode(data.Model(), udm.Type(), cbor.Codec, sha256.Hasher)
	if err != nil {
		return nil, fmt.Errorf("encoding UCAN: %s", err)
	}

	err = bs.Put(rt)
	if err != nil {
		return nil, fmt.Errorf("adding delegation root to store: %s", err)
	}

	return NewDelegation(rt, bs), nil
}
