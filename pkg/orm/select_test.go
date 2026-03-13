package orm

import (
	"testing"
)

func TestSelectString_SQL(t *testing.T) {
	tests := []struct {
		name    string
		builder *SelectString
		want    string
	}{
		{
			name:    "simple query",
			builder: SELECT("id", "name").FROM("users"),
			want:    "SELECT id,name FROM users",
		},
		{
			name:    "query with join",
			builder: SELECT("id", "name").FROM("users").JOIN("orders").ON("users.id = orders.user_id"),
			want:    "SELECT id,name FROM users JOIN orders ON users.id = orders.user_id",
		},
		{
			name:    "query with join",
			builder: SELECT("id", "name").FROM("users").JOIN("orders").ON("users.id = orders.user_id").JOIN("orders2").ON("orders.id = orders2.order_id"),
			want:    "SELECT id,name FROM users JOIN orders ON users.id = orders.user_id JOIN orders2 ON orders.id = orders2.order_id",
		},

		{
			name:    "query with left join",
			builder: SELECT("id", "name").FROM("users").LEFT_JOIN("orders").ON("users.id = orders.user_id"),
			want:    "SELECT id,name FROM users LEFT JOIN orders ON users.id = orders.user_id",
		},
		{
			name:    "query with right join",
			builder: SELECT("id", "name").FROM("users").RIGHT_JOIN("orders").ON("users.id = orders.user_id"),
			want:    "SELECT id,name FROM users RIGHT JOIN orders ON users.id = orders.user_id",
		},
		{
			name:    "query with full join",
			builder: SELECT("id", "name").FROM("users").FULL_JOIN("orders").ON("users.id = orders.user_id"),
			want:    "SELECT id,name FROM users FULL JOIN orders ON users.id = orders.user_id",
		},
		{
			name:    "query with cross join",
			builder: SELECT("id", "name").FROM("users").CROSS_JOIN("orders").ON("users.id = orders.user_id"),
			want:    "SELECT id,name FROM users CROSS JOIN orders ON users.id = orders.user_id",
		},

		{
			name:    "query with where",
			builder: SELECT("id", "name").FROM("users").WHERE(map[string]any{"id = ?": 1}),
			want:    "SELECT id,name FROM users WHERE 1 = 1 id = ?",
		},
		{
			name:    "query with group by",
			builder: SELECT("id", "name").FROM("users").WHERE(map[string]any{"id = ?": 1}).GROUP_BY("id"),
			want:    "SELECT id,name FROM users WHERE 1 = 1 id = ? GROUP BY id",
		},
		{
			name:    "query with having",
			builder: SELECT("id", "name").FROM("users").WHERE(map[string]any{"id = ?": 1}).GROUP_BY("id").HAVING("id > 0"),
			want:    "SELECT id,name FROM users WHERE 1 = 1 id = ? GROUP BY id HAVING id > 0",
		},
		{
			name:    "query with order by",
			builder: SELECT("id", "name").FROM("users").WHERE(map[string]any{"id = ?": 1}).GROUP_BY("id").HAVING("id > 0").ORDER_BY("id DESC"),
			want:    "SELECT id,name FROM users WHERE 1 = 1 id = ? GROUP BY id HAVING id > 0 ORDER BY id DESC",
		},
		{
			name:    "query with limit",
			builder: SELECT("id", "name").FROM("users").WHERE(map[string]any{"id = ?": 1}).GROUP_BY("id").HAVING("id > 0").ORDER_BY("id DESC").LIMIT(10),
			want:    "SELECT id,name FROM users WHERE 1 = 1 id = ? GROUP BY id HAVING id > 0 ORDER BY id DESC LIMIT 10",
		},
		{
			name:    "query with offset",
			builder: SELECT("id", "name").FROM("users").WHERE(map[string]any{"id = ?": 1}).GROUP_BY("id").HAVING("id > 0").ORDER_BY("id DESC").LIMIT(10).OFFSET(10),
			want:    "SELECT id,name FROM users WHERE 1 = 1 id = ? GROUP BY id HAVING id > 0 ORDER BY id DESC LIMIT 10 OFFSET 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.builder.SQL()
			if got != tt.want {
				t.Errorf("SQL(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestSelectCTE_SQL(t *testing.T) {
	tests := []struct {
		name    string
		builder *SelectString
		want    string
	}{
		{
			name:    "query with cte",
			builder: SELECT("id", "name").FROM("users").CTE(WITH("user_cte").AS(SELECT("id", "name").FROM("users").WHERE(map[string]any{"AND age > ?": 25}).SQL()).SQL()),
			want:    "WITH user_cte AS (SELECT id,name FROM users WHERE 1 = 1 AND age > ?)\nSELECT id,name FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.builder.SQL()
			if got != tt.want {
				t.Errorf("SQL(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestSelectModel_SQL(t *testing.T) {
	tests := []struct {
		name    string
		builder *SelectModel
		want    string
	}{
		{
			name:    "query with model",
			builder: SELECT1(&ExampleModel{}).FROM("users").WHERE(map[string]any{"AND age > ?": 25}),
			want:    "SELECT id,name,age FROM users WHERE 1 = 1 AND age > ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.builder.SQL()
			if got != tt.want {
				t.Errorf("SQL(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestSelectModels_SQL(t *testing.T) {
	tests := []struct {
		name    string
		builder *SelectModels[ExampleModel, *ExampleModel]
		want    string
	}{
		{
			name:    "query with models",
			builder: SELECT2(&[]*ExampleModel{}).FROM("users").WHERE(map[string]any{"AND age > ?": 25}),
			want:    "SELECT id,name,age FROM users WHERE 1 = 1 AND age > ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.builder.SQL()
			if got != tt.want {
				t.Errorf("SQL(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
