// Command spam-filter classifies email subjects as spam or not with
// multinomial naive Bayes over words, the method early spam filters (and
// SpamAssassin's Bayes rule) popularized: word frequencies per class, then
// the most likely class for a new message.
package main

import (
	"fmt"
	"log"

	"github.com/braswelljr/alGOrithm/pkg/ml/classification"
)

func main() {
	docs := []string{
		"win a free prize now", "claim your free money", "cheap meds limited offer",
		"you won a free cruise", "meeting moved to monday", "lunch on friday",
		"quarterly report attached", "can you review my pull request",
	}
	labels := []string{"spam", "spam", "spam", "spam", "ham", "ham", "ham", "ham"}
	model, err := classification.FitTextNB(docs, labels)
	if err != nil {
		log.Fatal(err)
	}
	for _, subject := range []string{"free prize inside", "review the report before monday", "limited offer just for you"} {
		class, _ := model.Predict(subject)
		fmt.Printf("%-34q -> %s\n", subject, class)
	}
}
