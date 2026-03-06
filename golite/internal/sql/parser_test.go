package sql

import (
	"reflect"
	"testing"
)

func TestParser_BasicStatements(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []Token
		want    *CmdList
		wantErr bool
	}{
		{
			name: "Simple BEGIN",
			tokens: []Token{
				{Type: TK_BEGIN, Value: "BEGIN"},
				{Type: TK_SEMI, Value: ";"},
				{Type: TK_EOF},
			},
			want: &CmdList{
				Statements: []Stmt{
					&BeginStmt{Type: TransDeferred},
				},
			},
		},
		{
			name: "Simple COMMIT",
			tokens: []Token{
				{Type: TK_COMMIT, Value: "COMMIT"},
				{Type: TK_SEMI, Value: ";"},
				{Type: TK_EOF},
			},
			want: &CmdList{
				Statements: []Stmt{
					&CommitStmt{},
				},
			},
		},
		{
			name: "Simple SELECT ALL",
			tokens: []Token{
				{Type: TK_SELECT, Value: "SELECT"},
				{Type: TK_ALL, Value: "ALL"},
				{Type: TK_SEMI, Value: ";"},
				{Type: TK_EOF},
			},
			want: &CmdList{
				Statements: []Stmt{
					&SelectStmt{Distinct: false},
				},
			},
		},
		{
			name: "Simple SELECT DISTINCT",
			tokens: []Token{
				{Type: TK_SELECT, Value: "SELECT"},
				{Type: TK_DISTINCT, Value: "DISTINCT"},
				{Type: TK_SEMI, Value: ";"},
				{Type: TK_EOF},
			},
			want: &CmdList{
				Statements: []Stmt{
					&SelectStmt{Distinct: true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.tokens)
			got, errs := p.Parse()
			
			if (len(errs) > 0) != tt.wantErr {
				t.Errorf("Parser.Parse() errors = %v, wantErr %v", errs, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.Parse() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
