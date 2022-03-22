package model

//Wal2JsonMessage defines root of wal2json document
type Wal2JsonMessage struct {
	Change []Wal2JsonChange `json:"change"`
}

//Wal2JsonChange defines children of root documents
type Wal2JsonChange struct {
	Kind         string          `json:"kind"`
	Schema       string          `json:"schema"`
	Table        string          `json:"table"`
	ColumnNames  []string        `json:"columnnames"`
	ColumnTypes  []string        `json:"columntypes"`
	ColumnValues []interface{}   `json:"columnvalues"`
	OldKeys      Wal2JsonOldKeys `json:"oldkeys"`
}

//Wal2JsonOldKeys defines children of OldKeys
type Wal2JsonOldKeys struct {
	KeyNames  []string      `json:"keynames"`
	KeyTypes  []string      `json:"keytypes"`
	KeyValues []interface{} `json:"keyvalues"`
}
