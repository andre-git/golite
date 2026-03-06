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

	default:
		return nil, fmt.Errorf("unsupported statement type for bytecode generation: %T", stmt)
	}

	g.emit(vdbe.Opcode{Op: OP_Halt, P1: 0, P2: 0})

	return g.ops, nil
}

func (g *Generator) emit(op vdbe.Opcode) {
	g.ops = append(g.ops, op)
}
