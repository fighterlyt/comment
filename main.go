package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
)

//todo: 如果插入到文件中，并且判断有无
var (
	fileName    = ""
	funcName    = ""
	commentShow = false
)

func main() {
	flag.StringVar(&fileName, "fileName", "", "fileName")
	flag.StringVar(&funcName, "funcName", "UpdateAmount", "funcName")
	flag.BoolVar(&commentShow, "commentShow", commentShow, "是否输出有无注释信息")

	flag.Parse()

	if fileName == `` {
		panic(`fileName 不能为空`)
	}

	var (
		file    *ast.File
		content []byte
		err     error
		fileSet *token.FileSet
	)

	if content, err = ioutil.ReadFile(fileName); err != nil {
		panic(fmt.Sprintf(`读取文件[%s]错误[%s]`, fileName, err.Error()))
	}

	fileSet = token.NewFileSet() // positions are relative to fileSet

	if file, err = parser.ParseFile(fileSet, funcName, string(content), parser.ParseComments); err != nil {
		panic("解析错误" + err.Error())
	}

	if funcName != "" {
		FilterFunc(file, fileSet, string(content), funcName)
	} else {
		FilterFunc(file, fileSet, string(content))
	}
}

type Func struct {
	FuncName  string
	Arguments []Argument
	Returns   []Argument
	Comments  bool
}

func (f Func) String() string {
	builder := &strings.Builder{}
	fmt.Fprintf(builder, "/*%s 方法说明\n", f.FuncName)
	fmt.Fprintf(builder, "\t参数:\n")

	for _, argument := range f.Arguments {
		fmt.Fprintf(builder, "\t*\t%s\t%s\n", argument.Name, argument.Type)
	}

	fmt.Fprintf(builder, "\t返回值:\n")

	for _, argument := range f.Returns {
		fmt.Fprintf(builder, "\t*\t%s\t%s\n", argument.Name, argument.Type)
	}

	if commentShow {
		fmt.Fprintf(builder, "有注释:[%v]\n", f.Comments)
	}

	fmt.Fprintf(builder, "*/")

	return builder.String()
}

// Argument 参数
type Argument struct {
	Name string // 名称
	Type string // 类型
}

func FilterFunc(file *ast.File, fileSet *token.FileSet, source string, funcNames ...string) {
	ast.Inspect(file, func(x ast.Node) bool {
		f, ok := x.(*ast.FuncType)
		if !ok {
			return true
		}

		commentMap := ast.NewCommentMap(fileSet, file, file.Comments)

		if int(f.Pos())+len("func") < int(f.Params.Opening)-1 {
			ft := Func{
				FuncName:  source[int(f.Pos())+len("func") : f.Params.Opening-1],
				Arguments: processArguments(f.Params.List, source),
				Comments:  len(commentMap[f]) != 0,
			}
			if f.Results != nil {
				ft.Returns = processArguments(f.Results.List, source)
			}

			if strings.Contains(ft.FuncName, "(") {
				start := strings.Index(ft.FuncName, ")")
				ft.FuncName = strings.TrimSpace(ft.FuncName[start+1:])
			}

			if len(funcNames) != 0 {
				for _, funcName := range funcNames {
					if strings.TrimSpace(funcName) == ft.FuncName {
						println(ft.String())
						break
					}
				}
			} else {
				println(ft.String())
			}
		}
		return false
	})
}

func processArguments(fields []*ast.Field, source string) []Argument {
	arguments := make([]Argument, 0, len(fields))
	maxFieldLength := 0
	maxTypeLength := 0

	for _, field := range fields {
		if len(field.Names) > 0 {
			typeName := source[field.Type.Pos()-1 : field.Type.End()-1]

			if len(typeName) > maxTypeLength {
				maxTypeLength = len(typeName)
			}

			for _, name := range field.Names {
				if len(name.Name) > maxFieldLength {
					maxFieldLength = len(name.Name)
				}

				arguments = append(arguments, Argument{
					Name: name.Name,
					Type: typeName,
				})
			}
		} else {
			typeName := source[field.Type.Pos()-1 : field.Type.End()-1]
			if len(typeName) > maxTypeLength {
				maxTypeLength = len(typeName)
			}
			if len(typeName) > maxFieldLength {
				maxFieldLength = len(typeName)
			}
			arguments = append(arguments, Argument{
				Name: typeName,
				Type: typeName,
			})
		}
	}

	for i, argument := range arguments {
		if len(argument.Name) != maxFieldLength {
			arguments[i].Name += strings.Repeat(" ", maxFieldLength-len(argument.Name))
		}

		if len(argument.Type) != maxTypeLength {
			arguments[i].Type += strings.Repeat(" ", maxTypeLength-len(argument.Type))
		}
	}

	return arguments
}
