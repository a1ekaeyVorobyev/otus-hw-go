package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var ErrWriteTag = errors.New("the tag was written with an error")

const (
	tInt    = "int"
	tInt32  = "int32"
	tInt64  = "int64"
	tString = "string"
)

type PE struct {
	Message string
	Err     error
}

type sCon struct {
	NameStruct string
	Name       string
	Conditions string
	Value      string
	Tag        string
	IsArray    bool
}

type Field struct {
	Name      string
	TypeField string
	Tag       string
	IsArray   bool
}

// шаблоны для генерации.
var (
	textArrayConditions = `
	for _, s := range {{.NameStruct}}.{{.Name}} {
		if {{.Conditions}} {
			ve = append(ve, ValidationError{
				Field:  "{{.Name}}",
				Err:	fmt.Errorf("{{.Name}} does not fulfill the condition tag = {{.Tag}}. Value {{.Name}} = %v", s),
			})
		}
	}`
	textConditions = `
	if {{.Conditions}} {
		ve = append(ve, ValidationError{
			Field:  "{{.Name}}",
			Err:	fmt.Errorf("{{.Name}} does not fulfill the condition tag = {{.Tag}}. Value {{.Name}} = %v", {{.NameStruct}}.{{.Name}}),
		})	
	}`
	textFunc = `
func ({{.NamePer}} {{.NameStuct}}) Validate() ([]ValidationError, error) {
	ve := []ValidationError{}
{{.Conditions}}

	return ve, nil
}`
	textHeader = `
/*
* CODE GENERATED AUTOMATICALLY WITH go-validate
* THIS FILE SHOULD NOT BE EDITED BY HAND
* Date and time of file generation {{.DtCurrent}}
*/
//nolint:gomnd,gofmt,goimports


package {{.NamePackeg}}

import (
	"fmt"
	"regexp"
)

type ValidationError struct {
	Field string
	Err   error
}
`
)

// распарвивакм файл и записываем полученный результат в другой.
func ParserFile(nameFile string) []PE {
	pe := []PE{}
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, nameFile, nil, parser.ParseComments)
	if err != nil {
		pe = append(pe, PE{
			Message: err.Error(),
			Err:     fmt.Errorf("error parse file"),
		})

		return pe
	}
	s, errPE := parse(node)
	if len(errPE) != 0 {
		pe = append(pe, errPE...)
		var result strings.Builder
		result.WriteString("/*")
		for _, i := range pe {
			result.WriteString("\n")
			result.WriteString(i.Message)
			result.WriteString("\n")
		}
		result.WriteString("*/")

		if err = writeTextFile(nameFile, result.String()); err != nil {
			pe = append(pe, PE{
				Message: "Error writing the file",
				Err:     err,
			})
		}

		return pe
	}

	if err = writeTextFile(nameFile, s); err != nil {
		pe = append(pe, PE{
			Message: "Error writing the file",
			Err:     err,
		})
	}

	return pe
}

func parse(node *ast.File) (string, []PE) {
	pe := []PE{}
	var result strings.Builder
	t, err := getHeader(node.Name.Name)
	if err != nil {
		pe = append(pe, err...)

		return "", pe
	}
	result.WriteString(t)

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
			nameStuct := currType.Name.Name
			isHaveTag := false
			aFiled := make([]Field, len(currStruct.Fields.List))
			for i, field := range currStruct.Fields.List {
				if field.Tag != nil {
					if strings.Contains(field.Tag.Value, "validate") {
						isHaveTag = true
						isArray := false
						typeField := ""
						switch v := field.Type.(type) {
						case *ast.Ident:
							typeField = v.Name
							typeField = getTrueType(typeField, node)
						case *ast.ArrayType:
							elemType := v.Elt.(*ast.Ident).Name
							typeField = getTrueType(elemType, node)
							isArray = true
						}
						s := Field{Name: field.Names[0].Name, TypeField: typeField, Tag: field.Tag.Value, IsArray: isArray}
						aFiled[i] = s
					}
				}
			}

			if isHaveTag {
				t, err := getValidator(nameStuct, aFiled)
				pe = append(pe, err...)
				result.WriteString(t)
			}
		}
	}
	if len(pe) != 0 {
		return "", pe
	}

	return result.String(), pe
}

// Получем с генирированные условия для структуры.
func getValidator(n string, a []Field) (string, []PE) {
	pe := []PE{}
	var err error
	t := template.New("")
	buf := new(bytes.Buffer)
	var result strings.Builder

	for _, l := range a {
		if l.Name == "" {
			continue
		}
		v, err := getTag(l.Tag)
		if err != nil {
			pe = append(pe, err...)

			continue
		}
		l.Tag = v
		v = fmt.Sprintf("\n\t //NameField = %s , Tag = %s , type =%s", l.Name, l.Tag, l.TypeField)
		result.WriteString(v)
		switch l.TypeField {
		case tInt:
			fallthrough
		case "int16":
			fallthrough
		case tInt32:
			t, err := cCondInt(n, l)
			if err != nil {
				pe = append(pe, err...)

				continue
			}
			result.WriteString(t)
		case tString:
			t, err := cCondStr(n, l)
			if err != nil {
				pe = append(pe, err...)

				continue
			}
			result.WriteString(t)
		default:
			pe = append(pe, PE{
				Message: fmt.Sprintf("this type- %s is not supported. %s", l.TypeField, v),
				Err:     ErrWriteTag,
			})
		}
	}
	if len(pe) != 0 {
		return "", pe
	}
	s := struct {
		NamePer    string
		NameStuct  string
		Conditions string
	}{NamePer: strings.ToLower(n), NameStuct: n, Conditions: result.String()}

	if _, err := t.Parse(textFunc); err != nil {
		pe = append(pe, PE{
			Message: "err.Error(error when generating from a template text function)",
			Err:     fmt.Errorf("error when generating from a template text function"),
		})

		return "", pe
	}

	if err = t.Execute(buf, s); err != nil {
		pe = append(pe, PE{
			Message: "err.Error(error when generating from a template textConditions or textArrayConditions)",
			Err:     fmt.Errorf("error when generating from a template textConditions or textArrayConditions"),
		})

		return "", pe
	}

	return buf.String(), nil
}

// проверка соответствие значения для полей типа int int32 int64.
func checkIntFiled(v, ti string) bool {
	ti = strings.ToLower(ti)
	switch ti {
	case tInt:
		if _, err := strconv.Atoi(v); err != nil {
			return false
		}
	case tInt32:
		if _, err := strconv.ParseInt(v, 10, 32); err != nil {
			return false
		}
	case tInt64:
		if _, err := strconv.ParseInt(v, 10, 64); err != nil {
			return false
		}
	}

	return true
}

// получаем условие для tag in для полей типа int int32 int64.
func getCondForInInt(v, ti string) (string, error) {
	var result strings.Builder
	for _, t := range strings.Split(v, ",") {
		if !checkIntFiled(t, ti) {
			return "", ErrWriteTag
		}
		result.WriteString(fmt.Sprintf("nameField != %s && ", t))
	}
	t := result.String()

	return t[:len(t)-4], nil
}

// получаем условие для tag in для полей типа string.
func getCondForInStr(v string) string {
	var result strings.Builder
	for _, t := range strings.Split(v, ",") {
		result.WriteString(fmt.Sprintf("nameField != \"%s\" && ", t))
	}
	t := result.String()

	return t[:len(t)-4]
}

// генерим условия для полей типа: int int32 int64.
func cCondInt(n string, f Field) (string, []PE) {
	pe := []PE{}
	var err error
	n = strings.ToLower(n)
	buf := new(bytes.Buffer)
	tConditions := ""
	for _, v := range strings.Split(f.Tag, "|") {
		s := strings.Split(v, ":")

		if len(s) != 2 {
			pe = append(pe, PE{
				Message: fmt.Sprintf("The tag %s struct  %s field %s was written with an error. %s", f.Tag, n, f.Name, v),
				Err:     ErrWriteTag,
			})

			return "", pe
		}
		switch s[0] {
		case "min":

			if !checkIntFiled(s[1], f.TypeField) {
				pe = append(pe, PE{
					Message: fmt.Sprintf("The tag %s struct  %s field %s was written with an error. %s", f.Tag, n, f.Name, v),
					Err:     ErrWriteTag,
				})

				return "", pe
			}
			tConditions = fmt.Sprintf(" nameField  < %v", s[1])
		case "max":

			if !checkIntFiled(s[1], f.TypeField) {
				pe = append(pe, PE{
					Message: fmt.Sprintf("The tag %s struct  %s field %s was written with an error. %s", f.Tag, n, f.Name, v),
					Err:     ErrWriteTag,
				})

				return "", pe
			}
			tConditions = fmt.Sprintf("nameField  > %v", s[1])
		case "in":

			if tConditions, err = getCondForInInt(s[1], f.TypeField); err != nil {
				pe = append(pe, PE{
					Message: fmt.Sprintf("The tag %s struct  %s field %s was written with an error. %s", f.Tag, n, f.Name, v),
					Err:     ErrWriteTag,
				})

				return "", pe
			}
		default:
			pe = append(pe, PE{
				Message: fmt.Sprintf("The tag was written with an error. %q", v),
				Err:     ErrWriteTag,
			})

			return "", pe
		}

		st := sCon{n, f.Name, tConditions, s[1], v, f.IsArray}
		t, err := genCondition(st)

		if len(err) != 0 {
			pe = append(pe, err...)

			return "", pe
		}
		buf.WriteString(t)
	}

	return buf.String(), nil
}

// генерим условия для полей типа string.
func cCondStr(n string, f Field) (string, []PE) {
	pe := []PE{}
	n = strings.ToLower(n)
	buf := new(bytes.Buffer)
	tConditions := ""
	for _, v := range strings.Split(f.Tag, "|") {
		s := strings.Split(v, ":")

		if len(s) != 2 {
			pe = append(pe, PE{
				Message: fmt.Sprintf("The tag %s struct  %s field %s was written with an error. %s", f.Tag, n, f.Name, v),
				Err:     ErrWriteTag,
			})

			return "", pe
		}

		if strings.TrimSpace(s[1]) == "" {
			pe = append(pe, PE{
				Message: fmt.Sprintf("The tag %s struct  %s field %s was written with an error. %s", f.Tag, n, f.Name, v),
				Err:     ErrWriteTag,
			})

			return "", pe
		}
		switch s[0] {
		case "len":

			if !checkIntFiled(s[1], tInt) {
				pe = append(pe, PE{
					Message: fmt.Sprintf("The tag %s struct  %s field %s was written with an error. %s", f.Tag, n, f.Name, v),
					Err:     ErrWriteTag,
				})

				return "", pe
			}
			tConditions = fmt.Sprintf(" len(nameField)  != %v", s[1])
		case "regexp":
			tConditions = fmt.Sprintf("!regexp.MustCompile(\"%s\").MatchString(nameField)", s[1])
		case "in":
			tConditions = getCondForInStr(s[1])
		default:
			pe = append(pe, PE{
				Message: fmt.Sprintf("The tag %s struct  %s field %s was written with an error. %s", f.Tag, n, f.Name, v),
				Err:     ErrWriteTag,
			})

			return "", pe
		}
		st := sCon{n, f.Name, tConditions, s[1], v, f.IsArray}
		t, err := genCondition(st)

		if len(err) != 0 {
			pe = append(pe, err...)

			return "", pe
		}
		buf.WriteString(t)
	}

	return buf.String(), nil
}

func genCondition(st sCon) (string, []PE) {
	pe := []PE{}
	text := ""
	buf := new(bytes.Buffer)
	t := template.New("")

	if st.IsArray {
		st.Conditions = strings.ReplaceAll(st.Conditions, "nameField", "s")
		text = textArrayConditions
	} else {
		st.Conditions = strings.ReplaceAll(st.Conditions, "nameField", st.NameStruct+"."+st.Name)
		text = textConditions
	}
	if _, err := t.Parse(text); err != nil {
		pe = append(pe, PE{
			Message: "err.Error(error when generating from a template text textConditions)",
			Err:     fmt.Errorf("error when generating from a template text textConditions"),
		})

		return "", pe
	}

	if err := t.Execute(buf, st); err != nil {
		pe = append(pe, PE{
			Message: "err.Error(error when generating from a template textConditions or textArrayConditions)",
			Err:     fmt.Errorf("error when generating from a template textConditions or textArrayConditions"),
		})

		return "", pe
	}

	return buf.String(), nil
}

// берем Tag проверям его и получаем его назад в более простой форме
// напримеg validate:"min:36" json:"id" validate:"max:66=>min:36|max:66.
func getTag(tag string) (string, []PE) {
	pe := []PE{}
	var result strings.Builder

	if !strings.Contains(tag, "validate") {
		pe = append(pe, PE{
			Message: fmt.Sprintf("it does not contain key word validate. %s", tag),
			Err:     ErrWriteTag,
		})

		return "", pe
	}
	a := strings.Split(tag, " ")
	for _, t := range a {
		if strings.Contains(t, "validate") {
			t, err := pacTag(t)
			if err != nil {
				pe = append(pe, err...)

				return "", pe
			}
			result.WriteString(t)
			result.WriteString("|")
		}
	}

	return result.String()[:result.Len()-1], nil
}

// Проверка и получение Таг  нужном нам виде.
func pacTag(tag string) (string, []PE) {
	pe := []PE{}
	tag = strings.TrimSpace(tag)
	index := strings.Index(tag, "\"")
	tag = tag[index+1:]

	if index < 0 {
		pe = append(pe, PE{
			Message: fmt.Sprintf("The tag was written with an error. %s", tag),
			Err:     ErrWriteTag,
		})

		return "", pe
	}
	index = strings.LastIndex(tag, "\"")

	if index < 1 {
		pe = append(pe, PE{
			Message: fmt.Sprintf("The tag was written with an error. %s", tag),
			Err:     ErrWriteTag,
		})

		return "", pe
	}
	tag = tag[:index]

	if strings.Contains(tag, `"`) {
		pe = append(pe, PE{
			Message: fmt.Sprintf("The tag was written with an error. %s", tag),
			Err:     ErrWriteTag,
		})

		return "", pe
	}

	return strings.TrimSpace(tag), nil
}

// создаем файл и записывем полученный результат.
func writeTextFile(n, t string) error {
	vname := strings.ReplaceAll(n, filepath.Ext(n), "_validation_generated.go")
	out, err := os.Create(vname)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = out.WriteString(t); err != nil {
		return err
	}

	return nil
}

// генерим заголовок файла.
func getHeader(n string) (string, []PE) {
	pe := []PE{}
	t := template.New("")
	s := struct {
		NamePackeg string
		DtCurrent  string
	}{n, time.Now().Format("2006.01.02 15:04:05")}
	buf := new(bytes.Buffer)

	if _, err := t.Parse(textHeader); err != nil {
		pe = append(pe, PE{
			Message: "err.Error(error when generating from a template text textHeader)",
			Err:     fmt.Errorf("error when generating from a template text textHeader"),
		})

		return "", pe
	}
	if err := t.Execute(buf, s); err != nil {
		pe = append(pe, PE{
			Message: err.Error(),
			Err:     fmt.Errorf("error when generating from a template textHeader"),
		})

		return "", pe
	}

	return buf.String(), nil
}

// получаем реальный тип поля если он был объявлен как тип.
func getTrueType(ft string, node *ast.File) string {
	if ft == tInt || ft == tString {
		return ft
	}

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
			currStruct, ok := currType.Type.(*ast.Ident)

			if !ok {
				continue
			}

			if ft != currType.Name.Name {
				continue
			}
			ft = currStruct.Name

			return ft
		}
	}

	return ft
}
