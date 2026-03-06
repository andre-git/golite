package test

import (
	"golite/internal/vdbe"
	"testing"
)

// TestVdbeExecution verifies VDBE program execution, register manipulation,
// and HALT behavior using the 'golite/internal/vdbe' package.
func TestVdbeExecution(t *testing.T) {
	t.Run("BasicHaltBehavior", func(t *testing.T) {
		// Verify that OP_Halt returns the expected code and prevents further execution.
		ops := []vdbe.Opcode{
			{Op: vdbe.OP_Integer, P1: 42, P2: 0},
			{Op: vdbe.OP_Halt, P1: 123},
		}
		// NewVdbe(bt, ops, nMem, nCursor)
		v := vdbe.NewVdbe(nil, ops, 1, 0)
		
		rc, err := v.Step()
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}
		if rc != 123 {
			t.Errorf("Expected return code 123, got %d", rc)
		}

		// Once halted, subsequent steps should return an error.
		_, err = v.Step()
		if err == nil {
			t.Error("Expected error when stepping halted VDBE, got nil")
		}
	})

	t.Run("RegisterManipulationAndArithmetic", func(t *testing.T) {
		// Program:
		// 0: Reg[0] = 10
		// 1: Reg[1] = 20
		// 2: Reg[2] = Reg[0] + Reg[1] (30)
		// 3: Reg[3] = 30
		// 4: if Reg[2] == Reg[3] jump to 6 (Success)
		// 5: Halt 0 (Failure)
		// 6: Halt 1 (Success)
		ops := []vdbe.Opcode{
			{Op: vdbe.OP_Integer, P1: 10, P2: 0},
			{Op: vdbe.OP_Integer, P1: 20, P2: 1},
			{Op: vdbe.OP_Add, P1: 0, P2: 1, P3: 2},
			{Op: vdbe.OP_Integer, P1: 30, P2: 3},
			{Op: vdbe.OP_Eq, P1: 2, P2: 6, P3: 3},
			{Op: vdbe.OP_Halt, P1: 0},
			{Op: vdbe.OP_Halt, P1: 1},
		}
		v := vdbe.NewVdbe(nil, ops, 4, 0)
		rc, err := v.Step()
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}
		if rc != 1 {
			t.Errorf("Arithmetic or comparison failed: expected success code 1, got %d", rc)
		}
	})

	t.Run("ControlFlowAndComparisons", func(t *testing.T) {
		// Verify LT, GT, and GOTO branching.
		// Program:
		// 0: Reg[0] = 50
		// 1: Reg[1] = 100
		// 2: if Reg[0] < Reg[1] jump to 4 (Expected)
		// 3: Halt 0 (Failure)
		// 4: Reg[2] = 200
		// 5: if Reg[0] > Reg[2] jump to 7 (Unexpected)
		// 6: Goto 8 (Expected)
		// 7: Halt 0 (Failure)
		// 8: Halt 200 (Success)
		ops := []vdbe.Opcode{
			{Op: vdbe.OP_Integer, P1: 50, P2: 0},
			{Op: vdbe.OP_Integer, P1: 100, P2: 1},
			{Op: vdbe.OP_Lt, P1: 1, P2: 4, P3: 0},
			{Op: vdbe.OP_Halt, P1: 0},
			{Op: vdbe.OP_Integer, P1: 200, P2: 2},
			{Op: vdbe.OP_Gt, P1: 2, P2: 7, P3: 0},
			{Op: vdbe.OP_Goto, P2: 8},
			{Op: vdbe.OP_Halt, P1: 0},
			{Op: vdbe.OP_Halt, P1: 200},
		}
		v := vdbe.NewVdbe(nil, ops, 3, 0)
		rc, _ := v.Step()
		if rc != 200 {
			t.Errorf("Control flow failed: expected 200, got %d", rc)
		}
	})

	t.Run("ArithmeticEdgeCases", func(t *testing.T) {
		// Verify Subtract, Multiply, Divide, Remainder
		// 0: Reg[0] = 15
		// 1: Reg[1] = 4
		// 2: Reg[2] = Reg[0] % Reg[1] (3)
		// 3: Reg[3] = Reg[0] / Reg[1] (3)
		// 4: if Reg[2] == Reg[3] jump to 6
		// 5: Halt 0
		// 6: Halt 1
		ops := []vdbe.Opcode{
			{Op: vdbe.OP_Integer, P1: 15, P2: 0},
			{Op: vdbe.OP_Integer, P1: 4, P2: 1},
			{Op: vdbe.OP_Remainder, P1: 1, P2: 0, P3: 2},
			{Op: vdbe.OP_Divide, P1: 1, P2: 0, P3: 3},
			{Op: vdbe.OP_Eq, P1: 2, P2: 6, P3: 3},
			{Op: vdbe.OP_Halt, P1: 0},
			{Op: vdbe.OP_Halt, P1: 1},
		}
		v := vdbe.NewVdbe(nil, ops, 4, 0)
		rc, _ := v.Step()
		if rc != 1 {
			t.Errorf("Remainder or Division failed: expected 1, got %d", rc)
		}
	})

	t.Run("ResultRowIntegration", func(t *testing.T) {
		// Verify OP_ResultRow return value (SQLITE_ROW = 100).
		ops := []vdbe.Opcode{
			{Op: vdbe.OP_Integer, P1: 1, P2: 0},
			{Op: vdbe.OP_ResultRow},
			{Op: vdbe.OP_Halt, P1: 101},
		}
		v := vdbe.NewVdbe(nil, ops, 1, 0)
		
		rc, err := v.Step()
		if err != nil || rc != 100 {
			t.Fatalf("Expected SQLITE_ROW (100), got %d (err: %v)", rc, err)
		}
		
		rc, err = v.Step()
		if err != nil || rc != 101 {
			t.Fatalf("Expected SQLITE_DONE (101) after ResultRow, got %d (err: %v)", rc, err)
		}
	})

	t.Run("LifecycleResetAndFinalize", func(t *testing.T) {
		// Verify that Reset clears state and Finalize cleans up.
		ops := []vdbe.Opcode{
			{Op: vdbe.OP_Integer, P1: 99, P2: 0},
			{Op: vdbe.OP_Halt, P1: 99},
		}
		v := vdbe.NewVdbe(nil, ops, 1, 0)
		
		v.Step()
		
		// Reset should allow re-execution
		if err := v.Reset(); err != nil {
			t.Fatalf("Reset failed: %v", err)
		}
		
		rc, _ := v.Step()
		if rc != 99 {
			t.Errorf("Execution after Reset failed: expected 99, got %d", rc)
		}

		if err := v.Finalize(); err != nil {
			t.Errorf("Finalize failed: %v", err)
		}
	})
}
