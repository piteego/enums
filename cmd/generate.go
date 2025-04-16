package cmd

import (
	"bytes"
	"fmt"
	"go/ast"
	goConstant "go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	tempCode = `
// Code generated by piteego/enums; DO NOT EDIT.
// Version: 1.0.0
// Executed At: {{Now}}
package {{.PackageName}}

const _{{.TypeName | ToLower}}_enum_names = "{{ConcatNames .Values}}"

var (
	_{{.TypeName | ToLower}}_enum_names_offsets = [...]uint8{0, {{CommaSepNameOffsets .Values}}}
	_{{.TypeName | ToLower}}_enum_name2index = map[string]{{.TypeName}}{
		{{range $i, $v := .Values}}{{if $i}}, {{end}}"{{$v.Name}}": {{$v.Value}}{{end}},
	}
)

func _() {
	var x [1]struct{}
	{{range $i, $v := .Values}}_ = x[{{$v.Name}}-{{$v.Value}}]
	{{end}}// This function is used to generate an 'invalid array index' compiler error in case of 
	// any changes in the enum values. Re-run the code generation to fix this.
}

func {{ToTitle .TypeName}}List() []{{.TypeName}} {
	return []{{.TypeName}}{
		{{range $i, $v := .Values}}{{if $i}}, {{end}}{{$v.Value}}{{end}},
	}
}

func {{ToTitle .TypeName}}ListNames() []string {
	return []string{
		{{range $i, $v := .Values}}{{if $i}}, {{end}}"{{$v.Name}}"{{end}},
	}
}

func New{{ToTitle .TypeName}}(name string) {{.TypeName}} {
	if val, exists := _{{.TypeName | ToLower}}_enum_name2index[name]; exists {
		return val
	}
	return {{.TypeName}}(-1) // TODO: return minimum value minus one!
}

// Is checks if the {{.TypeName}} enum value is equal to the target {{.TypeName}} or any of the optional values.
func (x {{.TypeName}}) Is(target {{.TypeName}}, or ...{{.TypeName}}) bool {
	if x == target {
		return true
	}
	for i := range or {
		if x == or[i] {
			return true
		}
	}
	return false
}

// Validate validates the {{.TypeName}} enum value and returns an error if the value is not valid.
func (x {{.TypeName}}) Validate() error {
	switch x {
	case {{CommaSepNamesOfUniqueValues .Values}}:
		return nil

	default:
		return fmt.Errorf("invalid '{{.PackageName}}.{{.TypeName}}' enum value: %d", x)
	}
}

// IsValid true if the {{.TypeName}} enum value is valid.
func (x {{.TypeName}}) IsValid() bool {
	switch x {
	case {{CommaSepNamesOfUniqueValues .Values}}:
		return true
	}
	return false
}
`
)

func newEnum(typeName string) (*enum, error) {
	sourceFile := os.Getenv("GOFILE")
	if sourceFile == "" {
		return nil, fmt.Errorf("failed to get GOFILE environment variable")
	}
	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, sourceFile, nil, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", sourceFile, err)
	}

	info := &types.Info{
		Defs: map[*ast.Ident]types.Object{},
	}
	if _, err = new(types.Config).Check(sourceFile, fs, []*ast.File{node}, info); err != nil {
		return nil, fmt.Errorf("failed to set node into file: %v", err)
	}
	constants := make([]constant, 0)
	for i := range node.Decls {
		switch decl := node.Decls[i].(type) {
		case *ast.GenDecl:
			if decl.Tok != token.CONST {
				continue
			}
			for j := range decl.Specs {
				switch spec := decl.Specs[j].(type) {
				case *ast.ValueSpec:
					for k := range spec.Names {
						obj, ok := info.Defs[spec.Names[k]]
						if !ok {
							return nil, fmt.Errorf("failed to find %s constants", spec.Names[k])
						}
						// Check if the constant has same type
						if strings.HasSuffix(obj.Type().String(), typeName) {
							if val, ok := goConstant.Int64Val(obj.(*types.Const).Val()); ok {
								constants = append(constants, constant{
									Name:  spec.Names[k].Name,
									Value: int(val),
								})
							}
						}
					}
				}
			}
		}
	}
	if len(constants) == 0 {
		return nil, fmt.Errorf("failed to find %s constants", typeName)
	}
	return &enum{
		TypeName:    typeName,
		PackageName: node.Name.Name,
		OutputPath: filepath.Join(
			filepath.Dir(sourceFile),
			strings.ToLower(strings.Replace(sourceFile, ".go", "", 1)+"_enum.go"),
		),
		Values: constants,
	}, nil
}

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

type constant struct {
	Name  string
	Value int
}

type enum struct {
	TypeName    string
	PackageName string
	OutputPath  string
	Values      []constant
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
