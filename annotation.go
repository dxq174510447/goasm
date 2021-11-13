package main

import (
	"bufio"
	"bytes"
	"strings"
)


/*
@RequestParam(name="haha",defaultValue=123)
*/


type ClassV1 struct {
	Name       string
	Annotation []*TypeAnnotationV1
	Method     []*ClassMethodV1
}

type ClassMethodV1 struct {
	Name       string
	Annotation []*TypeAnnotationV1
	In         []*ClassParameterV1
}

type ClassParameterV1 struct {
	Name       string
	Annotation []*TypeAnnotationV1
}

type TypeAnnotationV1 struct {
	Name  string
	Value map[string]interface{}
}


func ParseAnnotation(content string) ([]*TypeAnnotationV1,error) {
	 b1 := bytes.NewBufferString(content)
	 b2 := bufio.NewReader(b1)

	 var line []byte
	 var err error

	 var result []*TypeAnnotationV1
	 for {
		 line,_,err = b2.ReadLine()
		 if err != nil {
			 break
		 }
		 row := string(line)
		 row = strings.TrimSpace(row)
		 if row == "" {
			 continue
		 }
		 l1 := strings.Index(row,"@")
		 if l1 < 0 {
			 continue
		 }
		 row = row[l1+1:]
		 l1 = strings.Index(row,"(")

		 var annotationName string
		 annotationVal := make(map[string]interface{})
		 if l1 < 0 {
			 annotationName = row
			 p1 := strings.Index(annotationName," ")
			 if p1 > 0 {
				 annotationName = annotationName[0:p1]
			 }else{
				p1 = strings.Index(annotationName,"*")
				if p1 > 0 {
					annotationName = annotationName[0:p1]
				}
			 }

		 }else{
			 annotationName = row[0:l1]
			 l2 := strings.LastIndex(row,")")
			 val := row[l1+1:l2]
			 vs := strings.Split(val,",")
			 for _,fv := range vs {
				 fv1 := strings.TrimSpace(fv)
				 l3 := strings.Index(fv1,"=")

				 if l3 < 0 && len(vs) == 1 {
					 if fv1[0:1] == "\"" {
						 annotationVal["value"] = fv1[1:len(fv1)-1]
					 }else{
						 annotationVal["value"] = fv1
					 }
					 break
				 }

				 field := fv1[0:l3]
				 fieldVal := fv1[l3+1:]
				 if fieldVal[0:1] == "\"" {
					 fieldVal = fieldVal[1:len(fieldVal)-1]
				 }
				 annotationVal[field] = fieldVal
			 }
		 }
		 ann := &TypeAnnotationV1{}
		 ann.Name = annotationName
		 ann.Value = annotationVal
		 result = append(result,ann)
	 }
	 return result,nil
 }
