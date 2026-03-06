package sql

func (p *Parser) parseInsert() *InsertStmt {
	p.match(TK_INSERT)
	p.match(TK_INTO) 

	token := p.peek()
	if token.Type != TK_ID {
		p.syntaxError(token, "expected table name")
		return nil
	}
	p.pos++
	tableName := token.Value

	var columns []string
	if p.match(TK_LP) {
		for {
			tok := p.peek()
			if tok.Type != TK_ID {
				p.syntaxError(tok, "expected column name")
				break
			}
			p.pos++
			columns = append(columns, tok.Value)
			if !p.match(TK_COMMA) {
				break
			}
		}
		if !p.match(TK_RP) {
			p.syntaxError(p.peek(), "expected ')' after column list")
		}
	}

	stmt := &InsertStmt{
		Table:   tableName,
		Columns: columns,
	}

	if p.match(TK_VALUES) {
		for {
			if !p.match(TK_LP) {
				p.syntaxError(p.peek(), "expected '(' before values")
				break
			}
			var row []Expr
			for {
				expr := p.parseExpr()
				if expr == nil {
					break
				}
				row = append(row, expr)
				if !p.match(TK_COMMA) {
					break
				}
			}
			if !p.match(TK_RP) {
				p.syntaxError(p.peek(), "expected ')' after values")
			}
			stmt.Values = append(stmt.Values, row)
			if !p.match(TK_COMMA) {
				break
			}
		}
	} else if p.peek().Type == TK_SELECT {
		stmt.Select = p.parseSelect()
	}

	return stmt
}
