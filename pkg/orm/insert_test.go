package orm

import (
	"testing"
)

type testInsertModel struct {
	ID   int
	Name string
	Age  int
}

func (m *testInsertModel) Mapping() []*Mapping {
	return []*Mapping{
		{"id", &m.ID, m.ID},
		{"name", &m.Name, m.Name},
		{"age", &m.Age, m.Age},
	}
}

func TestInsert_SQL(t *testing.T) {
	tests := []struct {
		name  string
		build *Insert[*testInsertModel]
		want  string
	}{
		{
			name:  "basic insert",
			build: INSERT(&testInsertModel{Name: "Tom", Age: 18}).INTO("users"),
			want:  "INSERT INTO users (name, age) VALUES (?, ?)",
		},
		{
			name:  "insert multiple models",
			build: INSERT(&testInsertModel{Name: "y", Age: 18}, &testInsertModel{Name: "z", Age: 20}).INTO("users"),
			want:  "INSERT INTO users (name, age) VALUES (?, ?), (?, ?)",
		},
		{
			name:  "insert with columns and values",
			build: INSERT[*testInsertModel]().INTO("users").COLUMNS("name", "age").VALUES("y", 18),
			want:  "INSERT INTO users (name, age) VALUES (?, ?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.build.SQL()
			if got != tt.want {
				t.Errorf("SQL(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
