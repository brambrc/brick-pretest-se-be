# Product Scraper

This program is designed to scrape product information from web pages and save it to a PostgreSQL database, and then export the data to a CSV file.

## Prerequisites

Before you begin, ensure you have met the following requirements:

- You have installed the latest version of [Go](https://golang.org/dl/).
- You have a PostgreSQL database running and accessible.
- You have set up your environment variables or have a `.env` file with your database connection details.

## Cloning the Repository

To clone the repository, run the following command:

```bash
git clone ['https://github.com/brambrc/brick-pretest-se-be.git']
```


## Setting Up the Environment

Create a .env file in the root directory of the project with the following content:

```bash

DB_HOST=[Your database host]
DB_PORT=[Your database port]
DB_USER=[Your database user]
DB_PASSWORD=[Your database password]
DB_NAME=[Your database name]
CATEGORY_URL=[The URL of the category to scrape]

```


## Running the Program
To run the program, execute:

```bash
go run main.go
```


This command will start the scraping process and once 100 products are scraped, it will export the data to products.csv


## Exporting Data
The program will automatically export the scraped data to a CSV file named products.csv in the root directory once the scraping is complete.

