package generate

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	goConstant "go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Execute() error {
	var (
		// __type is the flag for the enum type name
		__type string
		// __output is the flag for the enum output file name
		//__output string
	)
	flag.StringVar(&__type, "type", "", "the name of the type")
	//flag.StringVar(&__output, "output", "", "the name of the output file")
	flag.Parse()
	if __type == "" {
		log.Fatalf("Type name is required")
	}
	data, err := newEnum(__type)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		panic(err)
	}
	if err = os.WriteFile(data.OutputPath, buf.Bytes(), 0644); err != nil {
		return err
	}
	log.Printf("Successfully generated %s file for %s.%s ...\n", data.OutputPath, data.PackageName, data.TypeName)
	return nil
}

type constant struct {
	Name  string
	Value int64
}

type enum struct {
	TypeName    string
	PackageName string
	OutputPath  string
	Values      []constant
}

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
		decl, isGenDecl := node.Decls[i].(*ast.GenDecl)
		if !isGenDecl {
			continue
		}
		if decl.Tok != token.CONST {
			continue
		}
		for j := range decl.Specs {
			spec, isValueSpec := decl.Specs[j].(*ast.ValueSpec)
			if !isValueSpec {
				continue
			}
			for k := range spec.Names {
				obj, ok := info.Defs[spec.Names[k]]
				if !ok {
					return nil, fmt.Errorf("failed to find %q constants", spec.Names[k])
				}
				constVal := obj.(*types.Const).Val()
				if constVal.Kind() != goConstant.Int {
					return nil, fmt.Errorf("invalid constant kind %q", constVal.Kind())
				}
				// Check if the constant has same type
				if strings.HasSuffix(obj.Type().String(), typeName) {
					if val, exact := goConstant.Int64Val(constVal); exact {
						constants = append(constants, constant{
							Name:  spec.Names[k].Name,
							Value: val,
						})
					}
				}
			}
		}
	}
	if len(constants) == 0 {
		return nil, fmt.Errorf("failed to find %q constants", typeName)
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
