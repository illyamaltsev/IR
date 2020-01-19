package dictionary

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type dictionary struct {
	uniqueWords        []string
	wordsCounter       int64
	uniqueWordsCounter int64
}

func NewEmptyDictionary() *dictionary {
	d := dictionary{
		uniqueWords:        make([]string, 0),
		wordsCounter:       0,
		uniqueWordsCounter: 0,
	}
	return &d
}

func (d *dictionary) FillFromTxt(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		text := scanner.Text() // get new line
		text = strings.Replace(text, ",", "", -1)
		text = strings.Replace(text, "\"", "", -1)
		text = strings.Replace(text, "/", "", -1)
		text = strings.Replace(text, ".", "", -1)
		text = strings.Replace(text, "»", "", -1)
		text = strings.Replace(text, "«", "", -1)
		text = strings.ToLower(text)
		words := strings.Fields(text) // split by space and etc.

		for i := range words {
			d.appendIfNotExist(words[i])
			d.wordsCounter++
		}
		//fmt.Println(words, len(words)) // [one two three four] 4
	}

}

func (d *dictionary) appendIfNotExist(word string) {
	if !stringInArr(word, d.uniqueWords) {
		d.uniqueWords = append(d.uniqueWords, word)
		d.uniqueWordsCounter++
	}
}

func stringInArr(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
