package helpers

import (
	"fmt"
	"log"
	"math/rand"
)

func PrintAndLog(message string) {
	log.Println(message)
	fmt.Println(message)
}

// GetRandomLetterSequence returns a sequence of English characters of length n.
func GetRandomLetterSequence(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
