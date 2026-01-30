package agent

import (
	"testing"
)

func TestNewLifecycle(t *testing.T) {
	lc := NewLifecycle()

	if lc == nil {
		t.Fatal("NewLifecycle returned nil")
	}

	if lc.CurrentPhase() != PhaseUnderstand {
		t.Errorf("Expected initial phase to be Understand, got %v", lc.CurrentPhase())
	}
}

func TestLifecyclePhaseTransitions(t *testing.T) {
	lc := NewLifecycle()

	// Test phase progression
	phases := []Phase{PhaseUnderstand, PhasePlan, PhaseAct, PhaseVerify}

	for i, expected := range phases {
		if lc.CurrentPhase() != expected {
			t.Errorf("Step %d: Expected phase %v, got %v", i, expected, lc.CurrentPhase())
		}
		lc.NextPhase()
	}

	// After Verify, should cycle back to Understand
	if lc.CurrentPhase() != PhaseUnderstand {
		t.Errorf("Expected phase to cycle back to Understand, got %v", lc.CurrentPhase())
	}
}

func TestLifecycleSetPhase(t *testing.T) {
	lc := NewLifecycle()

	lc.SetPhase(PhaseAct)

	if lc.CurrentPhase() != PhaseAct {
		t.Errorf("Expected PhaseAct, got %v", lc.CurrentPhase())
	}
}

func TestLifecycleSetPlan(t *testing.T) {
	lc := NewLifecycle()

	steps := []string{"Step 1", "Step 2", "Step 3"}
	lc.SetPlan(steps)

	plan := lc.GetPlan()

	if len(plan) != 3 {
		t.Fatalf("Expected 3 steps, got %d", len(plan))
	}

	for i, step := range plan {
		if step.Description != steps[i] {
			t.Errorf("Step %d: Expected '%s', got '%s'", i, steps[i], step.Description)
		}
		if step.Status != StepPending {
			t.Errorf("Step %d: Expected status Pending, got %v", i, step.Status)
		}
	}
}

func TestLifecycleStepExecution(t *testing.T) {
	lc := NewLifecycle()

	lc.SetPlan([]string{"First", "Second", "Third"})

	// Start first step
	step := lc.StartNextStep()
	if step == nil {
		t.Fatal("StartNextStep returned nil")
	}
	if step.Description != "First" {
		t.Errorf("Expected 'First', got '%s'", step.Description)
	}
	if step.Status != StepInProgress {
		t.Errorf("Expected StepInProgress, got %v", step.Status)
	}

	// Complete first step
	lc.CompleteCurrentStep("Done with first")

	plan := lc.GetPlan()
	if plan[0].Status != StepCompleted {
		t.Error("First step should be completed")
	}
	if plan[0].Result != "Done with first" {
		t.Errorf("Expected result 'Done with first', got '%s'", plan[0].Result)
	}

	// Start second step
	step = lc.StartNextStep()
	if step.Description != "Second" {
		t.Errorf("Expected 'Second', got '%s'", step.Description)
	}
}

func TestLifecycleFailStep(t *testing.T) {
	lc := NewLifecycle()

	lc.SetPlan([]string{"Will fail"})
	lc.StartNextStep()
	lc.FailCurrentStep(nil)

	plan := lc.GetPlan()
	if plan[0].Status != StepFailed {
		t.Error("Step should be failed")
	}

	if !lc.HasFailedSteps() {
		t.Error("HasFailedSteps should return true")
	}
}

func TestLifecycleAllStepsCompleted(t *testing.T) {
	lc := NewLifecycle()

	lc.SetPlan([]string{"One", "Two"})

	if lc.AllStepsCompleted() {
		t.Error("AllStepsCompleted should be false initially")
	}

	lc.StartNextStep()
	lc.CompleteCurrentStep("")
	lc.StartNextStep()
	lc.CompleteCurrentStep("")

	if !lc.AllStepsCompleted() {
		t.Error("AllStepsCompleted should be true after all steps completed")
	}
}

func TestLifecycleProgress(t *testing.T) {
	lc := NewLifecycle()

	lc.SetPlan([]string{"A", "B", "C", "D"})

	if lc.Progress() != 0 {
		t.Errorf("Expected 0%% progress, got %d%%", lc.Progress())
	}

	lc.StartNextStep()
	lc.CompleteCurrentStep("")

	if lc.Progress() != 25 {
		t.Errorf("Expected 25%% progress, got %d%%", lc.Progress())
	}

	lc.StartNextStep()
	lc.CompleteCurrentStep("")

	if lc.Progress() != 50 {
		t.Errorf("Expected 50%% progress, got %d%%", lc.Progress())
	}
}

func TestLifecycleReset(t *testing.T) {
	lc := NewLifecycle()

	lc.SetPhase(PhaseAct)
	lc.SetPlan([]string{"Step"})
	lc.StartNextStep()

	lc.Reset()

	if lc.CurrentPhase() != PhaseUnderstand {
		t.Error("Phase should be reset to Understand")
	}

	if len(lc.GetPlan()) != 0 {
		t.Error("Plan should be empty after reset")
	}
}

func TestPhaseString(t *testing.T) {
	tests := []struct {
		phase    Phase
		expected string
	}{
		{PhaseUnderstand, "Understand"},
		{PhasePlan, "Plan"},
		{PhaseAct, "Act"},
		{PhaseVerify, "Verify"},
	}

	for _, tt := range tests {
		if tt.phase.String() != tt.expected {
			t.Errorf("Phase %d: Expected '%s', got '%s'", tt.phase, tt.expected, tt.phase.String())
		}
	}
}

func TestStepStatusString(t *testing.T) {
	tests := []struct {
		status   StepStatus
		expected string
	}{
		{StepPending, "Pending"},
		{StepInProgress, "In Progress"},
		{StepCompleted, "Completed"},
		{StepFailed, "Failed"},
	}

	for _, tt := range tests {
		if tt.status.String() != tt.expected {
			t.Errorf("Status %d: Expected '%s', got '%s'", tt.status, tt.expected, tt.status.String())
		}
	}
}
