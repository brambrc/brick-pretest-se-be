package model

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Product struct {
	Name        string
	Description string
	ImageLink   string
	Price       string
	Rating      string
	StoreName   string
}

type Model struct {
	DB *sql.DB
}

func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	return nil
}

func ConnectToDB() (*sql.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Successfully connected!")
	return db, nil
}

func MigrateDB(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			image_link TEXT,
			price TEXT,
			rating TEXT,
			store_name TEXT
		);`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	fmt.Println("Migration completed successfully.")
	return nil
}

func InsertProduct(db *sql.DB, p Product) error {
	query := `
		INSERT INTO products (name, description, image_link, price, rating, store_name)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;`

	id := 0
	err := db.QueryRow(query, p.Name, p.Description, p.ImageLink, p.Price, p.Rating, p.StoreName).Scan(&id)
	if err != nil {
		return err
	}
	fmt.Printf("New product inserted with id %d\n", id)
	return nil
}

func ExportCSV(db *sql.DB) error {
	query := `
		SELECT name, description, image_link, price, rating, store_name
		FROM products
		ORDER BY id DESC
		LIMIT 100`

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	file, err := os.Create("products.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	header := []string{"Name", "Description", "ImageLink", "Price", "Rating", "StoreName"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data to CSV
	for rows.Next() {
		var name, description, imageLink, price, rating, storeName string
		if err := rows.Scan(&name, &description, &imageLink, &price, &rating, &storeName); err != nil {
			return err
		}
		record := []string{name, description, imageLink, price, rating, storeName}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	fmt.Println("Data exported to products.csv successfully.")
	return nil
}
