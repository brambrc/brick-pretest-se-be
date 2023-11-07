package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"web-scrap/model"

	"github.com/PuerkitoBio/goquery"
)

type Product struct {
	Name        string
	Description string
	ImageLink   string
	Price       string
	Rating      string
	StoreName   string
}

type ScraperController struct {
	DB *sql.DB
}

func NewScraperController(db *sql.DB) *ScraperController {
	return &ScraperController{DB: db}
}

func (sc *ScraperController) scrapeCategory(db *sql.DB, url string, a int) int {
	var an int = a // Start with the counter passed in
	// Make HTTP request
	res, err := makeRequest(url)
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
		return an
	}
	defer res.Body.Close()

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return an
	}

	// Use wait group to manage concurrency
	var wg sync.WaitGroup

	// Find each product and scrape concurrently
	doc.Find(".css-bk6tzz.e1nlzfl2").EachWithBreak(func(i int, s *goquery.Selection) bool {
		// Stop if we have scraped 100 products
		if an >= 100 {
			return false // Break out of the loop
		}

		// Extract the product URL
		productURL, exists := s.Find("a").Attr("href")
		if exists {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				scrapeProduct(db, url)
			}(productURL)
			an++
		}

		return true // Continue the loop
	})

	// Wait for all goroutines to finish, then return an
	wg.Wait()
	return an
}

func (sc *ScraperController) Scrape(db *sql.DB, counter int) string {

	//define controller function
	controller := NewScraperController(db)

	var paramurl string
	if counter != 0 {
		num := 2
		paramurl = os.Getenv("categoryURL") + "?page=" + fmt.Sprint(num) // Use fmt.Sprint for proper string conversion
	} else {
		paramurl = os.Getenv("categoryURL")
	}
	fmt.Println("Scraping data from: ", paramurl)

	result_counter := controller.scrapeCategory(db, paramurl, counter)

	if result_counter < 100 {
		return controller.Scrape(db, result_counter)
	}

	// Call exportCSV function here after scraping is complete.
	err := model.ExportCSV(db)
	if err != nil {
		return fmt.Sprintf("Error exporting to CSV: %v", err)
	}

	return "Scraping and export completed successfully."
}

func scrapeProduct(db *sql.DB, url string) {
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
	rating := doc.Find("div[data-testid='lblPDPDetailRatingNumber']").Text()

	// Extract product price
	price := doc.Find("div[data-testid='lblPDPDetailProductPrice']").Text()
	price = strings.TrimSpace(price)
	// Here you can clean the data if needed, for example, parsing the price string to a number

	// Print extracted data
	fmt.Println("Product Name:", productName)
	fmt.Println("Image URL:", imageURL)
	fmt.Println("Description:", description)
	fmt.Println("Price:", price)
	fmt.Println("Rating:", rating)
	fmt.Println("Store Name:", storeName)

	// Save to CSV and database...

	products := model.Product{
		Name:        productName,
		Description: description,
		ImageLink:   imageURL,
		Price:       price,
		Rating:      rating,
		StoreName:   storeName,
	}

	input := model.InsertProduct(db, products)

	if input != nil {
		fmt.Println("Error inserting product to database:", input)
		return
	}

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

func removeHtmlTags(input string) string {
	return strings.TrimSpace(regexp.MustCompile("<[^>]*>").ReplaceAllString(input, ""))
}
