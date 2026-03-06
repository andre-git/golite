package sql

import (
	"fmt"
	"golite/internal/vdbe"
)

const (
	OP_Halt        = vdbe.OP_Halt
	OP_Transaction = vdbe.OP_Transaction
	OP_CreateTable = vdbe.OP_CreateTable
)

type Generator struct {
	ops []vdbe.Opcode
}

func NewGenerator() *Generator {
	return &Generator{
		ops: make([]vdbe.Opcode, 0),
	}
}

func (g *Generator) Generate(stmt Stmt) ([]vdbe.Opcode, error) {
	g.ops = nil 

	switch s := stmt.(type) {
	case *BeginStmt:
		g.emit(vdbe.Opcode{Op: OP_Transaction, P1: 0, P2: 1})
	
	case *CommitStmt:
		g.emit(vdbe.Opcode{Op: OP_Transaction, P1: 0, P2: 0})
	
	case *RollbackStmt:
		g.emit(vdbe.Opcode{Op: OP_Transaction, P1: 0, P2: 2})

	case *CreateTableStmt:
		g.emit(vdbe.Opcode{Op: OP_Transaction, P1: 0, P2: 1})
		g.emit(vdbe.Opcode{Op: OP_CreateTable, P1: 0, P2: 0, P4: s.Name})

	case *ExplainStmt:
		return g.Generate(s.Stmt)

	case *SelectStmt:
		// Basic select placeholder
		g.emit(vdbe.Opcode{Op: vdbe.OP_Noop})

	case *InsertStmt:
		g.emit(vdbe.Opcode{Op: OP_Transaction, P1: 0, P2: 1})
		// Cursor 0 for the table, P2 placeholder for root page number
		g.emit(vdbe.Opcode{Op: vdbe.OP_OpenWrite, P1: 0, P2: 2})

		for _, row := range s.Values {
			regBase := 1
			for i, expr := range row {
				g.generateExpr(expr, regBase+i)
			}

			rowidReg := regBase + len(row)
			recordReg := rowidReg + 1

			// OP_NewRowid: P1=cursor, P2=destination register
			g.emit(vdbe.Opcode{Op: vdbe.OP_NewRowid, P1: 0, P2: rowidReg})
			// OP_MakeRecord: P1=first reg, P2=count, P3=destination record register
			g.emit(vdbe.Opcode{Op: vdbe.OP_MakeRecord, P1: regBase, P2: len(row), P3: recordReg})
			// OP_Insert: P1=cursor, P2=record register, P3=rowid register
			g.emit(vdbe.Opcode{Op: vdbe.OP_Insert, P1: 0, P2: recordReg, P3: rowidReg})
		}

	default:
		return nil, fmt.Errorf("unsupported statement type for bytecode generation: %T", stmt)
	}

	g.emit(vdbe.Opcode{Op: OP_Halt, P1: 0, P2: 0})

	return g.ops, nil
}

func (g *Generator) generateExpr(expr Expr, reg int) {
	switch e := expr.(type) {
	case *LiteralExpr:
		// Currently supports string/id literals; should be expanded for other types
		g.emit(vdbe.Opcode{Op: vdbe.OP_String, P2: reg, P4: e.Token.Value})
	}
}

func (g *Generator) emit(op vdbe.Opcode) {
	g.ops = append(g.ops, op)
}
