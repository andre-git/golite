package golite

import (
	"errors"
	"golite/internal/btree"
	"golite/internal/schema"
	"golite/internal/sql"
	"golite/internal/vdbe"
)

type DB struct {
    backends   []*Backend // main, temp, attached
    errCode    int
    autoCommit bool
}

type Backend struct {
    Name   string
    Btree  btree.Btree
    Schema *schema.Schema
}

type Stmt struct {
    db   *DB
    vdbe vdbe.Vdbe
}

func (db *DB) Close() error {
	for _, b := range db.backends {
		if b.Btree != nil {
		}
	}
	return nil
}

func (db *DB) Prepare(sqlStr string) (*Stmt, error) {
	if db == nil {
		return nil, errors.New("nil database connection")
	}
	
	lexer := sql.NewLexer(sqlStr)
	tokens := lexer.Tokenize()

	parser := sql.NewParser(tokens)
	cmdList, errs := parser.Parse()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if len(cmdList.Statements) == 0 {
		return nil, errors.New("no statement found")
	}

	generator := sql.NewGenerator(db.backends[0].Schema)
	ops, err := generator.Generate(cmdList.Statements[0])
	if err != nil {
		return nil, err
	}

	stmt := &Stmt{
		db:   db,
		vdbe: vdbe.NewVdbe(db.backends[0].Btree, ops, 100, 100), // Increase capacity
	}

	// For CREATE TABLE, we need to update the in-memory schema after execution
	if ct, ok := cmdList.Statements[0].(*sql.CreateTableStmt); ok {
		table := &schema.Table{
			Name:     ct.Name,
			RootPgno: 2, // Matches first Allocate() after Pager init
		}
		for _, c := range ct.Columns {
			table.Columns = append(table.Columns, &schema.Column{Name: c.Name})
		}
		db.backends[0].Schema.Tables[ct.Name] = table
	}

	return stmt, nil
}

func (db *DB) Exec(sql string) error {
	stmt, err := db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Finalize()

	for {
		row, err := stmt.Step()
		if err != nil {
			return err
		}
		if !row {
			break
		}
	}
	return nil
}

func (s *Stmt) Step() (bool, error) {
	if s.vdbe == nil {
		return false, errors.New("nil vdbe")
	}
	rc, err := s.vdbe.Step()
	if err != nil {
		return false, err
	}
	return rc == 100, nil
}

func (s *Stmt) Finalize() error {
	if s.vdbe != nil {
		s.vdbe.Finalize()
	}
	return nil
}

func (db *DB) ErrCode() int {
	return db.errCode
}
