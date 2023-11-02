	package main

	import (
		"fmt"
		"net/http"
		"net/url"
		"sync"
		"strings"
		"database/sql"
		"log"
		"os"
		"encoding/csv"
		"regexp"


		"github.com/PuerkitoBio/goquery"
		_ "github.com/lib/pq"
		"github.com/joho/godotenv"
	)

	type Product struct {
		Name        string
		Description string
		ImageLink   string
		Price       string
		Rating      string
		StoreName   string
	}

	func main(){
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		db, err := connectToDB()
		if err != nil {
			log.Fatal("Error connecting to the database: ", err)
		}
		defer db.Close()

		// Run migrations
		err = migrateDB(db)
		if err != nil {
			log.Fatal("Error migrating database: ", err)
		}


		message := scrap(db, 0)

		fmt.Println(message)
		
		
	}

	func connectToDB() (*sql.DB, error) {
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

	func migrateDB(db *sql.DB) error {
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

	func makeRequest(url string) (*http.Response, error) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
	
		// Set User-Agent and other headers as necessary
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	func scrap(db *sql.DB, counter int) string {
		var paramurl string
		if counter != 0 {
			num := 2
			paramurl = os.Getenv("categoryURL") + "?page=" + string(num)
		} else {
			paramurl = os.Getenv("categoryURL")
		}
		fmt.Println("Scraping data from: ", paramurl)

		result_counter := scrapeCategory(db, paramurl, counter)

		if result_counter <= 100 {
			scrap(db, result_counter)
		} else {
			//stop the loop
			fmt.Println("Scraping completed successfully.")
		}

		//call export csv function

		return "Scraping completed successfully."


	}

	func scrapeCategory(db *sql.DB ,url string, a int) int {
		var an int
		// Make HTTP request
		res, err := makeRequest(url)
		if err != nil {
			fmt.Println("Error making HTTP request:", err)
			return 0
		}
		defer res.Body.Close()

		// Parse HTML
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			fmt.Println("Error parsing HTML:", err)
			return 0
		}

		// Use wait group to manage concurrency
		var wg sync.WaitGroup

		// Find each product and scrape concurrently
		doc.Find(".css-bk6tzz.e1nlzfl2").Each(func(i int, s *goquery.Selection) {
			// Extract the product URL
			productURL, exists := s.Find("a").Attr("href")
			if exists {
				wg.Add(1)
				go func(url string) {
					defer wg.Done()
					scrapeProduct(db, url)
				}(productURL)
				// add counter here after all the loop done, return the a
				an++
				fmt.Println("counter: ", an)
			}
		})

		// Wait for all goroutines to finish, then return an
		wg.Wait()
		return an
	}

	func getURLProduct(urldata string) string {

		parsedURL, err := url.Parse(urldata)
		if err != nil {
			fmt.Println("Error parsing URL:", err)
			return ""
		}

		// Parse the query string
		queryParams := parsedURL.Query()

		// Extract the query string
		rQueryParam := queryParams.Get("r")
		if rQueryParam == "" {
			fmt.Println("The 'r' query parameter is not present. using default url", urldata)
			rQueryParam = urldata
			return rQueryParam
		} else {
			fmt.Println("The 'r' query parameter is:", rQueryParam)
		}

		return rQueryParam
	}	

	func scrapeProduct(db *sql.DB, url string) {
		// Make HTTP request to the product page

		// get url product page
		url = getURLProduct(url)

		res, err := makeRequest(url)
		if err != nil {
			fmt.Println("Error making HTTP request to product page:", err)
			return
		}
		defer res.Body.Close()

		// Parse HTML of the product page
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			fmt.Println("Error parsing HTML of product page:", err)
			return
		}

		// Extract product name
		productName := doc.Find(".css-1os9jjn").Text()

		// Extract image URL
		imageURL, _ := doc.Find(".css-1c345mg").Attr("src")

		// Extract product description
		descriptionHtml, _ := doc.Find("div[data-testid='lblPDPDescriptionProduk']").Html()
		description := strings.TrimSpace(descriptionHtml)
		description = strings.Replace(description, "<br>", "\n", -1)
		description = removeHtmlTags(description)

		// Extract store name
		storeName := doc.Find(".css-1wdzqxj-unf-heading.e1qvo2ff2").Text()

		// Extract product rating
		score := doc.Find(".score").Text()
		fmt.Println("SCORE: ", score)
		rating := strings.TrimSpace(score)

		// Extract product price
		price := doc.Find("div[data-testid='lblPDPDetailProductPrice']").Text()
		price = strings.TrimSpace(price) 
		// Here you can clean the data if needed, for example, parsing the price string to a number

		// Print extracted data
		fmt.Println("Product Name:", productName)
		// fmt.Println("Image URL:", imageURL)
		// fmt.Println("Description:", description)
		// fmt.Println("Store Name:", storeName)
		// fmt.Println("Rating:", rating)
		// fmt.Println("Price:", price)

		// Save to CSV and database...
		insertProduct(db, Product{
			Name:        productName,
			Description: description,
			ImageLink:   imageURL,
			Price:       price,
			Rating:      rating,
			StoreName:   storeName,
		})
	}

	func insertProduct(db *sql.DB, p Product) error {
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

	func exportCSV(db *sql.DB) error {
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

	func removeHtmlTags(input string) string {
		return strings.TrimSpace(regexp.MustCompile("<[^>]*>").ReplaceAllString(input, ""))
	}