package dictionary

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

type Dictionary struct {
	uniqueWords        []string
	wordsCounter       int64
	uniqueWordsCounter int64
	mutex              *sync.Mutex
	wg                 *sync.WaitGroup
}

// Create an empty dictionary
func NewEmptyDictionary() *Dictionary {
	return &Dictionary{
		uniqueWords:        make([]string, 0),
		wordsCounter:       0,
		uniqueWordsCounter: 0,
		mutex:              &sync.Mutex{},
		wg:                 &sync.WaitGroup{},
	}
}

// Build dictionary from files in given directory
func (d *Dictionary) BuildFromDir(dirname string) {
	files := enumerateFiles(dirname)

	for file := range files {
		d.wg.Add(1)
		go func(filePath string, wg *sync.WaitGroup) {
			for line := range enumerateFile(filePath) {
				for _, word := range tokenize(line) {
					d.appendIfNotExist(word)
					atomic.AddInt64(&d.wordsCounter, 1)
				}
			}
			wg.Done()
		}(file, d.wg)
	}

	d.wg.Wait()

}

// Add new unique word to the dictionary
func (d *Dictionary) appendIfNotExist(word string) {
	if !stringInArr(word, d.uniqueWords) {
		d.mutex.Lock()
		defer d.mutex.Unlock()
		d.uniqueWords = append(d.uniqueWords, word)
		atomic.AddInt64(&d.uniqueWordsCounter, 1)
	}
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
