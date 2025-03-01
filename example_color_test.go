package enums_test

import (
	"github.com/piteego/enums"
	"log"
)

var colorRegistry enums.Registry[string, color]

func colorEnum() enums.Enum[string, color] {
	colorRegistry.Once.Do(func() {
		log.Printf("initializing example color enum...")
		err := enums.Register(&colorRegistry.Enum,
			"traffic.color", map[string]color{
				"Red":    Red,
				"Yellow": Yellow,
				"Green":  Green,
			},
			enums.SetUnknown("Unknown", color(-1)),
		)
		if err != nil {
			log.Fatal(err)
		}
	})
	return colorRegistry.Enum
}

const (
	Red color = iota + 1
	Yellow
	Green
)

type color int
