package steps

import (
	"context"
	"sync"
)

// Pipeline executes the steps sequentially and the result of each step is passed into the next step
func Pipeline(steps ...Step) Step {
	return &pipelineStep{steps: steps, idx: -1}
}

type pipelineStep struct {
	steps []Step
	idx   int
	m     sync.Mutex
}

func (s *pipelineStep) Run(ctx context.Context) (context.Context, error) {
	s.m.Lock()
	defer s.m.Unlock()

	select {
	case <-ctx.Done():
		//we're done bail
		return ctx, ctx.Err()
	default:
	}

	var err error
	for i, step := range s.steps {
		ctx, err = step.Run(ctx)
		s.idx = i
		if err != nil {
			break
		}
		select {
		case <-ctx.Done():
			//we're done, bail
			return ctx, ctx.Err()
		default:
		}
	}
	return ctx, err
}

func (s *pipelineStep) Rollback(ctx context.Context) (context.Context, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if s.idx < 0 {
		return ctx, nil
	}

	var err error
	for i := s.idx; i >= 0; i-- {
		step := s.steps[i]
		ctx, err = step.Rollback(ctx)
		if err != nil {
			continue
		}
	}
	return ctx, nil
}
