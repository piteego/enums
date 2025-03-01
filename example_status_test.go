package enums_test

import (
	"github.com/piteego/enums"
	"log"
)

var statusRegistry enums.Registry[status, int8]

func statusEnum() enums.Enum[status, int8] {
	statusRegistry.Once.Do(func() {
		log.Printf("initializing example status enum...")
		err := enums.Register(&statusRegistry.Enum,
			"traffic.light", map[status]int8{
				Off: 0,
				On:  1,
			},
		)
		if err != nil {
			log.Fatal(err)
		}
	})
	return statusRegistry.Enum
}

const (
	Off status = "Off"
	On  status = "On"
)

type status string
