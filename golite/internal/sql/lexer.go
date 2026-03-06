package sql

import (
	"strings"
)

type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input, pos: 0}
}

func (l *Lexer) Tokenize() []Token {
	var tokens []Token

	for !l.isAtEnd() {
		tok := l.nextToken()
		if tok.Type == TK_ILLEGAL && tok.Value == " " {
			continue
		}
		if tok.Type != TK_ILLEGAL {
		    tokens = append(tokens, tok)
        }
	}

	tokens = append(tokens, Token{Type: TK_EOF, Pos: l.pos})
	return tokens
}

func (l *Lexer) isAtEnd() bool {
	return l.pos >= len(l.input)
}

func (l *Lexer) peek() byte {
	if l.isAtEnd() {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) advance() byte {
	if l.isAtEnd() {
		return 0
	}
	b := l.input[l.pos]
	l.pos++
	return b
}

func (l *Lexer) nextToken() Token {
	if l.isAtEnd() {
		return Token{Type: TK_EOF, Pos: l.pos}
	}

	startPos := l.pos
	c := l.peek()

	if isSpace(c) {
		l.advance()
		for !l.isAtEnd() && isSpace(l.peek()) {
			l.advance()
		}
		return Token{Type: TK_ILLEGAL, Value: " ", Pos: startPos} 
	}

	if isIdStart(c) {
		for !l.isAtEnd() && isIdChar(l.peek()) {
			l.advance()
		}
		val := l.input[startPos:l.pos]
		tokType := l.lookupKeyword(val)
		return Token{Type: tokType, Value: val, Pos: startPos}
	}

	if isDigit(c) || (c == '.' && isDigit(l.peekNext())) {
		l.advance()
		for !l.isAtEnd() && isDigit(l.peek()) {
			l.advance()
		}
		if !l.isAtEnd() && l.peek() == '.' {
			l.advance()
			for !l.isAtEnd() && isDigit(l.peek()) {
				l.advance()
			}
		}
		val := l.input[startPos:l.pos]
		return Token{Type: TK_ID, Value: val, Pos: startPos} 
	}

	if c == '\'' || c == '"' || c == '`' || c == '[' {
		quote := l.advance()
		endQuote := quote
		if quote == '[' {
			endQuote = ']'
		}

		for !l.isAtEnd() {
			n := l.advance()
			if n == endQuote {
				if quote == '\'' && !l.isAtEnd() && l.peek() == '\'' {
					l.advance()
				} else {
					break
				}
			}
		}

		val := l.input[startPos+1 : l.pos-1]
		// Handle escaped quotes
		if quote == '\'' {
			val = strings.ReplaceAll(val, "''", "'")
		}
		
		tokType := TK_STRING
		if quote != '\'' {
			tokType = TK_ID
		}
		return Token{Type: tokType, Value: val, Pos: startPos}
	}

	l.advance()
	val := string(c)
	tokType := TK_ILLEGAL

	switch c {
	case ';':
		tokType = TK_SEMI
	case '(':
		tokType = TK_LP
	case ')':
		tokType = TK_RP
	case ',':
		tokType = TK_COMMA
	case '=':
		tokType = TK_EQ
		if l.peek() == '=' {
			l.advance()
			val = "=="
		}
	case '+':
		tokType = TK_PLUS
	case '-':
		tokType = TK_MINUS
	case '*':
		tokType = TK_STAR
	case '/':
		tokType = TK_SLASH
	case '<':
		tokType = TK_LT
		if l.peek() == '=' {
			l.advance()
			val = "<="
			tokType = TK_LE
		} else if l.peek() == '>' {
			l.advance()
			val = "<>"
			tokType = TK_NE
		}
	case '>':
		tokType = TK_GT
		if l.peek() == '=' {
			l.advance()
			val = ">="
			tokType = TK_GE
		}
	case '!':
		if l.peek() == '=' {
			l.advance()
			val = "!="
			tokType = TK_NE
		}
	}

	return Token{Type: tokType, Value: val, Pos: startPos}
}

func (l *Lexer) peekNext() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\f' || c == '\r'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isIdStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c > 127
}

func isIdChar(c byte) bool {
	return isIdStart(c) || isDigit(c)
}

func (l *Lexer) lookupKeyword(s string) TokenType {
	upper := strings.ToUpper(s)
	switch upper {
	case "EXPLAIN":
		return TK_EXPLAIN
	case "QUERY":
		return TK_QUERY
	case "PLAN":
		return TK_PLAN
	case "BEGIN":
		return TK_BEGIN
	case "TRANSACTION":
		return TK_TRANSACTION
	case "DEFERRED":
		return TK_DEFERRED
	case "IMMEDIATE":
		return TK_IMMEDIATE
	case "EXCLUSIVE":
		return TK_EXCLUSIVE
	case "COMMIT":
		return TK_COMMIT
	case "END":
		return TK_END
	case "ROLLBACK":
		return TK_ROLLBACK
	case "SAVEPOINT":
		return TK_SAVEPOINT
	case "RELEASE":
		return TK_RELEASE
	case "TO":
		return TK_TO
	case "CREATE":
		return TK_CREATE
	case "TABLE":
		return TK_TABLE
	case "TEMP", "TEMPORARY":
		return TK_TEMP
	case "IF":
		return TK_IF
	case "NOT":
		return TK_NOT
	case "EXISTS":
		return TK_EXISTS
	case "SELECT":
		return TK_SELECT
	case "FROM":
		return TK_FROM
	case "WHERE":
		return TK_WHERE
	case "GROUP":
		return TK_GROUP
	case "BY":
		return TK_BY
	case "HAVING":
		return TK_HAVING
	case "ORDER":
		return TK_ORDER
	case "LIMIT":
		return TK_LIMIT
	case "INSERT":
		return TK_INSERT
	case "INTO":
		return TK_INTO
	case "VALUES":
		return TK_VALUES
	case "UPDATE":
		return TK_UPDATE
	case "SET":
		return TK_SET
	case "DELETE":
		return TK_DELETE
	case "DISTINCT":
		return TK_DISTINCT
	case "ALL":
		return TK_ALL
	case "AS":
		return TK_AS
	case "DEFAULT":
		return TK_DEFAULT
	case "INTEGER":
		return TK_INTEGER
	case "REAL":
		return TK_REAL
	case "TEXT":
		return TK_TEXT
	case "BLOB":
		return TK_BLOB
	case "NULL":
		return TK_NULL
	case "PRIMARY":
		return TK_PRIMARY
	case "KEY":
		return TK_KEY
	case "AUTOINCREMENT":
		return TK_AUTOINCREMENT
	default:
		return TK_ID
	}
}
