package dictionary

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

type Dictionary struct {
	UniqueWords        []string
	WordsCounter       int64
	UniqueWordsCounter int64
	mutex              *sync.Mutex
	wg                 *sync.WaitGroup
}

// Create an empty dictionary
func NewEmptyDictionary() *Dictionary {
	return &Dictionary{
		UniqueWords:        make([]string, 0),
		WordsCounter:       0,
		UniqueWordsCounter: 0,
		mutex:              &sync.Mutex{},
		wg:                 &sync.WaitGroup{},
	}
}

// Build dictionary from files in given directory
func (d *Dictionary) BuildFromDir(dirname string) {
	for file := range enumerateFiles(dirname) {
		d.wg.Add(1)
		go func(filePath string, wg *sync.WaitGroup) {
			for line := range enumerateFile(filePath) {
				for _, word := range tokenize(line) {
					d.appendIfNotExists(word)
					atomic.AddInt64(&d.WordsCounter, 1)
				}
			}
			wg.Done()
		}(file, d.wg)
	}
	d.wg.Wait()
}

// Add new unique word to the dictionary
func (d *Dictionary) appendIfNotExists(word string) {
	if !stringInArr(word, d.UniqueWords) {
		d.mutex.Lock()
		defer d.mutex.Unlock()
		d.UniqueWords = append(d.UniqueWords, word)
		atomic.AddInt64(&d.UniqueWordsCounter, 1)
	}
}

// Save serialized dictionary to the file
func (d *Dictionary) SaveToFile(filepath string) {
	data := d.toGOB64()
	var file *os.File
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// create a file
		file, err = os.Create(filepath)
		handleError(err)
	} else {
		// open file using READ & WRITE permission
		file, err = os.OpenFile(filepath, os.O_RDWR, 0644)
		handleError(err)
		// clear file content
		file.Truncate(0)
		file.Seek(0, 0)
	}
	defer file.Close()
	// write some text line-by-line to file
	_, err := file.WriteString(data)
	handleError(err)

	// save changes
	err = file.Sync()
	handleError(err)

}

// Serialize dictionary
func (d *Dictionary) toGOB64() string {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(d)
	if err != nil {
		fmt.Println(`failed gob Encode`, err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes())

}

// Get all files from the dir
func enumerateFiles(dirname string) chan string {
	output := make(chan string)
	// walk through the dir in goroutine and send files into output chanel
	go func() {
		filepath.Walk(dirname, func(path string, f os.FileInfo, err error) error {
			if !f.IsDir() {
				output <- path
			}
			return nil
		})
		close(output)
	}()
	return output
}

// Read file by line
func enumerateFile(filename string) chan string {
	output := make(chan string)
	go func() {
		file, err := os.Open(filename)
		if err != nil {
			return
		}
		defer file.Close()
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')

			if err == io.EOF {
				break
			}

			// send each line to our enumeration channel
			output <- line
		}
		close(output)
	}()
	return output
}

// Remove unnecessary charachters from line and split it into array
func tokenize(text string) []string {
	text = strings.Replace(text, ",", "", -1)
	text = strings.Replace(text, "\"", "", -1)
	text = strings.Replace(text, "/", "", -1)
	text = strings.Replace(text, ".", "", -1)
	text = strings.Replace(text, "»", "", -1)
	text = strings.Replace(text, "«", "", -1)
	text = strings.ToLower(text)
	return strings.Fields(text) // split by space and etc.
}

// Does list contains given word
func stringInArr(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
