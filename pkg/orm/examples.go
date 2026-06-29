package orm

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
)

// This file provides usage examples for the ORM.

// ExampleModel is a sample model for demonstration purposes.

type ExampleModel struct {
	ID   int
	Name string
	Age  int
}

func (m *ExampleModel) Mapping() []*Mapping {
	return []*Mapping{
		{"id", &m.ID, m.ID},
		{"name", &m.Name, m.Name},
		{"age", &m.Age, m.Age},
	}
}

// RunExamples executes a series of example queries.
func RunExamples(db *sql.DB) {
	fmt.Println("Running ORM examples...")

	// SELECT examples
	selectExamples(db)

	// INSERT examples
	insertExamples(db)

	// UPDATE examples
	updateExamples(db)

	// DELETE examples
	deleteExamples(db)
}

func selectExamples(db *sql.DB) {
	fmt.Println("--- SELECT Examples ---")

	// SELECT multiple models
	// SELECT id, name, age FROM users WHERE 1 = 1 AND age > 25
	var rows []*ExampleModel
	SELECT2(&rows).FROM("users").WHERE(map[string]any{"AND age > ?": 25}).DryRun(context.Background())

	// SELECT multiple models of full func
	// SELECT id, name, age FROM users WHERE 1 = 1 AND age > 25 GROUP BY age HAVING count(*) > 1 ORDER BY age DESC LIMIT 10 OFFSET 1
	SELECT2(&rows).
		FROM("users").
		WHERE(map[string]any{"AND age > ?": 25}).
		GROUP_BY("age").
		HAVING("count(*) > 1").
		ORDER_BY("age DESC").
		OFFSET(1).
		LIMIT(10).
		DryRun(context.Background())

	// Simple SELECT
	// SELECT id, name, age FROM users WHERE 1 = 1 AND id = 1
	row := &ExampleModel{}
	SELECT1(row).FROM("users").WHERE(map[string]any{"AND id = ?": 1}).DryRun(context.Background())

	// SELECT with aggregate function
	// SELECT count(*) FROM users
	var count sql.Null[int64]
	SELECT1(COUNT("*", &count)).FROM("users").DryRun(context.Background())

	// SELECT with JOIN
	// SELECT u.id, p.product_name FROM users u JOIN products p ON u.id = p.user_id
	SELECT("u.id", "p.product_name").FROM("users u").JOIN("products p").ON("u.id = p.user_id").DryRun(context.Background())

	// SELECT with CTE (Common Table Expression)
	// WITH user_cte AS (SELECT id, name FROM users WHERE 1 = 1 AND age > 25)
	// SELECT id, name FROM user_cte
	cte := WITH("user_cte").AS(SELECT("id", "name").FROM("users").WHERE(map[string]any{"AND age > ?": 25}).SQL()).SQL()
	SELECT("id", "name").FROM("user_cte").CTE(cte).DryRun(context.Background())

	// SELECT with Subquery
	// SELECT id, name FROM (SELECT id, name FROM users WHERE 1 = 1 AND age > 30) AS old_users
	subquery := SELECT("id", "name").FROM("users").WHERE(map[string]any{"AND age > ?": 30}).SQL()
	SELECT("id", "name").FROM("(" + subquery + ") AS old_users").DryRun(context.Background())

	// JOIN with Subquery
	// SELECT u.id, u.name, p.product_name FROM users u JOIN (SELECT id, product_name FROM products WHERE user_id = 1) AS p ON u.id = p.user_id
	SELECT("u.id", "u.name", "p.product_name").FROM("users u").JOIN("(" + subquery + ") AS p").ON("u.id = p.user_id").DryRun(context.Background())
}

func insertExamples(db *sql.DB) {
	fmt.Println("--- INSERT Examples ---")

	// Simple INSERT
	// INSERT INTO users (name, age) VALUES (?, ?)
	INSERT1().INTO("users").COLUMNS("name", "age").VALUES("John Doe", 30).DryRun(context.Background())

	// INSERT from model
	// INSERT INTO users (name, age) VALUES (?, ?)
	user := &ExampleModel{Name: "Jane Doe", Age: 25}
	INSERT(user).INTO("users").DryRun(context.Background())

	// INSERT multiple models
	// INSERT INTO users (name, age) VALUES (?, ?), (?, ?)
	INSERT(&ExampleModel{Name: "y", Age: 18}, &ExampleModel{Name: "z", Age: 20}).INTO("users").DryRun(context.Background())

	// INSERT with RETURNING clause
	// INSERT INTO users (name, age) VALUES ('John Doe', 30) ON DUPLICATE KEY UPDATE name = name
	INSERT1().INTO("users").COLUMNS("name", "age").VALUES("John Doe", 30).ON("DUPLICATE KEY").UPDATE(map[string]any{"name = name": nil}).DryRun(context.Background())
}

func updateExamples(db *sql.DB) {
	fmt.Println("--- UPDATE Examples ---")

	// Simple UPDATE
	// UPDATE users SET age = ? WHERE 1 = 1 AND name = ?
	UPDATE("users").SET(map[string]any{"age": 31}).WHERE(map[string]any{"AND name = ?": "John Doe"}).DryRun(context.Background())

	// UPDATE from model
	// UPDATE users SET age = ? WHERE 1 = 1 AND id = ?
	user := &ExampleModel{ID: 1, Age: 32}
	UPDATE("users").SET1(user).WHERE(map[string]any{"AND id = ?": user.ID}).DryRun(context.Background())
}

func deleteExamples(db *sql.DB) {
	fmt.Println("--- DELETE Examples ---")

	// Simple DELETE
	// DELETE FROM users WHERE 1 = 1 AND name = ?
	DELETE().FROM("users").WHERE(map[string]any{"AND name = ?": "John Doe"}).DryRun(context.Background(), nil)
}

func transactionExamples(db *sql.DB) {
	fmt.Println("--- Transaction Examples ---")

	// Simple Transaction
	// BEGIN;
	// UPDATE users SET balance = 30 WHERE 1 = 1 AND name = 'Jane Doe';
	// UPDATE users SET balance = 0 WHERE 1 = 1 AND name = 'John Doe';
	// COMMIT;
	tx := Tx(func(ctx context.Context, tx *sql.Tx) error {
		if _, err := UPDATE("users").SET(map[string]any{"balance": 30}).WHERE(map[string]any{"AND name = ?": "Jane Doe"}).Exec(ctx, tx); err != nil {
			return fmt.Errorf("UPDATE Jane Doe : %w", err)
		}
		if _, err := UPDATE("users").SET(map[string]any{"balance": 0}).WHERE(map[string]any{"AND name = ?": "John Doe"}).Exec(ctx, tx); err != nil {
			return fmt.Errorf("UPDATE John Doe : %w", err)
		}
		return nil
	})
	if err := tx.Exec(context.Background(), db); err != nil {
		slog.Error("tx.Do", "err", err)
	}
}
