package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var proxyDic string
var outDic string
var workspace string

func main() {

	var proxyFile string

	flag.StringVar(&proxyDic, "d", "", "-d proxy dictionary")
	flag.StringVar(&proxyFile, "f", "", "-f proxy file")
	flag.StringVar(&outDic, "o", "", "-o dictionary to out put new file")
	flag.Parse()

	workspace = getCurrentAbPath()
	if outDic == "" {
		outDic = workspace
	}

	if proxyFile != "" {
		if !filepath.IsAbs(proxyFile) {
			proxyFile = filepath.Join(workspace, proxyFile)
		}
		GenerateByFile(proxyFile)
	} else if proxyDic != "" {
		if !filepath.IsAbs(proxyDic) {
			proxyDic = filepath.Join(workspace, proxyDic)
		}
		GenerateByDic()
	}

}

func GenerateByFile(proxyFile string) {
	var file fs.FileInfo
	var err error
	if file, err = os.Stat(proxyFile); err == nil {
	} else {
		panic(fmt.Errorf("%s can not open", proxyFile))
	}

	var result *ProxyInterface = &ProxyInterface{}
	result.TargetFile = file.Name()
	result.TargetUri = proxyFile
	//result.TargetClazz = ""
	//result.TargetPackage = ""
	result.TargetAnno = make([]string, 0)
	result.TargetImport = make([]string, 0)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, proxyFile, nil, parser.ParseComments)

	// 把节点和comment 映射起来
	cmap := ast.NewCommentMap(fset, f, f.Comments)
	if err != nil {
		panic(err)
	}
	ast.Inspect(f, func(n ast.Node) bool {
		//var s string
		switch x := n.(type) {
		case *ast.GenDecl:
			if len(x.Specs) > 0 {
				l1 := x.Specs[0]
				if ts, ok := l1.(*ast.TypeSpec); ok {
					if ts.Type != nil {
						switch m1 := ts.Type.(type) {
						case *ast.InterfaceType:
							m1.Interface.IsValid()
							result.TargetClazz = ts.Name.Name
							if comments, ok1 := cmap[n]; ok1 {
								if len(comments) > 0 && len(comments[0].List) > 0 {
									for _, comment := range comments[0].List {
										result.TargetAnno = append(result.TargetAnno, comment.Text)
									}
								}
							}
						case *ast.StructType:
							result.TargetClazz = ts.Name.Name
							if comments, ok1 := cmap[n]; ok1 {
								if len(comments) > 0 && len(comments[0].List) > 0 {
									for _, comment := range comments[0].List {
										result.TargetAnno = append(result.TargetAnno, comment.Text)
									}
								}
							}
						}
					}
				}
			}
		case *ast.File:
			result.TargetPackage = x.Name.Name
		case *ast.ImportSpec:
			if x.Path.Value != "" {
				result.TargetImport = append(result.TargetImport, strings.ReplaceAll(x.Path.Value, "\"", ""))
			}
		case *ast.InterfaceType:
			for _, m := range x.Methods.List {
				methodName := m.Names[0].Name
				if !isFirstCharUpperCase(methodName) {
					continue
				}
				pm := &ProxyMethod{}
				pm.MethodName = methodName
				pm.TargetAnno = make([]string, 0)
				pm.ParamField = make([]*ProxyField, 0)
				pm.ReturnField = make([]*ProxyField, 0)
				result.Method = append(result.Method, pm)

				if m.Doc != nil {
					for _, c := range m.Doc.List {
						pm.TargetAnno = append(pm.TargetAnno, c.Text)
					}
				}
				fu := m.Type.(*ast.FuncType)

				if fu.Results != nil {
					for _, f := range fu.Results.List {
						pf := getProxyField(f, cmap)
						pm.ReturnField = append(pm.ReturnField, pf)
					}
				}

				if fu.Params != nil {
					for _, f := range fu.Params.List {
						pf := getProxyField(f, cmap)
						pm.ParamField = append(pm.ParamField, pf)
					}
				}
			}
		case *ast.FuncDecl:
			if x.Recv == nil || len(x.Recv.List) == 0{
				break
			}
			methodName := x.Name.String()

			pm := &ProxyMethod{}
			pm.MethodName = methodName
			pm.TargetAnno = make([]string, 0)
			pm.ParamField = make([]*ProxyField, 0)
			pm.ReturnField = make([]*ProxyField, 0)
			result.Method = append(result.Method, pm)

			if x.Doc != nil {
				for _, c := range x.Doc.List {
					pm.TargetAnno = append(pm.TargetAnno, c.Text)
				}
			}
			fu := x.Type

			if fu.Results != nil {
				for _, f := range fu.Results.List {
					pf := getProxyField(f, cmap)
					pm.ReturnField = append(pm.ReturnField, pf)
				}
			}

			if fu.Params != nil {
				for _, f := range fu.Params.List {
					pf := getProxyField(f, cmap)
					pm.ParamField = append(pm.ParamField, pf)
				}
			}

		}
		return true
	})

	//targetFile := strings.ReplaceAll(result.TargetFile, ".go", "_proxy.go")
	//targetUri := filepath.Join(outDic, targetFile)
	_, proxyPackage := filepath.Split(outDic)
	//fmt.Println(outDic, targetUri, proxyPackage)
	result.ProxyPackage = proxyPackage
	result.ProxyClazz = fmt.Sprintf("%sProxy", result.TargetClazz)
	result.ProxyInstance = firstLower(result.ProxyClazz)

	//fmt.Println(to2PrettyString(result))

	//err = generateProxyFile(result, targetUri)
	classInfo := Cover2ClassInfo(result)
	//fmt.Println(to2PrettyString(classInfo))
	bts,_ := yaml.Marshal(classInfo)
	fmt.Println(string(bts))

}
func GenerateByDic() {

}

func Cover2ClassInfo(source *ProxyInterface) *ClassV1{
	result := &ClassV1{}
	result.Name = source.TargetClazz
	//result.PkgName = fmt.Sprintf("%s.%s",source.TargetPackage,
	//	source.TargetClazz)
	for _,rowAnno := range source.TargetAnno {
		anno,_ := ParseAnnotation(rowAnno)
		if len(anno) > 0 {
			result.Annotation = append(result.Annotation,anno...)
		}
	}
	for _,method := range source.Method {
		m := &ClassMethodV1{}
		m.Name = method.MethodName
		for _,methodAnno := range method.TargetAnno {
			anno,_ := ParseAnnotation(methodAnno)
			if len(anno) > 0{
				m.Annotation =append(m.Annotation,anno...)
			}
		}

		for _,field := range method.ParamField {
			f := &ClassParameterV1{}
			f.Name = field.FieldName
			for _,fieldAnno := range field.TargetAnno {
				anno,_ := ParseAnnotation(fieldAnno)
				if len(anno) > 0{
					f.Annotation =append(f.Annotation,anno...)
				}
			}
			m.In = append(m.In,f)
		}

		result.Method = append(result.Method,m)
	}
	return result
}
