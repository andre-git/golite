package schema

import "golite/internal/util"

type Schema struct {
	SchemaCookie int
	Generation   int
	Tables       map[string]*Table
	Indexes      map[string]*Index
	FileFormat   uint8
	Encoding     uint8
}

type Table struct {
	Name      string
	Columns   []*Column
	Indexes   []*Index
	RootPgno  util.Pgno
	HasRowid  bool
	IsVirtual bool
	IsView    bool
}

type Column struct {
	Name         string
	Type         string
	Affinity     byte
	NotNull      bool
	IsPrimaryKey bool
}

type Index struct {
	Name         string
	Table        *Table
	Columns      []int // maps to Table.Columns indices
	RootPgno     util.Pgno
	IsUnique     bool
	IsPrimaryKey bool
}
