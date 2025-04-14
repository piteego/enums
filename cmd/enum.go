package cmd

import (
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
	"strings"
	"unicode"
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
									Ident: spec.Names[k].Name,
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
		Constants: constants,
	}, nil
}

type constant struct {
	Ident string
	Value int
}

type enum struct {
	TypeName    string
	PackageName string
	OutputPath  string
	Constants   []constant
}

func (e *enum) mapVarName() string { return fmt.Sprintf("_%sEnumMap", e.TypeName) }

func (e *enum) genMapVar() *ast.GenDecl {
	var enumElements []ast.Expr
	for i := range e.Constants {
		enumElements = append(enumElements, &ast.KeyValueExpr{
			Key: &ast.BasicLit{
				Kind:  token.STRING,
				Value: `"` + e.Constants[i].Ident + `"`,
			},
			Value: ast.NewIdent(e.Constants[i].Ident),
		})
	}

	return &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{ast.NewIdent(e.mapVarName())},
				Type: &ast.MapType{
					Key:   ast.NewIdent("string"),
					Value: ast.NewIdent(e.TypeName),
				},
				Values: []ast.Expr{
					&ast.CompositeLit{
						Type: &ast.MapType{
							Key:   ast.NewIdent("string"),
							Value: ast.NewIdent(e.TypeName),
						},
						Elts: enumElements,
					},
				},
			},
		},
	}
}

func (e *enum) genConstantCompilerFunc() *ast.FuncDecl {
	varDecl := &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{ast.NewIdent("x")},
				Type: &ast.ArrayType{
					Len: &ast.BasicLit{Kind: token.INT, Value: "1"},
					Elt: &ast.StructType{Fields: &ast.FieldList{
						Opening: token.Pos(1),
						Closing: token.Pos(1), // empty struct at same line
					}},
				},
			},
		},
	}

	assignments := make([]ast.Stmt, 0, len(e.Constants))
	for i := range e.Constants {
		assignments = append(assignments, &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("_")},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				&ast.IndexExpr{
					X: ast.NewIdent("x"),
					Index: &ast.BinaryExpr{
						X:  ast.NewIdent(e.Constants[i].Ident),
						Op: token.SUB,
						Y:  &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", e.Constants[i].Value)},
					},
				},
			},
		})
	}

	funcBody := &ast.BlockStmt{
		List: append([]ast.Stmt{&ast.DeclStmt{Decl: varDecl}}, assignments...),
	}

	return &ast.FuncDecl{
		Name: ast.NewIdent("_"),
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: funcBody,
	}
}

//func (e *enum) genNameFunc() *ast.FuncDecl {
//	funcName := e.TypeName + "Name"
//	ifStmt := &ast.IfStmt{
//		Cond: &ast.BinaryExpr{
//			X: &ast.CallExpr{
//				Fun:  ast.NewIdent("int"),
//				Args: []ast.Expr{ast.NewIdent("val")},
//			},
//			Op: token.EQL,
//			Y:  ast.NewIdent("value"),
//		},
//		Body: &ast.BlockStmt{
//			List: []ast.Stmt{
//				&ast.ReturnStmt{
//					Results: []ast.Expr{
//						ast.NewIdent("val"),
//						ast.NewIdent("nil"),
//					},
//				},
//			},
//		},
//	}
//
//	forStmts := &ast.RangeStmt{
//		Key:   ast.NewIdent("_"),
//		Value: ast.NewIdent("val"),
//		Tok:   token.DEFINE,
//		X:     ast.NewIdent(e.mapVarName()),
//		Body: &ast.BlockStmt{
//			List: []ast.Stmt{
//				ifStmt,
//			},
//		},
//	}
//
//	returnStmt := &ast.ReturnStmt{
//		Results: []ast.Expr{
//			&ast.BasicLit{Kind: token.STRING, Value: "0"},
//			&ast.CallExpr{
//				Fun: &ast.SelectorExpr{
//					X:   ast.NewIdent("fmt"),
//					Sel: ast.NewIdent("Errorf"),
//				},
//				Args: []ast.Expr{
//					&ast.BasicLit{Kind: token.STRING, Value: `"not found enum with index: %d"`},
//					ast.NewIdent("value"),
//				},
//			},
//		},
//	}
//
//	body := &ast.BlockStmt{
//		List: []ast.Stmt{
//			forStmts,
//			returnStmt,
//		},
//	}
//
//	return &ast.FuncDecl{
//		Name: ast.NewIdent(funcName),
//		Type: &ast.FuncType{
//			Params: &ast.FieldList{
//				List: []*ast.Field{
//					{
//						Names: []*ast.Ident{ast.NewIdent("value")},
//						Type:  ast.NewIdent("int"),
//					},
//				},
//			},
//			Results: &ast.FieldList{
//				List: []*ast.Field{
//					{Type: ast.NewIdent(e.TypeName)},
//					{Type: ast.NewIdent("error")},
//				},
//			},
//		},
//		Body: body,
//	}
//}

func (e *enum) genNamesFunc() *ast.FuncDecl {
	mapName := e.mapVarName()
	funcName := cases.Title(language.English, cases.Compact).String(e.TypeName) + "Names"

	makeLen := &ast.CallExpr{
		Fun: ast.NewIdent("make"),
		Args: []ast.Expr{
			&ast.ArrayType{Elt: ast.NewIdent("string")},
			ast.NewIdent("0"),
			&ast.CallExpr{
				Fun:  ast.NewIdent("len"),
				Args: []ast.Expr{ast.NewIdent(mapName)},
			},
		},
	}
	varDecl := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("names")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{makeLen},
	}

	forDecl := &ast.RangeStmt{
		Key:   ast.NewIdent("name"),
		Value: nil,
		Tok:   token.DEFINE,
		X:     ast.NewIdent(mapName),
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("names")},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: ast.NewIdent("append"),
							Args: []ast.Expr{
								ast.NewIdent("names"),
								&ast.Ident{Name: "name"},
							},
						},
					},
				},
			},
		},
	}

	body := &ast.BlockStmt{
		List: []ast.Stmt{
			varDecl, forDecl,
			&ast.ReturnStmt{
				Results: []ast.Expr{
					ast.NewIdent("names"),
				},
			},
		},
	}
	return &ast.FuncDecl{
		Name: ast.NewIdent(funcName),
		Type: &ast.FuncType{
			Params: nil,
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: &ast.ArrayType{Elt: ast.NewIdent("string")}},
				},
			},
		},
		Body: body,
	}
}

//func (e *enum) genValueFunc() *ast.FuncDecl {
//
//	funcName := e.TypeName + "Index"
//	paramName := findFirstChar(e.TypeName)
//
//	ifStmt := &ast.IfStmt{
//		Cond: &ast.BinaryExpr{
//			X:  ast.NewIdent("value"),
//			Op: token.EQL,
//			Y:  ast.NewIdent(paramName),
//		},
//		Body: &ast.BlockStmt{
//			List: []ast.Stmt{
//				&ast.ReturnStmt{
//					Results: []ast.Expr{
//						&ast.CallExpr{
//							Fun:  ast.NewIdent("int"),
//							Args: []ast.Expr{ast.NewIdent("value")},
//						},
//						ast.NewIdent("nil"),
//					},
//				},
//			},
//		},
//	}
//
//	rangeStmt := &ast.RangeStmt{
//		Key:   ast.NewIdent("_"),
//		Value: ast.NewIdent("value"),
//		Tok:   token.DEFINE,
//		X:     ast.NewIdent(e.mapVarName()),
//		Body: &ast.BlockStmt{
//			List: []ast.Stmt{ifStmt},
//		},
//	}
//
//	returnStmt := &ast.ReturnStmt{
//		Results: []ast.Expr{
//			&ast.BasicLit{Kind: token.INT, Value: "0"},
//			&ast.CallExpr{
//				Fun: &ast.SelectorExpr{
//					X:   ast.NewIdent("fmt"),
//					Sel: ast.NewIdent("Errorf"),
//				},
//				Args: []ast.Expr{
//					&ast.BasicLit{Kind: token.STRING, Value: `"not found enum: %v"`},
//					ast.NewIdent(paramName),
//				},
//			},
//		},
//	}
//
//	return &ast.FuncDecl{
//		Name: ast.NewIdent(funcName),
//		Type: &ast.FuncType{
//			Params: &ast.FieldList{
//				List: []*ast.Field{
//					{
//						Names: []*ast.Ident{ast.NewIdent(paramName)},
//						Type:  ast.NewIdent(e.TypeName),
//					},
//				},
//			},
//			Results: &ast.FieldList{
//				List: []*ast.Field{
//					{Type: ast.NewIdent("int")},
//					{Type: ast.NewIdent("error")},
//				},
//			},
//		},
//		Body: &ast.BlockStmt{
//			List: []ast.Stmt{rangeStmt, returnStmt},
//		},
//	}
//}

func (e *enum) genValuesFunc() *ast.FuncDecl {
	mapName := e.mapVarName()
	funcName := cases.Title(language.English, cases.Compact).String(e.TypeName) + "Values"

	makeLen := &ast.CallExpr{
		Fun: ast.NewIdent("make"),
		Args: []ast.Expr{
			&ast.ArrayType{Elt: ast.NewIdent("int")},
			ast.NewIdent("0"),
			&ast.CallExpr{
				Fun:  ast.NewIdent("len"),
				Args: []ast.Expr{ast.NewIdent(mapName)},
			},
		},
	}

	varDel := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("indexes")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{makeLen},
	}

	forDecl := &ast.RangeStmt{
		Key:   ast.NewIdent("_"),
		Value: ast.NewIdent("index"),
		Tok:   token.DEFINE,
		X:     ast.NewIdent(mapName),
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("indexes")},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: ast.NewIdent("append"),
							Args: []ast.Expr{ast.NewIdent("indexes"),
								&ast.CallExpr{
									Fun:  ast.NewIdent("int"),
									Args: []ast.Expr{ast.NewIdent("index")},
								},
							},
						},
					},
				},
			},
		},
	}

	body := &ast.BlockStmt{
		List: []ast.Stmt{varDel,
			forDecl,
			&ast.ReturnStmt{
				Results: []ast.Expr{ast.NewIdent("indexes")},
			},
		},
	}

	return &ast.FuncDecl{
		Name: ast.NewIdent(funcName),
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: &ast.ArrayType{Elt: ast.NewIdent("int")}},
				},
			},
		},
		Body: body,
	}
}

func (e *enum) genIsMethod() *ast.FuncDecl {

	receiverName := string(unicode.ToLower([]rune(e.TypeName)[0]))

	ifStmt := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  ast.NewIdent(receiverName),
			Op: token.EQL,
			Y:  ast.NewIdent("target"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{&ast.Ident{Name: "true"}},
				},
			},
		},
	}

	ifStmtRange := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.Ident{Name: receiverName},
			Op: token.EQL,
			Y: &ast.IndexExpr{
				X:     ast.NewIdent("or"),
				Index: ast.NewIdent("i"),
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{&ast.Ident{Name: "true"}},
				},
			},
		},
	}

	rangeStmt := &ast.RangeStmt{
		Key:   ast.NewIdent("i"),
		Value: nil,
		Tok:   token.DEFINE,
		X:     ast.NewIdent("or"),
		Body: &ast.BlockStmt{
			List: []ast.Stmt{ifStmtRange},
		},
	}

	returnStmt := &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.Ident{Name: "false"},
		},
	}

	return &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent(receiverName)},
					Type:  ast.NewIdent(e.TypeName),
				},
			},
		},
		Name: ast.NewIdent("Is"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{Names: []*ast.Ident{ast.NewIdent("target")}, Type: ast.NewIdent(e.TypeName)},
					{
						Names: []*ast.Ident{ast.NewIdent("or")},
						Type:  &ast.Ellipsis{Elt: ast.NewIdent(e.TypeName)}},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: ast.NewIdent("bool")},
				},
			},
		},

		Body: &ast.BlockStmt{
			List: []ast.Stmt{ifStmt, rangeStmt, returnStmt},
		},
	}
}
