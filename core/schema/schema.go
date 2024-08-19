package schema

import (
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/storacha-network/go-ucanto/core/ipld"
	"github.com/storacha-network/go-ucanto/core/result"
	"github.com/storacha-network/go-ucanto/core/result/failure"
)

type Reader[I, O any] interface {
	Read(input I) result.Result[O, failure.Failure]
}

type reader[I, O any] struct {
	readFunc func(input I) result.Result[O, failure.Failure]
}

func (r reader[I, O]) Read(input I) result.Result[O, failure.Failure] {
	return r.readFunc(input)
}

type schemaerr struct {
	message string
}

func (se schemaerr) Name() string {
	return "SchemaError"
}

func (se schemaerr) Error() string {
	return se.message
}

func (se schemaerr) Build() (ipld.Node, error) {
	np := basicnode.Prototype.Any
	nb := np.NewBuilder()
	ma, err := nb.BeginMap(2)
	if err != nil {
		return nil, err
	}
	ma.AssembleKey().AssignString("name")
	ma.AssembleValue().AssignString(se.Name())
	ma.AssembleKey().AssignString("message")
	ma.AssembleValue().AssignString(se.Error())
	ma.Finish()
	return nb.Build(), nil
}

func NewSchemaError(message string) failure.Failure {
	return schemaerr{message}
}
