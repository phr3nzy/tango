package tango

import (
	"fmt"
)

// ExecutionStrategy defines the interface for different execution strategies.
type ExecutionStrategy[Services, State any] interface {
	Execute(m *Machine[Services, State]) (*Response[Services, State], error)
	Compensate(m *Machine[Services, State]) (*Response[Services, State], error)
}

// SequentialStrategy is a default implementation of ExecutionStrategy that runs steps sequentially.
type SequentialStrategy[Services, State any] struct{}

func (s *SequentialStrategy[Services, State]) Execute(m *Machine[Services, State]) (*Response[Services, State], error) {
	for i := 0; i < len(m.Steps); i++ {
		step := m.Steps[i]

		response, err := m.executeStep(step)
		if err != nil {
			return nil, err
		}

		m.mu.Lock()
		m.ExecutedSteps = append(m.ExecutedSteps, step)
		m.Context.PreviousResult = response
		m.mu.Unlock()

		switch response.Status {
		case NEXT:
			continue
		case DONE:
			return response, nil
		case ERROR:
			cResponse, err := m.Compensate()
			if err != nil {
				return nil, fmt.Errorf("compensate error: %v", err)
			}
			return cResponse, fmt.Errorf("step %s failed: %v", step.Name, response.Result)
		case SKIP:
			i += response.SkipCount
		case JUMP:
			targetIndex := -1
			for index, s := range m.Steps {
				if s.Name == response.JumpTarget {
					targetIndex = index
					break
				}
			}
			if targetIndex >= 0 {
				i = targetIndex - 1
			} else {
				return nil, fmt.Errorf("jump target '%s' not found at %s", response.JumpTarget, step.Name)
			}
		}
	}

	return nil, nil
}

// Compensate runs the compensate functions of the executed steps.
func (s *SequentialStrategy[Services, State]) Compensate(m *Machine[Services, State]) (*Response[Services, State], error) {
	m.Context = m.InitialContext
	for i := len(m.ExecutedSteps) - 1; i >= 0; i-- {
		step := m.ExecutedSteps[i]
		if step.BeforeCompensate != nil {
			if err := step.BeforeCompensate(m.Context); err != nil {
				return nil, err
			}
		}
		if step.Compensate == nil {
			return nil, fmt.Errorf("step %s has no compensate function", step.Name)
		}
		if _, err := step.Compensate(m.Context); err != nil {
			return nil, err
		}
		if step.AfterCompensate != nil {
			if err := step.AfterCompensate(m.Context); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

// ConcurrentStrategy runs steps concurrently.
type ConcurrentStrategy[Services, State any] struct {
	Concurrency int
}

func (c *ConcurrentStrategy[Services, State]) Execute(m *Machine[Services, State]) (*Response[Services, State], error) {
	if c.Concurrency <= 1 {
		return (&SequentialStrategy[Services, State]{}).Execute(m)
	}

	sem := make(chan struct{}, c.Concurrency)
	responseChan := make(chan *Response[Services, State], len(m.Steps))
	errorChan := make(chan error, len(m.Steps))

	for i := 0; i < len(m.Steps); i++ {
		sem <- struct{}{}
		go func(step Step[Services, State]) {
			defer func() { <-sem }()
			response, err := m.executeStep(step)
			if err != nil {
				errorChan <- err
				return
			}
			responseChan <- response
			m.mu.Lock()
			m.ExecutedSteps = append(m.ExecutedSteps, step)
			m.Context.PreviousResult = response
			m.mu.Unlock()
		}(m.Steps[i])
	}

	for i := 0; i < c.Concurrency; i++ {
		sem <- struct{}{}
	}

	close(responseChan)
	close(errorChan)

	select {
	case <-errorChan:
		cResponse, err := m.Compensate()
		if err != nil {
			return nil, fmt.Errorf("compensate error: %v", err)
		}
		return cResponse, err
	default:
	}

	for response := range responseChan {
		if response.Status == DONE {
			return response, nil
		}
	}

	return nil, nil
}

// Compensate runs the compensate functions of the executed steps.
func (c *ConcurrentStrategy[Services, State]) Compensate(m *Machine[Services, State]) (*Response[Services, State], error) {
	if c.Concurrency <= 1 {
		return (&SequentialStrategy[Services, State]{}).Compensate(m)
	}

	sem := make(chan struct{}, c.Concurrency)
	errorChan := make(chan error, len(m.ExecutedSteps))

	for i := len(m.ExecutedSteps) - 1; i >= 0; i-- {
		sem <- struct{}{}
		go func(step Step[Services, State]) {
			defer func() { <-sem }()

			if step.BeforeCompensate != nil {
				if err := step.BeforeCompensate(m.Context); err != nil {
					errorChan <- err
					return
				}
			}
			if step.Compensate == nil {
				errorChan <- fmt.Errorf("step %s has no compensate function", step.Name)
				return
			}
			if _, err := step.Compensate(m.Context); err != nil {
				errorChan <- err
				return
			}
			if step.AfterCompensate != nil {
				if err := step.AfterCompensate(m.Context); err != nil {
					errorChan <- err
					return
				}
			}
		}(m.ExecutedSteps[i])
	}

	for i := 0; i < c.Concurrency; i++ {
		sem <- struct{}{}
	}

	close(errorChan)

	select {
	case err := <-errorChan:
		return nil, err
	default:
		return nil, nil
	}
}
