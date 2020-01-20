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

	cuckoo "github.com/seiflotfy/cuckoofilter"
)

type Dictionary struct {
	UniqueWords        []string
	WordsCounter       int64
	UniqueWordsCounter int64
	cuckoo             *cuckoo.Filter
	mutex              *sync.Mutex
	wg                 *sync.WaitGroup
}

// Create an empty dictionary
func NewEmptyDictionary() *Dictionary {
	return &Dictionary{
		UniqueWords:        make([]string, 0),
		WordsCounter:       0,
		UniqueWordsCounter: 0,
		cuckoo:             cuckoo.NewFilter(1000000),
		mutex:              &sync.Mutex{},
		wg:                 &sync.WaitGroup{},
	}
}

func worker(lines chan string, wg *sync.WaitGroup, d *Dictionary) {
	// Decreasing internal counter for wait-group as soon as goroutine finishes
	defer wg.Done()

	for line := range lines {
		for _, word := range tokenize(line) {
			d.appendIfNotExists(word)
			atomic.AddInt64(&d.WordsCounter, 1)
		}
	}

}

// Build dictionary from files in given directory
func (d *Dictionary) BuildFromDir(dirname string) {
	lines := make(chan string)
	wg := new(sync.WaitGroup)

	// create a pool of workers
	for i := 0; i < 250; i++ {
		wg.Add(1)
		go worker(lines, wg, d)
	}

	for file := range enumerateFiles(dirname) {
		d.wg.Add(1)
		go func(filePath string, wg *sync.WaitGroup) {
			for line := range enumerateFile(filePath) {
				lines <- line
			}
			wg.Done()
		}(file, d.wg)
	}
	d.wg.Wait()
	// Closing channel (waiting in goroutines won't continue any more)
	close(lines)

	// Waiting for all goroutines to finish (otherwise they die as main routine dies)
	wg.Wait()
}

// Add new unique word to the dictionary
func (d *Dictionary) appendIfNotExists(word string) {
	if !d.cuckoo.Lookup([]byte(word)) {
		d.mutex.Lock()
		defer d.mutex.Unlock()
		d.cuckoo.Insert([]byte(word))
		d.UniqueWords = append(d.UniqueWords, word)
		atomic.AddInt64(&d.UniqueWordsCounter, 1)
	}
	// if !stringInArr(word, d.UniqueWords) {
	// 	d.mutex.Lock()
	// 	defer d.mutex.Unlock()
	// 	d.UniqueWords = append(d.UniqueWords, word)
	// 	atomic.AddInt64(&d.UniqueWordsCounter, 1)
	// }
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
