package engine

type OperatorContext struct {
	Dir string
}

type Operator interface {
	Validate() error
	Apply(ctx OperatorContext) error
}
