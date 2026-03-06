package sql

import (
	"fmt"
	"golite/internal/schema"
	"golite/internal/vdbe"
	"strconv"
)

const (
	OP_Halt        = vdbe.OP_Halt
	OP_Transaction = vdbe.OP_Transaction
	OP_CreateTable = vdbe.OP_CreateTable
)

type Generator struct {
	ops          []vdbe.Opcode
	schema       *schema.Schema
	currentTable *schema.Table 
	cursors      map[string]int 
}

func NewGenerator(s *schema.Schema) *Generator {
	return &Generator{
		ops:    make([]vdbe.Opcode, 0),
		schema: s,
		cursors: make(map[string]int),
	}
}

func (g *Generator) Generate(stmt Stmt) ([]vdbe.Opcode, error) {
	g.ops = nil
	g.currentTable = nil
	g.cursors = make(map[string]int)

	switch s := stmt.(type) {
	case *BeginStmt:
		g.emit(vdbe.Opcode{Op: OP_Transaction, P1: 0, P2: 1})

	case *CommitStmt:
		g.emit(vdbe.Opcode{Op: OP_Transaction, P1: 0, P2: 0})

	case *RollbackStmt:
		g.emit(vdbe.Opcode{Op: OP_Transaction, P1: 0, P2: 2})

	case *CreateTableStmt:
		g.emit(vdbe.Opcode{Op: OP_Transaction, P1: 0, P2: 1})
		g.emit(vdbe.Opcode{Op: OP_CreateTable, P1: 0, P2: 1, P4: s.Name})

	case *ExplainStmt:
		return g.Generate(s.Stmt)

	case *SelectStmt:
		return g.generateSelect(s)

	case *InsertStmt:
		table, ok := g.schema.Tables[s.Table]
		if !ok {
			return nil, fmt.Errorf("no such table: %s", s.Table)
		}
		g.currentTable = table

		// Verify column count
		for _, row := range s.Values {
			expectedCount := len(table.Columns)
			if len(s.Columns) > 0 {
				expectedCount = len(s.Columns)
			}
			if len(row) != expectedCount {
				return nil, fmt.Errorf("table %s has %d columns but %d values were supplied", s.Table, expectedCount, len(row))
			}
		}

		g.emit(vdbe.Opcode{Op: OP_Transaction, P1: 0, P2: 1})
		g.emit(vdbe.Opcode{Op: vdbe.OP_OpenWrite, P1: 0, P2: int(table.RootPgno)})

		for _, row := range s.Values {
			regBase := 1
			for i, expr := range row {
				g.generateExpr(expr, regBase+i)
			}

			rowidReg := regBase + len(row)
			recordReg := rowidReg + 1

			g.emit(vdbe.Opcode{Op: vdbe.OP_NewRowid, P1: 0, P2: rowidReg})
			g.emit(vdbe.Opcode{Op: vdbe.OP_MakeRecord, P1: regBase, P2: len(row), P3: recordReg})
			g.emit(vdbe.Opcode{Op: vdbe.OP_Insert, P1: 0, P2: recordReg, P3: rowidReg})
		}

	case *UpdateStmt:
		table, ok := g.schema.Tables[s.Table]
		if !ok {
			return nil, fmt.Errorf("no such table: %s", s.Table)
		}
		g.currentTable = table
		g.emit(vdbe.Opcode{Op: vdbe.OP_Noop})

	case *DeleteStmt:
		table, ok := g.schema.Tables[s.Table]
		if !ok {
			return nil, fmt.Errorf("no such table: %s", s.Table)
		}
		g.currentTable = table
		g.emit(vdbe.Opcode{Op: vdbe.OP_Noop})

	default:
		return nil, fmt.Errorf("unsupported statement type for bytecode generation: %T", stmt)
	}

	g.emit(vdbe.Opcode{Op: OP_Halt, P1: 0, P2: 0})

	return g.ops, nil
}

func (g *Generator) generateSelect(s *SelectStmt) ([]vdbe.Opcode, error) {
	if s.From == nil || len(s.From.Items) == 0 {
		g.generateExpr(s.Columns[0], 1)
		g.emit(vdbe.Opcode{Op: vdbe.OP_ResultRow, P1: 1, P2: len(s.Columns)})
		g.emit(vdbe.Opcode{Op: vdbe.OP_Halt, P1: 0, P2: 0})
		return g.ops, nil
	}

	type loopInfo struct {
		cursor     int
		rewindAddr int
		nextAddr   int
	}
	loops := make([]loopInfo, len(s.From.Items))

	for i, item := range s.From.Items {
		cursorIdx := i
		g.cursors[item.Name] = cursorIdx
		if item.Alias != "" {
			g.cursors[item.Alias] = cursorIdx
		}

		table, ok := g.schema.Tables[item.Name]
		if !ok {
			return nil, fmt.Errorf("no such table: %s", item.Name)
		}
		g.currentTable = table

		g.emit(vdbe.Opcode{Op: vdbe.OP_OpenRead, P1: cursorIdx, P2: int(table.RootPgno)})
		loops[i].cursor = cursorIdx
		loops[i].rewindAddr = len(g.ops)
		g.emit(vdbe.Opcode{Op: vdbe.OP_Rewind, P1: cursorIdx, P2: 0}) 
		loops[i].nextAddr = len(g.ops)
	}

	// 1) Evaluate WHERE simple filter
	if s.Where != nil {
		if bin, ok := s.Where.(*BinaryExpr); ok && bin.Op.Type == TK_EQ {
			g.generateExpr(bin.Left, 1)
			g.generateExpr(bin.Right, 2)
			// Jump to NEXT if NOT EQUAL
			g.emit(vdbe.Opcode{Op: vdbe.OP_Ne, P1: 1, P2: 0, P3: 2}) 
		} else {
			g.generateExpr(s.Where, 1)
		}
	}

	// 2) Columns
	isStar := false
	if len(s.Columns) == 1 {
		if lit, ok := s.Columns[0].(*LiteralExpr); ok && lit.Token.Value == "*" {
			isStar = true
		}
	}

	if isStar {
		for i := range g.currentTable.Columns {
			g.emit(vdbe.Opcode{Op: vdbe.OP_Column, P1: loops[0].cursor, P2: i, P3: 10 + i})
		}
		g.emit(vdbe.Opcode{Op: vdbe.OP_ResultRow, P1: 10, P2: len(g.currentTable.Columns)})
	} else {
		for i, colExpr := range s.Columns {
			g.generateExpr(colExpr, 10+i)
		}
		g.emit(vdbe.Opcode{Op: vdbe.OP_ResultRow, P1: 10, P2: len(s.Columns)})
	}

	// 3) Emit OP_Next, and patch addresses
	for i := len(loops) - 1; i >= 0; i-- {
		nextInstrAddr := len(g.ops)
		g.emit(vdbe.Opcode{Op: vdbe.OP_Next, P1: loops[i].cursor, P2: loops[i].nextAddr})
		
		// Patch OP_Rewind to jump past OP_Next when table is empty
		g.ops[loops[i].rewindAddr].P2 = len(g.ops)

		// Patch OP_Ne jump to jump to Next
		if s.Where != nil {
			for j := loops[i].nextAddr; j < nextInstrAddr; j++ {
				if g.ops[j].Op == vdbe.OP_Ne && g.ops[j].P2 == 0 {
					g.ops[j].P2 = nextInstrAddr 
				}
			}
		}
	}

	return g.ops, nil
}

func (g *Generator) generateExpr(expr Expr, reg int) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *LiteralExpr:
		if g.currentTable != nil && e.Token.Type == TK_ID {
			if e.Token.Value == "*" {
				return
			}
			for i, col := range g.currentTable.Columns {
				if col.Name == e.Token.Value {
					g.emit(vdbe.Opcode{Op: vdbe.OP_Column, P1: 0, P2: i, P3: reg})
					return
				}
			}
		}

		if e.Token.Type == TK_INTEGER {
			val, _ := strconv.Atoi(e.Token.Value)
			g.emit(vdbe.Opcode{Op: vdbe.OP_Integer, P1: val, P2: reg})
		} else {
			g.emit(vdbe.Opcode{Op: vdbe.OP_String, P2: reg, P4: e.Token.Value})
		}

	case *BinaryExpr:
		g.generateExpr(e.Left, reg)
		g.generateExpr(e.Right, reg+1)
		switch e.Op.Type {
		case TK_PLUS:
			g.emit(vdbe.Opcode{Op: vdbe.OP_Add, P1: reg, P2: reg + 1, P3: reg})
		case TK_MINUS:
			g.emit(vdbe.Opcode{Op: vdbe.OP_Subtract, P1: reg, P2: reg + 1, P3: reg})
		case TK_STAR:
			g.emit(vdbe.Opcode{Op: vdbe.OP_Multiply, P1: reg, P2: reg + 1, P3: reg})
		case TK_SLASH:
			g.emit(vdbe.Opcode{Op: vdbe.OP_Divide, P1: reg, P2: reg + 1, P3: reg})
		case TK_EQ:
			g.emit(vdbe.Opcode{Op: vdbe.OP_Eq, P1: reg, P2: 0, P3: reg + 1}) 
		}
	}
}

func (g *Generator) emit(op vdbe.Opcode) {
	g.ops = append(g.ops, op)
}
