package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

func main() {

	data := runWebScraper()
	m := make(map[string]string)
	// fmt.Println(data)
	for i := 0; i < len(data); i++ {
		row := data[i]
		fmt.Println(row)

		m[row[0].Name] = row[0].Status

	}

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(m)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	godotenv.Load(".env")

	SpaceRegion := os.Getenv("DO_SPACE_REGION")
	accessKey := os.Getenv("ACCESS_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	SpaceName := os.Getenv("DO_SPACE_NAME")
	spaceEndpoint := os.Getenv("SPACE_ENDPOINT")

	fmt.Println("space", SpaceRegion)

	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:    aws.String(spaceEndpoint),
		Region:      aws.String(SpaceRegion),
	}

	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)

	object := s3.PutObjectInput{
		Bucket: aws.String(SpaceName),
		Key:    aws.String("output2.json"),
		Body:   bytes.NewReader(buf.Bytes()),
		ACL:    aws.String("public-read"),
	}

	_, uploadErr := s3Client.PutObject(&object)

	if uploadErr != nil {
		fmt.Errorf("error uploading file: %v", uploadErr)
	}

}

// TrailStatus represents the data structure for each trail
type TrailStatus struct {
	Name   string
	Status string
}

// getStatusFromTitle extracts the status from the title attribute of the img element
func getStatusFromTitle(title string) string {
	// Assuming the title is in the format "Open", "Closed", etc.
	// Adjust this function based on the actual format of the title attribute.
	return strings.TrimSpace(title)
}

func runWebScraper() [][]TrailStatus {
	fmt.Println("Running Webscraper")

	// Make an HTTP request to the URL
	res, err := http.Get("https://mountaincreek.com/skiing-riding/mountain-report/overview/")
	if err != nil {
		fmt.Println("Error while making the HTTP request:", err)
		return nil
	}
	defer res.Body.Close()

	// Load the HTML document using goquery
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("Error while loading the HTML document:", err)
		return nil
	}

	// Find the header that contains "TRAIL & LIFT STATUS"
	header := doc.Find("h2:contains('TRAIL & LIFT STATUS')").First()

	// Check if the header was found
	if header != nil {
		// Print the text content of the header
		fmt.Println("Header:", header.Text())
	}

	// Find all tables that are siblings of the header
	tables := header.NextUntil(":header").Filter("table")

	// Extract the data from the table
	liftTrailData := make([][]TrailStatus, 0)

	// Process each table
	tables.Each(func(j int, table *goquery.Selection) {
		// Process table rows
		table.Find("tr").Each(func(i int, row *goquery.Selection) {
			rowData := make([]TrailStatus, 0)
			// Extract the data from each table cell (td)

			cells := row.Find("td")
			if cells.Length() >= 2 {
				// Extract the name from the first td
				name := cells.Eq(0).Text()

				// Extract the status from the title attribute of the img element in the second td
				imgTitle := cells.Eq(1).Find("img.trailicon").AttrOr("title", "")
				status := getStatusFromTitle(imgTitle)

				// Create a TrailStatus struct and append it to the array
				trail := TrailStatus{Name: name, Status: status}
				rowData = append(rowData, trail)
			}

			// row.Find("td").Each(func(j int, cell *goquery.Selection) {
			// 	fmt.Printf("%s\t", cell.Text())

			// 	text := strings.TrimSpace(cell.Text())
			// 	title := cell.Find("img.trailicon").AttrOr("title", "")

			// 	rowData = append(rowData, text)
			// 	rowData = append(rowData, title)
			// })
			if len(rowData) > 0 {
				// fmt.Println((rowData))
				liftTrailData = append(liftTrailData, rowData)
			}
		})
		fmt.Println() // Separate tables with a blank line
	})

	fmt.Println("Webscraper Finished: ", liftTrailData)

	return liftTrailData

	// Store the extracted data in cache
	// (implementation left as an exercise for the reader)
}
