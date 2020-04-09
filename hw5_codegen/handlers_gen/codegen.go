// go build gen/* && ./codegen.exe pack/unpack.go  pack/marshaller.go
// go run pack/*
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

type resulttpl struct {
	ResultName string
	ErrorName  string
}
type fieldtpl struct {
	StructName string
	Field
}
type Field struct {
	UrlName   string
	FieldName string
	FieldType string
	Paramname string
	Default   string
	Min       int
	Max       int
	Required  bool
	Enum      []string
}

func NewField(_name string, _type string) *Field {
	var f = new(Field)
	f.UrlName = strings.ToLower(_name)
	f.FieldName = _name
	f.FieldType = _type
	f.Min = -1
	f.Max = -1
	//f.Enum = map[string]struct{}{}
	return f
}

const (
	_required  string = "required"
	_paramname string = "paramname"
	_enum      string = "enum"
	_default   string = "default"
	_min       string = "min"
	_max       string = "max"
)

func minusFunc(x, y int) int {
	return x - y
}

var minus = template.FuncMap{"minus": minusFunc,}

var (
	resultTpl = template.Must(template.New("resultTpl").Parse(`
if {{.ErrorName}}!=nil{
switch {{.ErrorName}}.(type) {
			case  ApiError:
				w.WriteHeader({{.ErrorName}}.(ApiError).HTTPStatus)
			default  :
				w.WriteHeader(http.StatusInternalServerError)
			}
	answer,answerErr=json.Marshal(MyResponse{Error:{{.ErrorName}}.Error(),Response:nil})
} else {
	answer,answerErr=json.Marshal(MyResponse{Error:"",Response:{{.ResultName}}})
}
	if answerErr!=nil{
		w.WriteHeader(500)
	} else{
			w.Write(answer)
	}
	return

  	`))
	fieldValidatorTpl = template.Must(template.New("fieldValidatorTpl").Funcs(minus).Parse(`
	// {{- .FieldName}}
    {{if .Paramname -}}
	{{.UrlName}}:=r.FormValue("{{.Paramname -}}")
	{{else}}
	{{.UrlName}}:=r.FormValue("{{.UrlName}}")
	{{end}}
	{{if .Default}}
	if {{.UrlName}}==""{
		{{.UrlName}}="{{.Default}}"
	}
	{{end}}
	{{if .Required -}}
	if {{.UrlName}}=="" {
	answer,answerErr=json.Marshal(MyResponse{Error:"{{.UrlName}} must me not empty",Response:nil})
		if answerErr!=nil{
			w.WriteHeader(500)
		} else{
			w.WriteHeader(400)
			w.Write(answer)
		}
	return
	}
{{- end -}}
{{- if (or (ne .Min -1) (ne .Max -1))}}
{{if (eq .FieldType "int")}}
{{.UrlName}}len,err:=strconv.Atoi({{.UrlName}}) 
if err!=nil{
	answer,answerErr=json.Marshal(MyResponse{Error:"{{.UrlName}} must be int",Response:nil})
		if answerErr!=nil{
			w.WriteHeader(500)
		} else{
			w.WriteHeader(400)
			w.Write(answer)
		}
	return
}
{{else}}
{{.UrlName}}len:=len({{.UrlName}})
{{end}}
{{- if (ne .Min -1)}}
if {{.UrlName}}len<{{.Min}} {
{{if (eq .FieldType "int")}}
answer,answerErr=json.Marshal(MyResponse{Error:"{{.UrlName}} must be >= {{.Min}}",Response:nil})
{{else}}
answer,answerErr=json.Marshal(MyResponse{Error:"{{.UrlName}} len must be >= {{.Min}}",Response:nil})
{{end}}
		if answerErr!=nil{
			w.WriteHeader(500)
		} else{
			w.WriteHeader(400)
			w.Write(answer)
		}
	return
}
{{end}}
{{- if (ne .Max -1)}}
if {{.UrlName}}len>{{.Max}} {
{{if (eq .FieldType "int")}}
answer,answerErr=json.Marshal(MyResponse{Error:"{{.UrlName}} must be <= {{.Max}}",Response:nil})
{{else}}
answer,answerErr=json.Marshal(MyResponse{Error:"{{.UrlName}} len must be <= {{.Max}}",Response:nil})
{{end}}
		if answerErr!=nil{
			w.WriteHeader(500)
		} else{
			w.WriteHeader(400)
			w.Write(answer)
		}
	return
}
{{end}}
{{end -}}
{{if .Enum}}
{{$cnt:= .Enum|len}}
{{$last:= (minus $cnt 1)}}
if {{range $i,$k:=.Enum}}{{if (eq $i $last)}}{{$.UrlName}}!="{{$k}}"{{else}}{{$.UrlName}}!="{{$k}}" && {{end}}{{end}} {
answer,answerErr=json.Marshal(MyResponse{Error:"{{$.UrlName}} must be one of "+strings.ReplaceAll("{{.Enum}}"," ",", "),Response:nil})
		if answerErr!=nil{
			w.WriteHeader(500)
		} else{
			w.WriteHeader(400)
			w.Write(answer)
		}
	return
}
{{end}}
{{if (eq .FieldType "int")}}
{{- .StructName}}.{{.FieldName}}={{.UrlName}}len
{{else}}
{{- .StructName}}.{{.FieldName}}={{.UrlName}}
{{end -}}
`))
)

type Comment struct {
	Url    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}
type ReflectStruct struct {
	comment Comment
	inparam []string
	results []string
}

var functions map[string]map[string]ReflectStruct

func main() {

	functions = make(map[string]map[string]ReflectStruct)
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, `import (`)
	fmt.Fprintln(out, `"context"`)
	fmt.Fprintln(out, ` "encoding/json"`)
	fmt.Fprintln(out, ` "net/http"`)
	fmt.Fprintln(out, ` "strconv"`)
	fmt.Fprintln(out, ` "strings"`)
	fmt.Fprintln(out, `)`)
	fmt.Fprintln(out) // empty line

	for _, f := range node.Decls {
		g, ok := f.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if g.Doc != nil {
			for _, coment := range g.Doc.List {

				if strings.Index(coment.Text, "// apigen:api") == 0 {
					inparams := []string{}
					if g.Type.Params != nil {
						for _, inparam := range g.Type.Params.List {
							if nameinparam, ok := inparam.Type.(*ast.Ident); ok {
								inparams = append(inparams, nameinparam.Name)
							}
						}
					}
					results := []string{}

					if g.Type.Params != nil {
						for _, result := range g.Type.Results.List {
							nameresult := ""
							switch e := result.Type.(type) {
							case *ast.Ident:
								nameresult = e.Name
								break
							case *ast.StarExpr:
								nameresult = e.X.(*ast.Ident).Name
							}
							results = append(results, nameresult)
						}
					}
					if g.Recv != nil {

						for _, reciver := range g.Recv.List {
							var namestr string
							switch e := reciver.Type.(type) {
							case *ast.Ident:
								namestr = e.Name
								break
							case *ast.StarExpr:
								namestr = e.X.(*ast.Ident).Name
							}

							if _, ok := functions[namestr]; !ok {
								functions[namestr] = map[string]ReflectStruct{}
							}
							var comment Comment
							if err := json.Unmarshal([]byte(strings.ReplaceAll(coment.Text, "// apigen:api", "")), &comment); err != nil {
								log.Fatal("%v", err)
							}
							functions[namestr][g.Name.Name] = ReflectStruct{
								comment: comment,
								inparam: inparams,
								results: results,
							}

						}
					}
				}
			}
		}
	}
	for structkey, structvalue := range functions {
		fmt.Fprintln(out, "func (srv *"+structkey+") ServeHTTP(w http.ResponseWriter,r *http.Request) {")
		fmt.Fprintln(out, "switch r.URL.Path {")
		for funckey, funcvalue := range structvalue {
			fmt.Fprintln(out, "case "+`"`+funcvalue.comment.Url+`":`)
			if funcvalue.comment.Method != "" {
				fmt.Fprintln(out, "if r.Method=="+`"`+funcvalue.comment.Method+`" {`)
				if funcvalue.comment.Auth {
					checkAutorized(out, funckey)
				}
				fmt.Fprintln(out, "} else {") //error
				errorAnswer(out, 406, "bad method")
				fmt.Fprintln(out, "}")
			} else {
				if funcvalue.comment.Auth {
					checkAutorized(out, funckey)
				}
				fmt.Fprintln(out, "srv.handle"+funckey+"(w,r)")
			}
		}
		fmt.Fprintln(out, "default:")
		errorAnswer(out, 404, "unknown method")
		fmt.Fprintln(out, "}")
		fmt.Fprintln(out, "}")
		fmt.Fprintln(out)
		for funckey, funcvalue := range structvalue {
			fmt.Fprintln(out, "func (srv *"+structkey+") handle"+funckey+"(w http.ResponseWriter, r *http.Request) {")
			fmt.Fprintln(out, "var ctx=context.Background()")
			fmt.Fprintln(out, "r.ParseForm()")
			fmt.Fprintln(out, "var answer []byte")
			fmt.Fprintln(out, "var answerErr error")

			var res, reserr string
			for _, par := range funcvalue.inparam { //loop by func in param
				fmt.Fprintln(out, "_"+par+":="+par+"{}")
				fields := GetStructValidator(par, node)
				for _, field := range fields {
					fieldValidatorTpl.Execute(out, fieldtpl{"_" + par, field})
				}
				if len(funcvalue.results) >= 1 {
					res = "_" + strings.ToLower(funcvalue.results[0])
				}
				if len(funcvalue.results) <= 2 {
					reserr = "_" + strings.ToLower(funcvalue.results[1])
				}
				fmt.Fprintf(out, "%v,%v:=", res, reserr)
				fmt.Fprintln(out, "srv."+funckey+"(ctx,_"+par+")")
				resultTpl.Execute(out, resulttpl{res, reserr})
			}
			fmt.Fprintln(out, "}")

			fmt.Fprintln(out)
		}

	}
}
func errorAnswer(out *os.File, status int, errText string) (int, error) {
	return fmt.Fprintln(out, `answer,answerErr:=json.Marshal(MyResponse{Error:"`+errText+`",Response:nil})
		if answerErr!=nil{
			w.WriteHeader(500)
		} else{
			w.WriteHeader(`+strconv.Itoa(status)+`)
			w.Write(answer)
		}`)
}
func checkAutorized(out *os.File, funckey string) {
	fmt.Fprintln(out, `if r.Header.Get("X-Auth")=="100500"{`)
	fmt.Fprintln(out, "srv.handle"+funckey+"(w,r)")
	fmt.Fprintln(out, `} else {`)
	errorAnswer(out, 403, "unauthorized")
	fmt.Fprintln(out, `}`)
}
func GetStructValidator(name string, node *ast.File) []Field {
	structfields := []Field{}
	for _, f := range node.Decls {
		g, ok := f.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range g.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				continue
			}
			if name != currType.Name.Name {
				continue
			}
		FieldsLoop:
			for _, field := range currStruct.Fields.List {
				if field.Tag != nil {
					tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
					if tag.Get("apivalidator") == "" {
						continue FieldsLoop
					}
					fieldName := field.Names[0].Name
					fieldType := field.Type.(*ast.Ident).Name
					field, err := fieldToValidator(fieldName, fieldType, tag.Get("apivalidator"))
					if err != nil {
						fmt.Errorf("%v", err)
					} else {
						structfields = append(structfields, *field)
					}
				}
			}
		}
	}
	return structfields

}
func fieldToValidator(fieldName, fieldType, tag string) (field *Field, err error) {
	field = NewField(fieldName, fieldType)
	for _, param := range strings.Split(tag, `,`) {
		kv := strings.Split(param, `=`)
		var key string
		var value string
		switch len(kv) {
		case 1:
			key = kv[0]
		case 2:
			key = kv[0]
			value = kv[1]
		default:
			key = kv[0]
		}
		err = paramToValidator(field, key, value)
	}
	return field, err
}
func paramToValidator(field *Field, key, value string) (err error) {
	switch key {
	case _required:
		field.Required = true
	case _paramname:
		field.Paramname = value
	case _enum:
		for _, enumpart := range strings.Split(value, "|") {
			field.Enum = append(field.Enum, enumpart)
		}
	case _default:
		field.Default = value
	case _min:
		field.Min, err = strconv.Atoi(value)
	case _max:
		field.Max, err = strconv.Atoi(value)
	default:
		return fmt.Errorf("%v", "unknown validator tag key")
	}
	return
}
