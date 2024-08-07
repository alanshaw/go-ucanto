package delegation

import (
	"github.com/storacha-network/go-ucanto/core/dag/blockstore"
	"github.com/storacha-network/go-ucanto/core/ipld"
	"github.com/storacha-network/go-ucanto/ucan"
)

type Proof struct {
	delegation Delegation
	link       ucan.Link
}

func (p Proof) Delegation() (Delegation, bool) {
	return p.delegation, p.delegation != nil
}

func (p Proof) Link() ucan.Link {
	if p.delegation != nil {
		return p.delegation.Link()
	}
	return p.link
}

func FromDelegation(delegation Delegation) Proof {
	return Proof{delegation, nil}
}

func FromLink(link ucan.Link) Proof {
	return Proof{nil, link}
}

type Proofs []Proof

func NewProofsView(links []ipld.Link, bs blockstore.BlockReader) Proofs {
	proofs := make(Proofs, 0, len(links))
	for _, link := range links {
		if delegation, err := NewDelegationView(link, bs); err == nil {
			proofs = append(proofs, FromDelegation(delegation))
		} else {
			proofs = append(proofs, FromLink(link))
		}
	}
	return proofs
}

// WriteInto writes a set of proofs, some of which may be full delegations to a blockstore
func (proofs Proofs) WriteInto(bs blockstore.BlockWriter) ([]ipld.Link, error) {
	links := make([]ucan.Link, 0, len(proofs))
	for _, p := range proofs {
		links = append(links, p.Link())
		if delegation, isDelegation := p.Delegation(); isDelegation {
			err := blockstore.WriteInto(delegation, bs)
			if err != nil {
				return nil, err
			}
		}
	}
	return links, nil
}
