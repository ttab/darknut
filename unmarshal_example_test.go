package darknut_test

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ttab/darknut"
	"github.com/ttab/newsdoc"
)

type exampleItem struct {
	UUID       uuid.UUID     `newsdoc:"uuid"`
	Meta       exampleBlock  `newsdoc:"meta,type=example/meta"`
	Associates []exampleLink `newsdoc:"links,rel=associate"`
}

type exampleBlock struct {
	Date      *time.Time `newsdoc:"data.date,format=2006-01-02"`
	Timestamp time.Time  `newsdoc:"data.timestamp"`
	Public    bool       `newsdoc:"data.public"`
}

type exampleLink struct {
	Type string `newsdoc:"type"`
	Name string `newsdoc:"name"`
	Age  *int   `newsdoc:"data.age"`
}

func ExampleUnmarshalDocument() {
	doc := newsdoc.Document{
		UUID: "d04e9871-c3df-4fb0-878f-23f8d5ada7c2",
		Meta: []newsdoc.Block{
			{
				Type: "example/meta",
				Data: newsdoc.DataMap{
					"date":      "2023-08-13",
					"timestamp": "2023-08-21T14:34:47+02:00",
					"public":    "true",
				},
			},
		},
		Links: []newsdoc.Block{
			{
				Rel:  "associate",
				Name: "Ren",
				Type: "dog/chihuahua",
			},
			{
				Rel:  "associate",
				Name: "Stimpy",
				Type: "cat/manx",
				Data: newsdoc.DataMap{"age": "3"},
			},
		},
	}

	var item exampleItem

	err := darknut.UnmarshalDocument(doc, &item)
	if err != nil {
		panic(err)
	}

	fmt.Printf(`
UUID: %v
Date: %v
Timestamp: %v
Public: %v

`, item.UUID, item.Meta.Date, item.Meta.Timestamp, item.Meta.Public)

	for _, a := range item.Associates {
		fmt.Printf("* %s (%s)", a.Name, a.Type)

		if a.Age != nil {
			fmt.Printf(" age: %v", *a.Age)
		}

		fmt.Println()
	}

	// Output:
	// UUID: d04e9871-c3df-4fb0-878f-23f8d5ada7c2
	// Date: 2023-08-13 00:00:00 +0000 UTC
	// Timestamp: 2023-08-21 14:34:47 +0200 CEST
	// Public: true
	//
	// * Ren (dog/chihuahua)
	// * Stimpy (cat/manx) age: 3
}
