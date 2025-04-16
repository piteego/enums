package cmd

import (
	"bytes"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
)

func Generate(typeName string) error {
	tmpl, err := template.New("enum").Funcs(template.FuncMap{
		"Now":                         func() string { return time.Now().Format(time.RFC3339) },
		"ToLower":                     strings.ToLower,
		"ToTitle":                     cases.Title(language.English, cases.Compact).String,
		"CommaSepNames":               joinNames(",", false),
		"CommaSepNamesOfUniqueValues": joinNames(", ", true),
		"ConcatNames":                 joinNames("", false),
		"CommaSepNameOffsets":         joinNameOffsets,
	}).Parse(tempCode)
	if err != nil {
		return err
	}
	data, err := newEnum(typeName)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		panic(err)
	}
	return os.WriteFile(data.OutputPath, buf.Bytes(), 0644)
}

// joinNames joins all enum value names using given separator
func joinNames(sep string, onlyUniqueValues bool) func([]constant) string {
	return func(constants []constant) string {
		names := make([]string, 0, len(constants))
		values := make(map[int]struct{}, len(constants))
		for i := range constants {
			if onlyUniqueValues {
				if _, exists := values[constants[i].Value]; exists {
					continue
				}
				values[constants[i].Value] = struct{}{}
			}
			names = append(names, constants[i].Name)
		}
		return strings.Join(names, sep)
	}
}

// joinNameOffsets calculates and joins the byte offsets
func joinNameOffsets(values []constant) string {
	var offsets []string
	current := 0
	for _, v := range values {
		current += len(v.Name)
		offsets = append(offsets, strconv.Itoa(current))
	}
	return strings.Join(offsets, ", ")
}
