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

var (
	// funcKeyWord 函数或者方法关键字
	funcKeyWord = `func`
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
		content []byte //代码内容
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

// Func 方法或者函数
type Func struct {
	FuncName  string    //名称
	Arguments Arguments // 参数
	Returns   Arguments // 返回值
	Comments  bool      // 是否有注释
}

func (f Func) String() string {
	if f.Arguments.maxTypeLength > f.Returns.maxTypeLength {
		f.Returns.maxTypeLength = f.Arguments.maxTypeLength
	} else {
		f.Arguments.maxTypeLength = f.Returns.maxTypeLength
	}

	if f.Arguments.maxFieldLength > f.Returns.maxFieldLength {
		f.Returns.maxFieldLength = f.Arguments.maxFieldLength
	} else {
		f.Arguments.maxFieldLength = f.Returns.maxFieldLength
	}

	f.Arguments.fill()
	f.Returns.fill()

	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("/*%s 方法说明\n", f.FuncName))
	builder.WriteString(fmt.Sprintf("\t参数:\n"))

	for i, argument := range f.Arguments.arguments {
		builder.WriteString(fmt.Sprintf("\t*\t%s\t%s\t参数%d\n", argument.Name, argument.Type, i+1))
	}

	builder.WriteString(fmt.Sprintf("\t返回值:\n"))

	for i, argument := range f.Returns.arguments {
		builder.WriteString(fmt.Sprintf("\t*\t%s\t%s\t返回值%d\n", argument.Name, argument.Type, i+1))
	}

	if commentShow {
		builder.WriteString(fmt.Sprintf("有注释:[%v]\n", f.Comments))
	}

	builder.WriteString(fmt.Sprintf("*/"))

	return builder.String()
}

// Argument 参数
type Argument struct {
	Name string // 名称
	Type string // 类型
}
type Arguments struct {
	arguments      []Argument // 参数
	maxTypeLength  int        // 类型的最大长度
	maxFieldLength int        // 字段的最大长度
}

func (a Arguments) fill() {
	for i, argument := range a.arguments { //这一段是为了对齐，按照最长长度计算，长度不足的补足空格
		if len(argument.Name) != a.maxFieldLength {
			a.arguments[i].Name += strings.Repeat(" ", a.maxFieldLength-len(argument.Name))
		}

		if len(argument.Type) != a.maxTypeLength {
			a.arguments[i].Type += strings.Repeat(" ", a.maxTypeLength-len(argument.Type))
		}
	}
}

func FilterFunc(file *ast.File, fileSet *token.FileSet, source string, funcNames ...string) {
	ast.Inspect(file, func(x ast.Node) bool {
		f, ok := x.(*ast.FuncType) // 如果不是函数类型，那么查看子节点
		if !ok {
			return true
		}
		// 方法类型继续处理
		commentMap := ast.NewCommentMap(fileSet, file, file.Comments)

		if int(f.Pos())+len(funcKeyWord) < int(f.Params.Opening)-1 {
			ft := Func{
				FuncName:  source[int(f.Pos())+len(funcKeyWord) : f.Params.Opening-1],
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

func processArguments(fields []*ast.Field, source string) Arguments {
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

	return Arguments{
		arguments:      arguments,
		maxTypeLength:  maxTypeLength,
		maxFieldLength: maxFieldLength,
	}
}
