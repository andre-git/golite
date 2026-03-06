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
	TK_MINUS
	TK_STAR
	TK_SLASH
	TK_LT
	TK_GT
	TK_LE
	TK_GE
	TK_NE
	TK_DEFAULT
	TK_INTEGER
	TK_REAL
	TK_TEXT
	TK_BLOB
	TK_NULL
	TK_PRIMARY
	TK_KEY
	TK_AUTOINCREMENT
	TK_JOIN
	TK_INNER
	TK_LEFT
	TK_OUTER
	TK_CROSS
	TK_ON
	TK_USING
	TK_PRAGMA
	TK_DOT
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

type SubqueryExpr struct {
    Select *SelectStmt
}
func (*SubqueryExpr) node() {}
func (*SubqueryExpr) expr() {}

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
	Name     string
	DBName   string
	Alias    string
	Subquery *SelectStmt
	JoinType string
	On       Expr
	Using    []string
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

type PragmaStmt struct {
    Name  string
    DB    string
    Value string
}
func (PragmaStmt) node() {}
func (PragmaStmt) stmt() {}

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
	case TK_PRAGMA:
		return p.parsePragma()
	default:
		p.syntaxError(p.peek(), "syntax error")
		return nil
	}
}

func (p *Parser) parsePragma() *PragmaStmt {
	p.match(TK_PRAGMA)
	stmt := &PragmaStmt{}
	if p.peek().Type == TK_ID {
		stmt.Name = p.peek().Value
		p.pos++
	}
	if p.match(TK_EQ) || p.match(TK_LP) {
		stmt.Value = p.peek().Value
		p.pos++
		p.match(TK_RP)
	}
	return stmt
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

	if p.match(TK_INTEGER, TK_REAL, TK_TEXT, TK_BLOB, TK_ID) {
		col.Type = &ColumnType{Name: p.tokens[p.pos-1].Value}
	}

	for {
		if p.match(TK_DEFAULT) {
			_ = p.parseExpr()
		} else if p.match(TK_PRIMARY) {
			p.match(TK_KEY)
			p.match(TK_AUTOINCREMENT)
		} else {
			break
		}
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
	if tok.Type == TK_MINUS || tok.Type == TK_PLUS {
		p.pos++
		right := p.parsePrimaryExpr()
		return &BinaryExpr{Left: &LiteralExpr{Token: Token{Type: TK_ID, Value: "0"}}, Op: tok, Right: right}
	}
	if tok.Type == TK_ID || tok.Type == TK_STRING {
		p.pos++
		return &LiteralExpr{Token: tok}
	}
	if p.match(TK_LP) {
		if p.peek().Type == TK_SELECT {
			sub := p.parseSelect()
			p.match(TK_RP)
			return &SubqueryExpr{Select: sub}
		}
		expr := p.parseExpr()
		p.match(TK_RP)
		return expr
	}
	return nil
}

func (p *Parser) getPrecedence(t TokenType) int {
	switch t {
	case TK_STAR, TK_SLASH:
		return 11
	case TK_PLUS, TK_MINUS:
		return 10
	case TK_LT, TK_LE, TK_GT, TK_GE:
		return 9
	case TK_EQ, TK_NE:
		return 8
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
			stmt.Columns = append(stmt.Columns, &LiteralExpr{Token: p.peek()})
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
		stmt.From = &SrcList{}
		for {
			var item SrcItem
			if p.match(TK_LP) {
				item.Subquery = p.parseSelect()
				p.match(TK_RP)
			} else if p.peek().Type == TK_ID {
				item.Name = p.peek().Value
				p.pos++
				if p.match(TK_DOT) {
					item.DBName = item.Name
					item.Name = p.peek().Value
					p.pos++
				}
			}

			if p.match(TK_AS) {
				if p.peek().Type == TK_ID {
					item.Alias = p.peek().Value
					p.pos++
				}
			}

			stmt.From.Items = append(stmt.From.Items, item)

			// Handle Joins
			if p.match(TK_JOIN, TK_INNER, TK_LEFT, TK_CROSS) {
				// Very simplified join parsing
				continue 
			}

			if !p.match(TK_COMMA) {
				break
			}
		}
	}

	if p.match(TK_WHERE) {
		stmt.Where = p.parseExpr()
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
