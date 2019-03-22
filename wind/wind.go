package wind

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"os"
)

func savePage(url string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Printf("faled to get %s", url)
	}

	res := ""
	doc.Find("body > #main-column > .section-wrap > .forecast-point-10days > tbody").Each(func(i int, s *goquery.Selection) {
		res += s.Text()
	})
	if err != nil {
		fmt.Print("failes to get DOM")
	}

	filepath := "./tmp.txt"
	ioutil.WriteFile(filepath, []byte(res), os.ModePerm)
}

func formating() {

}
