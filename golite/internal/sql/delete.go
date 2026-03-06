package sql

func (p *Parser) parseDelete() *DeleteStmt {
	p.match(TK_DELETE)
	p.match(TK_FROM)

	token := p.peek()
	if token.Type != TK_ID {
		p.syntaxError(token, "expected table name")
		return nil
	}
	p.pos++
	tableName := token.Value

	stmt := &DeleteStmt{
		Table: tableName,
	}

	if p.match(TK_WHERE) {
		stmt.Where = p.parseExpr()
	}

	return stmt
}
