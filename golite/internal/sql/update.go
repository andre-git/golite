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
		
		if !p.match(TK_EQ) {
			p.syntaxError(p.peek(), "expected '='")
		}

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
