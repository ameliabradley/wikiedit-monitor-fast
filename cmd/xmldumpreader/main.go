package main

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/leebradley/wikiedit-monitor-fast/pkg/base36"
)

type Revision struct {
	ID   int    `xml:"id"`
	Text Text   `xml:"text"`
	Sha1 string `xml:"sha1"`
}

type Text struct {
	Deleted string `xml:"deleted,attr"`
	Value   string `xml:",chardata"`
}

type Page struct {
	Title     string     `xml:"title"`
	Revisions []Revision `xml:"revision"`
	// ID string `xml:"revision>id"`
	// Deleted string `xml:"revision>text>deleted,attr"`
}

func main() {
	decoder := xml.NewDecoder(os.Stdin)
	fmt.Println("STARTING")

	for {
		// fmt.Println("TOKEN")
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			// If we just read a StartElement token
			// ...and its name is "page"
			if se.Name.Local == "page" {
				var p Page
				// decode a whole chunk of following XML into the
				// variable p which is a Page (se above)
				decoder.DecodeElement(&p, &se)

				for _, rev := range p.Revisions {
					hash := sha1.Sum([]byte(rev.Text.Value))
					encoded := base36.EncodeBytes(hash[:])
					// TODO: Instead of comparing with Wikipedia's hash
					// I'll want to compare with my own hash for the stream revision
					if rev.Sha1 != encoded {
						fmt.Printf("BAD  %s %s\n", rev.Sha1, encoded)
					} else {
						fmt.Printf("GOOD %s\n", rev.Sha1)
					}
				}
				// fmt.Printf("%+v\n", p)
			}
		}
	}
}
