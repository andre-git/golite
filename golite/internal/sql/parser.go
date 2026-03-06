package sql

import (
	"fmt"
)

type TokenType int

const (
	TK_ILLEGAL TokenType = iota
	TK_EOF
	TK_SEMI
	TK_EXPLAIN
	TK_QUERY
	TK_PLAN
	TK_BEGIN
	TK_TRANSACTION
	TK_DEFERRED
	TK_IMMEDIATE
	TK_EXCLUSIVE
	TK_COMMIT
	TK_END
	TK_ROLLBACK
	TK_SAVEPOINT
	TK_RELEASE
	TK_TO
	TK_CREATE
	TK_TABLE
	TK_TEMP
	TK_IF
	TK_NOT
	TK_EXISTS
	TK_ID
	TK_STRING
	TK_SELECT
	TK_FROM
	TK_WHERE
	TK_GROUP
	TK_BY
	TK_HAVING
	TK_ORDER
	TK_LIMIT
	TK_INSERT
	TK_INTO
	TK_VALUES
	TK_UPDATE
	TK_SET
	TK_DELETE
	TK_DISTINCT
	TK_ALL
	TK_AS
	TK_LP
	TK_RP
	TK_COMMA
	TK_EQ
	TK_PLUS
)

type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

type ParseError struct {
	Message string
	Token   Token
}

func (e *ParseError) Error() string {
	if e.Token.Value != "" {
		return fmt.Sprintf("near %q: %s", e.Token.Value, e.Message)
	}
	return e.Message
}

type Node interface {
	node()
}

type Stmt interface {
	Node
	stmt()
}

type Expr interface {
	Node
	expr()
}

type LiteralExpr struct { Token Token }
func (*LiteralExpr) node() {}
func (*LiteralExpr) expr() {}

type BinaryExpr struct {
    Left  Expr
    Op    Token
    Right Expr
}
func (*BinaryExpr) node() {}
func (*BinaryExpr) expr() {}

type CmdList struct {
	Statements []Stmt
}

func (CmdList) node() {}

type ExplainStmt struct {
	QueryPlan bool
	Stmt      Stmt
}

func (ExplainStmt) node() {}
func (ExplainStmt) stmt() {}

type TransactionType int

const (
	TransDeferred TransactionType = iota
	TransImmediate
	TransExclusive
)

type BeginStmt struct {
	Type TransactionType
}

func (BeginStmt) node() {}
func (BeginStmt) stmt() {}

type CommitStmt struct{}

func (CommitStmt) node() {}
func (CommitStmt) stmt() {}

type RollbackStmt struct {
	SavepointName string
}

func (RollbackStmt) node() {}
func (RollbackStmt) stmt() {}

type SavepointStmt struct {
	Name string
}

func (SavepointStmt) node() {}
func (SavepointStmt) stmt() {}

type ReleaseStmt struct {
	Name string
}

func (ReleaseStmt) node() {}
func (ReleaseStmt) stmt() {}

type CreateTableStmt struct {
	Temp         bool
	IfNotExists  bool
	Name         string
	DBName       string
	Columns      []*ColumnDef
	Constraints  []TableConstraint
	WithoutRowid bool
	Strict       bool
	Select       *SelectStmt 
}

func (CreateTableStmt) node() {}
func (CreateTableStmt) stmt() {}

type ColumnDef struct {
	Name        string
	Type        *ColumnType
	Constraints []ColumnConstraint
}

type ColumnType struct {
	Name string
}

type ColumnConstraint interface {
	columnConstraint()
}

type TableConstraint interface {
	tableConstraint()
}

type SelectStmt struct {
	Distinct bool
	Columns  []Expr 
	From     *SrcList
	Where    Expr
	GroupBy  []Expr
	Having   Expr
	OrderBy  []OrderingTerm
	Limit    Expr
}

func (SelectStmt) node() {}
func (SelectStmt) stmt() {}

type OrderingTerm struct {
	X    Expr
	Desc bool
}

type SrcList struct {
	Items []SrcItem
}

type SrcItem struct {
	Name  string
	Alias string
}

type InsertStmt struct {
	Table   string
	Columns []string
	Values  [][]Expr
	Select  *SelectStmt
}

func (InsertStmt) node() {}
func (InsertStmt) stmt() {}

type UpdateStmt struct {
	Table string
	Sets  []UpdateSet
	Where Expr
}

func (UpdateStmt) node() {}
func (UpdateStmt) stmt() {}

type UpdateSet struct {
	Column string
	Value  Expr
}

type DeleteStmt struct {
	Table string
	Where Expr
}

func (DeleteStmt) node() {}
func (DeleteStmt) stmt() {}

type Parser struct {
	tokens []Token
	pos    int
	errors []error
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

func (p *Parser) Parse() (*CmdList, []error) {
	var stmts []Stmt

	for !p.isAtEnd() {
		stmt := p.parseCmd()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
		
		if !p.match(TK_SEMI) && !p.isAtEnd() {
			p.syntaxError(p.peek(), "syntax error")
			p.synchronize() 
		}
	}

	if len(p.errors) > 0 {
		return nil, p.errors
	}

	return &CmdList{Statements: stmts}, nil
}

func (p *Parser) parseCmd() Stmt {
	switch p.peek().Type {
	case TK_SELECT:
		return p.parseSelect()
	case TK_INSERT:
		return p.parseInsert()
	case TK_UPDATE:
		return p.parseUpdate()
	case TK_DELETE:
		return p.parseDelete()
	case TK_BEGIN:
		return p.parseBegin()
	case TK_COMMIT, TK_END:
		return p.parseCommit()
	case TK_CREATE:
		return p.parseCreateTable()
	default:
		p.syntaxError(p.peek(), "syntax error")
		return nil
	}
}

func (p *Parser) parseCreateTable() *CreateTableStmt {
	p.match(TK_CREATE)
	stmt := &CreateTableStmt{}

	if p.match(TK_TEMP) {
		stmt.Temp = true
	}

	if !p.match(TK_TABLE) {
		p.syntaxError(p.peek(), "expected TABLE")
		return nil
	}

	if p.match(TK_IF) {
		if !p.match(TK_NOT) {
			p.syntaxError(p.peek(), "expected NOT")
			return nil
		}
		if !p.match(TK_EXISTS) {
			p.syntaxError(p.peek(), "expected EXISTS")
			return nil
		}
		stmt.IfNotExists = true
	}

	if p.peek().Type != TK_ID {
		p.syntaxError(p.peek(), "expected table name")
		return nil
	}
	stmt.Name = p.peek().Value
	p.pos++

	if !p.match(TK_LP) {
		p.syntaxError(p.peek(), "expected '('")
		return nil
	}

	for {
		col := p.parseColumnDef()
		if col == nil {
			return nil
		}
		stmt.Columns = append(stmt.Columns, col)

		if !p.match(TK_COMMA) {
			break
		}
	}

	if !p.match(TK_RP) {
		p.syntaxError(p.peek(), "expected ')'")
		return nil
	}

	return stmt
}

func (p *Parser) parseColumnDef() *ColumnDef {
	if p.peek().Type != TK_ID {
		p.syntaxError(p.peek(), "expected column name")
		return nil
	}
	name := p.peek().Value
	p.pos++

	col := &ColumnDef{Name: name}

	if p.peek().Type == TK_ID {
		col.Type = &ColumnType{Name: p.peek().Value}
		p.pos++
	}

	return col
}

func (p *Parser) parseExpr() Expr {
	return p.parseBinaryExpr(0)
}

func (p *Parser) parseBinaryExpr(minPrecedence int) Expr {
	left := p.parsePrimaryExpr()
	if left == nil {
		return nil
	}
	for {
		op := p.peek()
		prec := p.getPrecedence(op.Type)
		if prec < minPrecedence || prec == 0 {
			break
		}
		p.pos++
		right := p.parseBinaryExpr(prec + 1)
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parsePrimaryExpr() Expr {
	tok := p.peek()
	if tok.Type == TK_ID || tok.Type == TK_STRING {
		p.pos++
		return &LiteralExpr{Token: tok}
	}
	if p.match(TK_LP) {
		expr := p.parseExpr()
		p.match(TK_RP)
		return expr
	}
	return nil
}

func (p *Parser) getPrecedence(t TokenType) int {
	switch t {
	case TK_PLUS:
		return 10
	}
	return 0
}

func (p *Parser) parseBegin() *BeginStmt {
	p.match(TK_BEGIN)
	p.match(TK_TRANSACTION) // Optional
	return &BeginStmt{Type: TransDeferred}
}

func (p *Parser) parseCommit() *CommitStmt {
	p.match(TK_COMMIT, TK_END)
	p.match(TK_TRANSACTION) // Optional
	return &CommitStmt{}
}

func (p *Parser) parseSelect() *SelectStmt {
	p.match(TK_SELECT)
	stmt := &SelectStmt{}
	if p.match(TK_DISTINCT) {
		stmt.Distinct = true
	} else {
		p.match(TK_ALL)
	}

	for {
		if p.peek().Value == "*" {
			stmt.Columns = append(stmt.Columns, &LiteralExpr{Token: Token{Type: TK_ID, Value: "*"}})
			p.pos++
		} else {
			expr := p.parseExpr()
			if expr == nil {
				break
			}
			stmt.Columns = append(stmt.Columns, expr)
		}

		if !p.match(TK_COMMA) {
			break
		}
	}

	if p.match(TK_FROM) {
		// Just consume for now: IDs, commas, and AS
		for !p.isAtEnd() && (p.peek().Type == TK_ID || p.peek().Type == TK_COMMA || p.peek().Type == TK_AS) {
			p.pos++
		}
	}

	return stmt
}

func (p *Parser) syntaxError(tok Token, msg string) {
	if tok.Type == TK_EOF {
		p.errors = append(p.errors, &ParseError{Message: "incomplete input"})
	} else {
		p.errors = append(p.errors, &ParseError{Message: msg, Token: tok})
	}
}

func (p *Parser) isAtEnd() bool {
	return p.pos >= len(p.tokens) || p.tokens[p.pos].Type == TK_EOF
}

func (p *Parser) peek() Token {
	if p.isAtEnd() {
		return Token{Type: TK_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.peek().Type == t {
			p.pos++
			return true
		}
	}
	return false
}

func (p *Parser) synchronize() {
	p.pos++
	for !p.isAtEnd() {
		if p.tokens[p.pos-1].Type == TK_SEMI {
			return
		}
		switch p.peek().Type {
		case TK_CREATE, TK_BEGIN, TK_COMMIT, TK_ROLLBACK:
			return
		}
		p.pos++
	}
}
