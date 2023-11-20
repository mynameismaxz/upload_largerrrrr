package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	UPLOAD_PATH = "uploads"
)

var fileChunkCount map[string]int

func init() {
	fileChunkCount = make(map[string]int)
}

func main() {
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))
	// if you want to use logger in the middleware, enable this.
	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// making static path for static folder
	e.Static("/static", "static")
	e.POST("/upload", uploadHandler)
	e.POST("/upload_chunk", uploadChunkHandler)

	// routine for cleanup file in upload folder (every 1 minute)
	go func() {
		for {
			fmt.Println("Clean up upload folder")
			// remove all files in upload folder but not remove folder
			os.RemoveAll(UPLOAD_PATH)
			// create upload folder
			os.MkdirAll(UPLOAD_PATH, os.ModePerm)
			time.Sleep(1 * time.Minute)
		}
	}()

	e.Logger.Fatal(e.Start(":8090"))
}

// Upload handler is a function to handle upload path from "/upload" endpoint.
// That sutable for upload file with small size.
func uploadHandler(c echo.Context) error {
	// Source file section
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination file section
	dst, err := os.Create(UPLOAD_PATH + "/" + file.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":    strconv.Itoa(http.StatusOK),
		"file":      file.Filename,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	})
}

// Upload chunk handler is a function to handle upload path from "/upload_chunk" endpoint.
// That sutable for upload file with big size.
func uploadChunkHandler(c echo.Context) error {
	chunkIndex, err := strconv.Atoi(c.FormValue("chunkIndex"))
	if err != nil {
		return err
	}

	totalChunks, err := strconv.Atoi(c.FormValue("totalChunks"))
	if err != nil {
		return err
	}

	inputNameWithExtension := c.FormValue("filenameWithExtension")

	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Create a directory to store file chunks
	os.MkdirAll("chunks", os.ModePerm)

	// Create or append to the chunk file
	chunkFilename := "chunks/" + strconv.Itoa(chunkIndex) + "-" + file.Filename
	dst, err := os.OpenFile(chunkFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	// Update the count of received chunks
	fileChunkCount[inputNameWithExtension]++

	// fmt.Printf("%d - %d \n", fileChunkCount[filename], totalChunks)

	// Check if all chunks have been received
	if fileChunkCount[inputNameWithExtension] >= totalChunks {
		// All chunks received, time to reassemble
		if err := reassembleFile(inputNameWithExtension, totalChunks); err != nil {
			return err
		}
		delete(fileChunkCount, inputNameWithExtension) // Clean up the counter
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":    strconv.Itoa(http.StatusOK),
		"file":      inputNameWithExtension,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	})
}

// ReassembleFile is a function to reassemble file from chunks into single file again.
// This helper function is used in uploadChunkHandler in case of all chunks have been received.
func reassembleFile(filename string, totalChunks int) error {
	// Create the final file
	file, err := os.Create("uploads/" + filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for i := 0; i < totalChunks; i++ {
		chunkFilename := "chunks/" + strconv.Itoa(i) + "-" + "blob"
		fmt.Printf("Process chunk filename: %s\n", chunkFilename)

		chunk, err := os.OpenFile(chunkFilename, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			fmt.Println("Error opening chunk file:", err)
			return err
		}

		if _, err = io.Copy(file, chunk); err != nil {
			fmt.Println("Error appending chunk data to file:", err)
			chunk.Close()
			return err
		}
		chunk.Close()

		// Optionally, delete the chunk file after it's been appended
		os.Remove(chunkFilename)
	}

	return nil
}
