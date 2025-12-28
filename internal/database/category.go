package database

import (
	"fmt"
)

// Category represents a MoneyWiz category
type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// GetCategories retrieves all categories from the database
func (db *DB) GetCategories() ([]Category, error) {
	query := `
		SELECT Z_PK, ZNAME2
		FROM ZSYNCOBJECT
		WHERE Z_ENT = 19 AND ZNAME2 IS NOT NULL
		ORDER BY ZNAME2
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		err := rows.Scan(&cat.ID, &cat.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}
