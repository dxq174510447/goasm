package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

type ProxyInterface struct {

	// 代理对象包名
	ProxyPackage string

	// 代理对象类名
	ProxyClazz string

	// 代理对象实例名
	ProxyInstance string

	// 目录包
	TargetPackage string
	// 目标类名
	TargetClazz string
	// 目标文件名
	TargetFile string
	// 目标文件绝对路径
	TargetUri string

	// 引入的包
	TargetImport []string

	// 目标对象注释
	TargetAnno []string

	// 方法
	Method []*ProxyMethod
}

type ProxyMethod struct {
	MethodName  string
	TargetAnno  []string
	ParamField  []*ProxyField
	ReturnField []*ProxyField
}

type ProxyField struct {
	FieldName   string
	FieldType   string
	TypePackage string
	TargetAnno  []string

	// 是否是特殊类型 0否 1指针 2 slice 3 map
	SpecialType int
	// slice下的节点
	ElementField *ProxyField
	// map key节点
	KeyField *ProxyField
	// map value节点
	ValueField *ProxyField
}

// https://tehub.com/a/44BceBfRK0
// getCurrentAbPath 最终方案-全兼容
func getCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	tmpDir, _ := filepath.EvalSymlinks(os.TempDir())
	if strings.Contains(dir, tmpDir) {
		return getCurrentAbPathByCaller()
	}
	return dir
}

func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

func isFirstCharUpperCase(content string) bool {
	firstChar := content[0]
	if firstChar < 65 || firstChar > 90 {
		return false
	} else {
		return true
	}
}

func getProxyField1(expr ast.Expr, cmap ast.CommentMap) *ProxyField {
	var fieldName string
	var fieldType string
	var typePackage string
	// 是否是特殊类型 0否 1指针 2 slice 3 map
	var specialType int

	var elementField *ProxyField
	var keyField *ProxyField
	var valField *ProxyField

	var result *ProxyField
	switch field := expr.(type) {
	case *ast.MapType:
		specialType = 3
		keyField = getProxyField1(field.Key, cmap)
		valField = getProxyField1(field.Value, cmap)
	case *ast.ArrayType:
		specialType = 2
		elementField = getProxyField1(field.Elt, cmap)
	case *ast.SelectorExpr:
		fieldType = field.Sel.Name
		if field.X != nil {
			tp := field.X.(*ast.Ident)
			typePackage = tp.Name
		}
	case *ast.Ident:
		fieldType = field.Name
	case *ast.StarExpr:
		specialType = 1
		if field.X != nil {

			switch f1 := field.X.(type) {
			case *ast.SelectorExpr:
				fieldType = f1.Sel.Name
				if f1.X != nil {
					tp := f1.X.(*ast.Ident)
					typePackage = tp.Name
				}
			case *ast.Ident:
				fieldType = f1.Name
			}
		}
	}

	result = &ProxyField{}
	result.FieldName = fieldName
	result.FieldType = fieldType
	result.TypePackage = typePackage
	result.SpecialType = specialType
	result.ElementField = elementField
	result.KeyField = keyField
	result.ValueField = valField

	return result
}

func to2String(r interface{}) string {
	result, er := json.Marshal(r)
	if er != nil {
		panic(er)
	}
	return string(result)
}

func to2PrettyString(r interface{}) string {
	result, er := json.MarshalIndent(r,"", "\t")
	if er != nil {
		panic(er)
	}
	return string(result)
}

func getProxyField(f *ast.Field, cmap ast.CommentMap) *ProxyField {
	var fieldName string
	var annos []string
	if len(f.Names) > 0 {
		fieldName = f.Names[0].Name

		if comments, ok1 := cmap[f.Names[0]]; ok1 {
			if len(comments) > 0 && len(comments[0].List) > 0 {
				for _, comment := range comments[0].List {
					annos = append(annos, comment.Text)
				}
			}
		}

	}

	result := getProxyField1(f.Type, cmap)
	result.FieldName = fieldName

	if len(result.TargetAnno) == 0 {
		if len(annos) > 0 {
			result.TargetAnno = annos
		} else {
			if comments, ok1 := cmap[f]; ok1 {
				if len(comments) > 0 && len(comments[0].List) > 0 {
					for _, comment := range comments[0].List {
						result.TargetAnno = append(result.TargetAnno, comment.Text)
					}
				}
			}
		}
	}
	return result
}

//reg := regexp.MustCompile(`(?m)(^\s+|\s+$)`)
//var removeEmptyRowReg *regexp.Regexp = regexp.MustCompile(`(?m)^\s*$\n`)
//var removeEmptyRowReg *regexp.Regexp = regexp.MustCompile(`(?m)^$\n`)

//func removeEmptyRow(content string) string {
//	return removeEmptyRowReg.ReplaceAllString(content, "")
//}

var tplFunc template.FuncMap = template.FuncMap{
	"PrintAscii": func(ascii int) string {
		return fmt.Sprintf("%c", ascii)
	},
	"FieldTypeStr": func(field *ProxyField) string {
		return getFieldTypeStr(field)
	},
}

func getFieldTypeStr(field *ProxyField) string {
	sb := strings.Builder{}
	// 是否是特殊类型 0否 1指针 2 slice 3 map
	switch field.SpecialType {
	case 1:
		sb.WriteString("*")
		if field.TypePackage != "" {
			sb.WriteString(field.TypePackage)
			sb.WriteString(".")
		}
		sb.WriteString(field.FieldType)
	case 2:
		sb.WriteString("[]")
		sb.WriteString(getFieldTypeStr(field.ElementField))
	case 3:
		sb.WriteString("map[")
		sb.WriteString(getFieldTypeStr(field.KeyField))
		sb.WriteString("]")
		sb.WriteString(getFieldTypeStr(field.ValueField))
	default:
		if field.TypePackage != "" {
			sb.WriteString(field.TypePackage)
			sb.WriteString(".")
		}
		sb.WriteString(field.FieldType)
	}
	return sb.String()
}

var tpl *template.Template = template.Must(template.New("proxyService").Funcs(tplFunc).Parse(serviceProxyFile))

func generateProxyFile(proxy *ProxyInterface, targetUrl string) error {
	buf := &bytes.Buffer{}
	err := tpl.Execute(buf, proxy)
	if err != nil {
		return err
	}
	//m := removeEmptyRow(buf.String())
	err = ioutil.WriteFile(targetUrl, buf.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

// FirstUpper 字符串首字母大写
func firstUpper(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// FirstLower 字符串首字母小写
func firstLower(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}
