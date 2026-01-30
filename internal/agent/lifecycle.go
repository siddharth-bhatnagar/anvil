package agent

import (
	"sync"
)

// Phase represents the current phase of the agent lifecycle
type Phase int

const (
	// PhaseUnderstand - gathering information and understanding the request
	PhaseUnderstand Phase = iota
	// PhasePlan - creating a plan of action
	PhasePlan
	// PhaseAct - executing the plan
	PhaseAct
	// PhaseVerify - verifying the results
	PhaseVerify
)

// String returns the string representation of a phase
func (p Phase) String() string {
	switch p {
	case PhaseUnderstand:
		return "Understand"
	case PhasePlan:
		return "Plan"
	case PhaseAct:
		return "Act"
	case PhaseVerify:
		return "Verify"
	default:
		return "Unknown"
	}
}

// Lifecycle manages the Understand → Plan → Act → Verify phases
type Lifecycle struct {
	mu           sync.RWMutex
	currentPhase Phase
	planSteps    []PlanStep
	currentStep  int
}

// PlanStep represents a single step in the execution plan
type PlanStep struct {
	ID          int
	Description string
	Status      StepStatus
	Result      string
	Error       error
}

// StepStatus represents the status of a plan step
type StepStatus int

const (
	StepPending StepStatus = iota
	StepInProgress
	StepCompleted
	StepFailed
)

// String returns the string representation of a step status
func (s StepStatus) String() string {
	switch s {
	case StepPending:
		return "Pending"
	case StepInProgress:
		return "In Progress"
	case StepCompleted:
		return "Completed"
	case StepFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// NewLifecycle creates a new lifecycle manager
func NewLifecycle() *Lifecycle {
	return &Lifecycle{
		currentPhase: PhaseUnderstand,
		planSteps:    make([]PlanStep, 0),
		currentStep:  -1,
	}
}

// CurrentPhase returns the current phase
func (l *Lifecycle) CurrentPhase() Phase {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.currentPhase
}

// SetPhase sets the current phase
func (l *Lifecycle) SetPhase(phase Phase) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.currentPhase = phase
}

// NextPhase advances to the next phase in the lifecycle
func (l *Lifecycle) NextPhase() Phase {
	l.mu.Lock()
	defer l.mu.Unlock()

	switch l.currentPhase {
	case PhaseUnderstand:
		l.currentPhase = PhasePlan
	case PhasePlan:
		l.currentPhase = PhaseAct
	case PhaseAct:
		l.currentPhase = PhaseVerify
	case PhaseVerify:
		// Cycle back to understand for next task
		l.currentPhase = PhaseUnderstand
	}

	return l.currentPhase
}

// SetPlan sets the execution plan
func (l *Lifecycle) SetPlan(steps []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.planSteps = make([]PlanStep, len(steps))
	for i, desc := range steps {
		l.planSteps[i] = PlanStep{
			ID:          i,
			Description: desc,
			Status:      StepPending,
		}
	}
	l.currentStep = -1
}

// GetPlan returns the current plan steps
func (l *Lifecycle) GetPlan() []PlanStep {
	l.mu.RLock()
	defer l.mu.RUnlock()

	steps := make([]PlanStep, len(l.planSteps))
	copy(steps, l.planSteps)
	return steps
}

// StartNextStep marks the next pending step as in progress
func (l *Lifecycle) StartNextStep() *PlanStep {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Find next pending step
	for i := range l.planSteps {
		if l.planSteps[i].Status == StepPending {
			l.planSteps[i].Status = StepInProgress
			l.currentStep = i
			step := l.planSteps[i]
			return &step
		}
	}

	return nil
}

// CompleteCurrentStep marks the current step as completed
func (l *Lifecycle) CompleteCurrentStep(result string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.currentStep >= 0 && l.currentStep < len(l.planSteps) {
		l.planSteps[l.currentStep].Status = StepCompleted
		l.planSteps[l.currentStep].Result = result
	}
}

// FailCurrentStep marks the current step as failed
func (l *Lifecycle) FailCurrentStep(err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.currentStep >= 0 && l.currentStep < len(l.planSteps) {
		l.planSteps[l.currentStep].Status = StepFailed
		l.planSteps[l.currentStep].Error = err
	}
}

// CurrentStep returns the current step being executed
func (l *Lifecycle) CurrentStep() *PlanStep {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.currentStep >= 0 && l.currentStep < len(l.planSteps) {
		step := l.planSteps[l.currentStep]
		return &step
	}

	return nil
}

// AllStepsCompleted returns true if all steps are completed
func (l *Lifecycle) AllStepsCompleted() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, step := range l.planSteps {
		if step.Status != StepCompleted {
			return false
		}
	}

	return len(l.planSteps) > 0
}

// HasFailedSteps returns true if any step has failed
func (l *Lifecycle) HasFailedSteps() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, step := range l.planSteps {
		if step.Status == StepFailed {
			return true
		}
	}

	return false
}

// Reset resets the lifecycle to the initial state
func (l *Lifecycle) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.currentPhase = PhaseUnderstand
	l.planSteps = make([]PlanStep, 0)
	l.currentStep = -1
}

// Progress returns the completion percentage (0-100)
func (l *Lifecycle) Progress() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.planSteps) == 0 {
		return 0
	}

	completed := 0
	for _, step := range l.planSteps {
		if step.Status == StepCompleted {
			completed++
		}
	}

	return (completed * 100) / len(l.planSteps)
}
