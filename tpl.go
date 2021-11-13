package main

const serviceProxyFile = `package {{.ProxyPackage}}

{{if .TargetImport -}}
import (
	{{- range $index, $element := .TargetImport}}
	"{{$element}}"
	{{- end}}
	"{{.TargetPackage}}"
)
{{- end}}
  
type {{.ProxyClazz}} struct {
	ProxyImpl {{.TargetPackage}}.{{.TargetClazz}} {{PrintAscii 96}}FrameAutowired:"" FrameProxy:""{{PrintAscii 96}}
}

{{- range $index, $element := .Method}}
func (u *{{$.ProxyClazz}}){{$element.MethodName}}({{range $paramIndex, $param := $element.ParamField}}{{if $paramIndex}},{{end}}{{$param.FieldName}} {{FieldTypeStr $param}}{{end}}) {{if gt (len $element.ReturnField) 1}}({{end}}{{range $paramIndex, $param := $element.ReturnField}}{{if $paramIndex}},{{end}}{{FieldTypeStr $param}}{{end}}{{if gt (len $element.ReturnField) 1}}){{end}} {
	//
}
{{- end}}
  
var {{.ProxyInstance}} {{.ProxyClazz}} = {{.ProxyClazz}}{}
  
func init() {
	_ = {{.TargetPackage}}.{{.TargetClazz}}(&{{.ProxyInstance}})
}

`
