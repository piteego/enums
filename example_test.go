package enums_test

import (
	"fmt"
	"github.com/piteego/enums"
)

func ExampleIs() {
	isGreen := enums.Is(color(3), Green)
	fmt.Println(isGreen)
	isGreenOrYellow := enums.Is(color(0), Green, Yellow)
	fmt.Println(isGreenOrYellow)
	// Output:
	// true
	// false
}

func ExampleEnum_IndexOf_withUserDefinedUnknownIndex() {
	red := colorEnum().IndexOf("Red")
	fmt.Println(red)
	unknown := colorEnum().IndexOf("red")
	fmt.Println(unknown)
	// Output:
	// 1
	// -1
}

func ExampleEnum_IndexOf_withDefaultUnknownIndex() {
	on := statusEnum().IndexOf("On")
	fmt.Println(on)
	unknown := statusEnum().IndexOf("off")
	fmt.Println(unknown)
	// Output:
	// 1
	// 0
}

func ExampleEnum_NameOf_withUserDefinedUnknownName() {
	red := colorEnum().NameOf(1)
	fmt.Println(red)
	unknown := colorEnum().NameOf(7)
	fmt.Println(unknown)
	// Output:
	// Red
	// Unknown
}

func ExampleEnum_NameOf_withDefaultUnknownName() {
	on := statusEnum().NameOf(1)
	fmt.Println(on)
	unknown := statusEnum().NameOf(7)
	fmt.Printf("%q", unknown)
	// Output:
	// On
	// ""
}

func ExampleRegister() {
	type enum int8
	const (
		enum1 enum = iota + 1
		enum2
		enum3
		enum4
	)
	var registry enums.Registry[string, enum]
	registry.Once.Do(func() {
		err := enums.Register(&registry.Enum,
			"test.enum", map[string]enum{
				"Enum1": enum1,
				"Enum2": enum2,
				"Enum3": enum3,
				"Enum4": enum4,
			})
		if err != nil {
			panic(err)
		}
	})
	// Output
}
