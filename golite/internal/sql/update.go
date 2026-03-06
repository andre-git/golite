package sql

func (p *Parser) parseUpdate() *UpdateStmt {
	p.match(TK_UPDATE)

	token := p.peek()
	if token.Type != TK_ID {
		p.syntaxError(token, "expected table name")
		return nil
	}
	p.pos++
	tableName := token.Value

	if !p.match(TK_SET) {
		p.syntaxError(p.peek(), "expected SET")
		return nil
	}

	var sets []UpdateSet
	for {
		colTok := p.peek()
		if colTok.Type != TK_ID {
			p.syntaxError(colTok, "expected column name")
			break
		}
		p.pos++
		
		// Assuming '=' is represented as TK_ID or similar for now
		// In a real lexer, this would be TK_EQ
		p.match(TK_ID) 

		val := p.parseExpr()
		sets = append(sets, UpdateSet{
			Column: colTok.Value,
			Value:  val,
		})

		if !p.match(TK_COMMA) {
			break
		}
	}

	stmt := &UpdateStmt{
		Table: tableName,
		Sets:  sets,
	}

	if p.match(TK_WHERE) {
		stmt.Where = p.parseExpr()
	}

	return stmt
}
