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

// ErrWriteTag error return model.
var (
	ErrWriteTag          = errors.New("the tag was written with an error")
	geneateFileWithError = true
)

const (
	tInt    = "int"
	tInt32  = "int32"
	tInt64  = "int64"
	tString = "string"
)

// ParsErr it's struct from return.
type ParsErr struct {
	Message string
	Err     error
}

// StCon Struct from generete template.
type StCon struct {
	NameStruct string
	Name       string
	Conditions string
	Value      string
	Tag        string
	IsArray    bool
}

// StField Struct from generete template.
type StField struct {
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
	textError = `
	ve = append(ve, ValidationError{
		Field:  "{{.Name}} does not generate the condition tag = {{.Tag}}. Field = {{.NameStruct}}.{{.Name}}",
		Err:	fmt.Errorf("error with codegeneration"),
	})`
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

// ParserFile распарвивакм файл и записываем полученный результат в другой.
func ParserFile(nameFile string) []ParsErr {
	pe := []ParsErr{}
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, nameFile, nil, parser.ParseComments)
	if err != nil {
		pe = append(pe, ParsErr{
			Message: err.Error(),
			Err:     fmt.Errorf("error parse file"),
		})

		return pe
	}
	s, errPE := parse(node)
	if len(errPE) != 0 && !geneateFileWithError {
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
			pe = append(pe, ParsErr{
				Message: "Error writing the file",
				Err:     err,
			})
		}

		return pe
	}

	if err = writeTextFile(nameFile, s); err != nil {
		pe = append(pe, ParsErr{
			Message: "Error writing the file",
			Err:     err,
		})
	}

	return pe
}

func parse(node *ast.File) (string, []ParsErr) {
	pe := []ParsErr{}
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
			aFiled := make([]StField, len(currStruct.Fields.List))
			for i, field := range currStruct.Fields.List {
				if field.Tag != nil {
					if strings.Contains(field.Tag.Value, "validate") {
						typeField, isArray := getType(field, node)
						aFiled[i] = StField{Name: field.Names[0].Name, TypeField: typeField, Tag: field.Tag.Value, IsArray: isArray}
					}
				}
			}

			if len(aFiled) > 0 {
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

//
func getType(field *ast.Field, node *ast.File) (string, bool) {
	typeField := ""
	isArray := false
	if v, ok := field.Type.(*ast.Ident); ok {
		typeField = v.Name
	}
	switch v := field.Type.(type) {
	case *ast.Ident:
		typeField = getTrueType(typeField, node)
	case *ast.ArrayType:
		elemType := v.Elt.(*ast.Ident).Name
		typeField = getTrueType(elemType, node)
		isArray = true
	}

	return typeField, isArray
}

// Получем сгенирированные условия для структуры.
func getValidator(nameStruct string, a []StField) (string, []ParsErr) {
	pe := []ParsErr{}
	var result strings.Builder
	var t string
	nameValue := strings.ToLower(nameStruct)
	if nameStruct == nameValue {
		nameValue = "_" + nameValue
	}
	for _, l := range a {
		var err []ParsErr
		if l.Name == "" {
			continue
		}
		l.Tag, err = getTag(l.Tag)
		if err != nil {
			pe = append(pe, err...)

			continue
		}
		v := fmt.Sprintf("\n\t //NameField = %s , Tag = %s , type =%s", l.Name, l.Tag, l.TypeField)
		result.WriteString(v)
		switch l.TypeField {
		case tInt:
			fallthrough
		case tInt32:
			fallthrough
		case tInt64:
			t, err = createCondInt(nameValue, l)
		case tString:
			t, err = createCondStr(nameValue, l)
		default:
			err = append(err, ParsErr{
				Message: fmt.Sprintf("this type- %s is not supported. %s", l.TypeField, v),
				Err:     ErrWriteTag,
			})
		}
		if err != nil {
			if geneateFileWithError {
				t, err := genError(nameStruct, l)
				pe = append(pe, err...)
				result.WriteString(t)
			}
			pe = append(pe, err...)
		} else {
			result.WriteString(t)
		}
	}
	if len(pe) != 0 && !geneateFileWithError {
		return "", pe
	}
	t, err := generateFromCond(nameStruct, textFunc, result)
	if len(pe) != 0 && !geneateFileWithError {
		pe = append(pe, err...)

		return "", pe
	}

	return t, nil
}

// Генерим из шаблона условия.
func generateFromCond(nameStruct, textFunc string, result strings.Builder) (string, []ParsErr) {
	pe := []ParsErr{}
	if result.String() == "" {
		return "", pe
	}
	t := template.New("")
	buf := new(bytes.Buffer)
	nameVal := strings.ToLower(nameStruct)
	if nameVal == nameStruct {
		nameVal = "_" + nameVal
	}
	s := struct {
		NamePer    string
		NameStuct  string
		Conditions string
	}{NamePer: nameVal, NameStuct: nameStruct, Conditions: result.String()}

	if _, err := t.Parse(textFunc); err != nil {
		pe = append(pe, ParsErr{
			Message: "err.Error(error when generating from a template text function)",
			Err:     fmt.Errorf("error when generating from a template text function"),
		})

		return "", pe
	}

	if err := t.Execute(buf, s); err != nil {
		pe = append(pe, ParsErr{
			Message: "err.Error(error when generating from a template textConditions or textArrayConditions)",
			Err:     fmt.Errorf("error when generating from a template textConditions or textArrayConditions"),
		})

		return "", pe
	}

	return buf.String(), pe
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
func createCondInt(nameStruct string, f StField) (string, []ParsErr) {
	pe := []ParsErr{}
	var err error
	buf := new(bytes.Buffer)
	tConditions := ""

	for _, v := range strings.Split(f.Tag, "|") {
		s := strings.Split(v, ":")
		if len(s) != 2 {
			pe = append(pe, ParsErr{
				Message: fmt.Sprintf("the tag %s struct  %s field %s was written with an error. %s", f.Tag, nameStruct, f.Name, v),
				Err:     ErrWriteTag,
			})

			return "", pe
		}
		switch s[0] {
		case "min":

			if !checkIntFiled(s[1], f.TypeField) {
				err = fmt.Errorf("the tag %s struct  %s field %s was written with an error. %s", f.Tag, nameStruct, f.Name, v)
			}
			tConditions = fmt.Sprintf(" nameField  < %v", s[1])
		case "max":

			if !checkIntFiled(s[1], f.TypeField) {
				err = fmt.Errorf("the tag %s struct  %s field %s was written with an error. %s", f.Tag, nameStruct, f.Name, v)
			}

			tConditions = fmt.Sprintf("nameField  > %v", s[1])
		case "in":
			tConditions, err = getCondForInInt(s[1], f.TypeField)
		default:

			err = fmt.Errorf("the tag was written with an error. %q", v)
		}

		if err != nil {
			pe = append(pe, ParsErr{
				Message: fmt.Sprintf("The tag %s struct  %s field %s was written with an error. %s", f.Tag, nameStruct, f.Name, v),
				Err:     ErrWriteTag,
			})

			return "", pe
		}

		st := StCon{nameStruct, f.Name, tConditions, s[1], v, f.IsArray}
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
func createCondStr(nameStruct string, f StField) (string, []ParsErr) {
	pe := []ParsErr{}
	buf := new(bytes.Buffer)
	tConditions := ""
	for _, v := range strings.Split(f.Tag, "|") {
		s := strings.Split(v, ":")
		if len(s) != 2 {
			pe = append(pe, ParsErr{
				Message: fmt.Sprintf("The tag %s struct  %s field %s was written with an error. %s", f.Tag, nameStruct, f.Name, v),
				Err:     ErrWriteTag,
			})

			return "", pe
		}

		if strings.TrimSpace(s[1]) == "" {
			pe = append(pe, ParsErr{
				Message: fmt.Sprintf("The tag %s struct  %s field %s was written with an error. %s", f.Tag, nameStruct, f.Name, v),
				Err:     ErrWriteTag,
			})

			return "", pe
		}
		switch s[0] {
		case "len":
			if !checkIntFiled(s[1], tInt) {
				pe = append(pe, ParsErr{
					Message: fmt.Sprintf("The tag %s struct  %s field %s was written with an error. %s", f.Tag, nameStruct, f.Name, v),
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
			pe = append(pe, ParsErr{
				Message: fmt.Sprintf("The tag %s struct  %s field %s was written with an error. %s", f.Tag, nameStruct, f.Name, v),
				Err:     ErrWriteTag,
			})

			return "", pe
		}
		st := StCon{nameStruct, f.Name, tConditions, s[1], v, f.IsArray}
		t, err := genCondition(st)
		if len(err) != 0 {
			pe = append(pe, err...)

			return "", pe
		}
		buf.WriteString(t)
	}

	return buf.String(), nil
}

//  генирим шаблоны
func genCondition(st StCon) (string, []ParsErr) {
	pe := []ParsErr{}
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
		pe = append(pe, ParsErr{
			Message: "err.Error(error when generating from a template text textConditions)",
			Err:     fmt.Errorf("error when generating from a template text textConditions"),
		})

		return "", pe
	}
	if err := t.Execute(buf, st); err != nil {
		pe = append(pe, ParsErr{
			Message: "err.Error(error when generating from a template textConditions or textArrayConditions)",
			Err:     fmt.Errorf("error when generating from a template textConditions or textArrayConditions"),
		})

		return "", pe
	}

	return buf.String(), nil
}

//  генирим шаблоны еггог
func genError(nameStruct string, f StField) (string, []ParsErr) {
	buf := new(bytes.Buffer)
	pe := []ParsErr{}
	st := StCon{nameStruct, f.Name, "", f.Tag, f.Tag, f.IsArray}
	t := template.New("")
	/*
		v := fmt.Sprintf("\n\t //NameField = %s , Tag = %s , type =%s", l.Name, l.Tag, l.TypeField)
		result.WriteString(v)
	*/
	if _, err := t.Parse(textError); err != nil {
		pe = append(pe, ParsErr{
			Message: "err.Error(error when generating from a template text textConditions)",
			Err:     fmt.Errorf("error when generating from a template text textConditions"),
		})

		return "", pe
	}
	if err := t.Execute(buf, st); err != nil {
		pe = append(pe, ParsErr{
			Message: "err.Error(error when generating from a template textConditions or textArrayConditions)",
			Err:     fmt.Errorf("error when generating from a template textConditions or textArrayConditions"),
		})

		return "", pe
	}

	return buf.String(), nil
}

// берем Tag проверям его и получаем его назад в более простой форме
// напримеg validate:"min:36" json:"id" validate:"max:66=>min:36|max:66.
func getTag(tag string) (string, []ParsErr) {
	pe := []ParsErr{}
	var result strings.Builder

	if !strings.Contains(tag, "validate") {
		pe = append(pe, ParsErr{
			Message: fmt.Sprintf("it does not contain key word validate. %s", tag),
			Err:     ErrWriteTag,
		})

		return "", pe
	}
	a := strings.Split(tag, " ")
	for _, t := range a {
		if strings.Contains(t, "validate") {
			t, err := checkTag(t)
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
func checkTag(tag string) (string, []ParsErr) {
	pe := []ParsErr{}
	tag = strings.TrimSpace(tag)
	index := strings.Index(tag, "\"")
	tag = tag[index+1:]

	if index < 0 {
		pe = append(pe, ParsErr{
			Message: fmt.Sprintf("The tag was written with an error. %s", tag),
			Err:     ErrWriteTag,
		})

		return "", pe
	}
	index = strings.LastIndex(tag, "\"")

	if index < 1 {
		pe = append(pe, ParsErr{
			Message: fmt.Sprintf("The tag was written with an error. %s", tag),
			Err:     ErrWriteTag,
		})

		return "", pe
	}
	tag = tag[:index]

	if strings.Contains(tag, `"`) {
		pe = append(pe, ParsErr{
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
func getHeader(n string) (string, []ParsErr) {
	pe := []ParsErr{}
	t := template.New("")
	s := struct {
		NamePackeg string
		DtCurrent  string
	}{n, time.Now().Format("2006.01.02 15:04:05")}
	buf := new(bytes.Buffer)

	if _, err := t.Parse(textHeader); err != nil {
		pe = append(pe, ParsErr{
			Message: "err.Error(error when generating from a template text textHeader)",
			Err:     fmt.Errorf("error when generating from a template text textHeader"),
		})

		return "", pe
	}
	if err := t.Execute(buf, s); err != nil {
		pe = append(pe, ParsErr{
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
