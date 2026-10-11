package payment

import (
	"context"
)

type ExpireOutput struct {
	Count int64
}

type ExpireUseCase struct {
	deps Deps
}

func NewExpireUseCase(deps Deps) *ExpireUseCase {
	return &ExpireUseCase{deps: deps}
}

func (uc *ExpireUseCase) Execute(ctx context.Context) (*ExpireOutput, error) {
	n, err := uc.deps.Payments.ExpirePending(ctx)
	if err != nil {
		return nil, err
	}
	return &ExpireOutput{Count: n}, nil
}
