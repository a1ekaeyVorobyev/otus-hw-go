package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTag(t *testing.T) {
	t.Run("Get Tag", func(t *testing.T) {
		atag := []string{
			`json:"id" validate:"len:36"`,
			`json:"id" validate:"min:36|max:56"`,
			`validate:"min:36" json:"id" validate:"max:66"`,
		}
		atagAfter := []string{
			`len:36`,
			`min:36|max:56`,
			`min:36|max:66`,
		}
		for i, tag := range atag {
			tagExecute, err := getTag(tag)
			require.Empty(t, err)
			require.Equal(t, tagExecute, atagAfter[i])
		}
	})
	t.Run("Not tag", func(t *testing.T) {
		tag := `json:"id"`
		tagExecute, err := getTag(tag)
		for _, e := range err {
			require.Equal(t, e.Err, ErrWriteTag)
		}
		require.Equal(t, tagExecute, "")
	})
	t.Run("Eorror Write tag", func(t *testing.T) {
		atag := []string{
			`json:"id" validate:"len:26`,
			`json:"id" validate:len:36`,
			`json:"id" validate:""`,
			`json:"id" validate:"len:46"""`,
		}
		for _, tag := range atag {
			tagExecute, err := getTag(tag)
			for _, e := range err {
				require.Equal(t, e.Err, ErrWriteTag)
			}
			require.Equal(t, tagExecute, "")
		}
	})
}

func TestCreateConditionsString(t *testing.T) {
	nameStruct := "user"
	t.Run("Get Tag Array", func(t *testing.T) {
		f := StField{
			Name:      "Age",
			TypeField: "string",
			Tag:       "len:36",
			IsArray:   true,
		}
		s, err := createCondStr(nameStruct, f)
		text := `
	for _, s := range user.Age {
		if  len(s)  != 36 {
			ve = append(ve, ValidationError{
				Field:  "Age",
				Err:	fmt.Errorf("Age does not fulfill the condition tag = len:36. Value Age = %v", s),
			})
		}
	}`
		require.Empty(t, err)
		require.Equal(t, s, text)
	})
	t.Run("Get Tag len", func(t *testing.T) {
		f := StField{
			Name:      "Age",
			TypeField: "string",
			Tag:       "len:36",
			IsArray:   false,
		}
		s, err := createCondStr(nameStruct, f)
		text := `
	if  len(user.Age)  != 36 {
		ve = append(ve, ValidationError{
			Field:  "Age",
			Err:	fmt.Errorf("Age does not fulfill the condition tag = len:36. Value Age = %v", user.Age),
		})	
	}`
		require.Empty(t, err)
		require.Equal(t, s, text)
	})
	t.Run("Get Tag in", func(t *testing.T) {
		f := StField{
			Name:      "Age",
			TypeField: "string",
			Tag:       "in:foo,bar",
			IsArray:   false,
		}
		s, err := createCondStr(nameStruct, f)
		text := `
	if user.Age != "foo" && user.Age != "bar" {
		ve = append(ve, ValidationError{
			Field:  "Age",
			Err:	fmt.Errorf("Age does not fulfill the condition tag = in:foo,bar. Value Age = %v", user.Age),
		})	
	}`
		require.Empty(t, err)
		require.Equal(t, s, text)
	})
	t.Run("Get Tag regex", func(t *testing.T) {
		f := StField{
			Name:      "Age",
			TypeField: "string",
			Tag:       `regexp:^\\w+@\\w+\\.\\w+$`,
			IsArray:   false,
		}
		s, err := createCondStr(nameStruct, f)
		text := `
	if !regexp.MustCompile("^\\w+@\\w+\\.\\w+$").MatchString(user.Age) {
		ve = append(ve, ValidationError{
			Field:  "Age",
			Err:	fmt.Errorf("Age does not fulfill the condition tag = regexp:^\\w+@\\w+\\.\\w+$. Value Age = %v", user.Age),
		})	
	}`
		require.Empty(t, err)
		require.Equal(t, s, text)
	})
	t.Run("Error Get Tag", func(t *testing.T) {
		f := StField{
			Name:      "Age",
			TypeField: "string",
			Tag:       "len:36d",
			IsArray:   false,
		}
		s, err := createCondStr("User", f)
		for _, e := range err {
			require.Equal(t, e.Err, ErrWriteTag)
		}
		require.Equal(t, s, "")
	})
}

func TestCreateConditionsInt(t *testing.T) {
	nameStruct := "user"
	t.Run("Get Tag Array", func(t *testing.T) {
		f := StField{
			Name:      "Age",
			TypeField: "int",
			Tag:       "min:18",
			IsArray:   true,
		}
		s, err := createCondInt(nameStruct, f)
		text := `
	for _, s := range user.Age {
		if  s  < 18 {
			ve = append(ve, ValidationError{
				Field:  "Age",
				Err:	fmt.Errorf("Age does not fulfill the condition tag = min:18. Value Age = %v", s),
			})
		}
	}`
		require.Empty(t, err)
		require.Equal(t, s, text)
	})
	t.Run("Get Tag min", func(t *testing.T) {
		f := StField{
			Name:      "Age",
			TypeField: "int",
			Tag:       "min:18",
			IsArray:   false,
		}
		s, err := createCondInt(nameStruct, f)
		text := `
	if  user.Age  < 18 {
		ve = append(ve, ValidationError{
			Field:  "Age",
			Err:	fmt.Errorf("Age does not fulfill the condition tag = min:18. Value Age = %v", user.Age),
		})	
	}`
		require.Empty(t, err)
		require.Equal(t, s, text)
	})
	t.Run("Get Tag max", func(t *testing.T) {
		f := StField{
			Name:      "Age",
			TypeField: "int",
			Tag:       "max:90",
			IsArray:   false,
		}
		s, err := createCondInt(nameStruct, f)
		text := `
	if user.Age  > 90 {
		ve = append(ve, ValidationError{
			Field:  "Age",
			Err:	fmt.Errorf("Age does not fulfill the condition tag = max:90. Value Age = %v", user.Age),
		})	
	}`
		require.Empty(t, err)
		require.Equal(t, s, text)
	})
	t.Run("Get Tag In", func(t *testing.T) {
		f := StField{
			Name:      "Age",
			TypeField: "int",
			Tag:       "in:200,404,500",
			IsArray:   false,
		}
		s, err := createCondInt(nameStruct, f)
		text := `
	if user.Age != 200 && user.Age != 404 && user.Age != 500 {
		ve = append(ve, ValidationError{
			Field:  "Age",
			Err:	fmt.Errorf("Age does not fulfill the condition tag = in:200,404,500. Value Age = %v", user.Age),
		})	
	}`
		require.Empty(t, err)
		require.Equal(t, s, text)
	})
	t.Run("Get complex Tag", func(t *testing.T) {
		f := StField{
			Name:      "Age",
			TypeField: "int",
			Tag:       "in:200,404,500|min:18",
			IsArray:   false,
		}
		s, err := createCondInt(nameStruct, f)
		text := `
	if user.Age != 200 && user.Age != 404 && user.Age != 500 {
		ve = append(ve, ValidationError{
			Field:  "Age",
			Err:	fmt.Errorf("Age does not fulfill the condition tag = in:200,404,500. Value Age = %v", user.Age),
		})	
	}
	if  user.Age  < 18 {
		ve = append(ve, ValidationError{
			Field:  "Age",
			Err:	fmt.Errorf("Age does not fulfill the condition tag = min:18. Value Age = %v", user.Age),
		})	
	}`
		require.Empty(t, err)
		require.Equal(t, s, text)
	})
	t.Run("Error complex Tag", func(t *testing.T) {
		f := StField{
			Name:      "Age",
			TypeField: "int",
			Tag:       "in:200,404,500;min:18",
			IsArray:   false,
		}
		s, err := createCondStr(nameStruct, f)
		for _, e := range err {
			require.Equal(t, e.Err, ErrWriteTag)
		}
		require.Equal(t, s, "")
	})
	t.Run("Error", func(t *testing.T) {
		f := StField{
			Name:      "Age",
			TypeField: "int",
			Tag:       "min",
			IsArray:   true,
		}
		s, err := createCondStr(nameStruct, f)
		for _, e := range err {
			require.Equal(t, e.Err, ErrWriteTag)
		}
		require.Equal(t, s, "")
		f.Tag = "man:18"
		s, err = createCondStr(nameStruct, f)
		for _, e := range err {
			require.Equal(t, e.Err, ErrWriteTag)
		}
		require.Equal(t, s, "")
		f.Tag = "max:2147483650"
		f.TypeField = "int32"
		s, err = createCondStr(nameStruct, f)
		for _, e := range err {
			require.Equal(t, e.Err, ErrWriteTag)
		}
		require.Equal(t, s, "")
		f.Tag = "max:12a"
		s, err = createCondStr(nameStruct, f)
		for _, e := range err {
			require.Equal(t, e.Err, ErrWriteTag)
		}
		require.Equal(t, s, "")
	})
}

func TestGetConditionsForInInt(t *testing.T) {
	t.Run("get in tag from int", func(t *testing.T) {
		in := "12,36,56"
		out := "nameField != 12 && nameField != 36 && nameField != 56"
		s, err := getCondForInInt(in, "int")
		require.Empty(t, err)
		require.Equal(t, s, out)
	})
	t.Run("error get in tag from int", func(t *testing.T) {
		in := "12,,36,56"
		s, err := getCondForInInt(in, "int")
		require.Equal(t, err, ErrWriteTag)
		require.Equal(t, s, "")
		in = "12,fgfdg,36,56"
		s, err = getCondForInInt(in, "int")
		require.Equal(t, err, ErrWriteTag)
		require.Equal(t, s, "")
		in = "12:36:56"
		s, err = getCondForInInt(in, "int")
		require.Equal(t, err, ErrWriteTag)
		require.Equal(t, s, "")
	})
}

func TestVAlidator(t *testing.T) {
	t.Run("Check Validator ", func(t *testing.T) {
		f := []StField{
			{
				Name:      "Name",
				TypeField: "string",
				Tag:       `validate:"len:36"`,
				IsArray:   true,
			}, {
				Name:      "Age",
				TypeField: "int",
				Tag:       `validate:"min:18"`,
				IsArray:   true,
			},
		}
		text := `
func (user User) Validate() ([]ValidationError, error) {
	ve := []ValidationError{}

	 //NameField = Name , Tag = len:36 , type =string
	for _, s := range user.Name {
		if  len(s)  != 36 {
			ve = append(ve, ValidationError{
				Field:  "Name",
				Err:	fmt.Errorf("Name does not fulfill the condition tag = len:36. Value Name = %v", s),
			})
		}
	}
	 //NameField = Age , Tag = min:18 , type =int
	for _, s := range user.Age {
		if  s  < 18 {
			ve = append(ve, ValidationError{
				Field:  "Age",
				Err:	fmt.Errorf("Age does not fulfill the condition tag = min:18. Value Age = %v", s),
			})
		}
	}

	return ve, nil
}`
		s, err := getValidator("User", f)
		for _, e := range err {
			require.Equal(t, e.Err, ErrWriteTag)
		}
		require.Equal(t, s, text)
	})
	t.Run("Check Validator when nameStruct have lower case", func(t *testing.T) {
		f := []StField{
			{
				Name:      "Name",
				TypeField: "string",
				Tag:       `validate:"len:36"`,
				IsArray:   true,
			}, {
				Name:      "Age",
				TypeField: "int",
				Tag:       `validate:"min:18"`,
				IsArray:   true,
			},
		}
		text := `
func (_user user) Validate() ([]ValidationError, error) {
	ve := []ValidationError{}

	 //NameField = Name , Tag = len:36 , type =string
	for _, s := range _user.Name {
		if  len(s)  != 36 {
			ve = append(ve, ValidationError{
				Field:  "Name",
				Err:	fmt.Errorf("Name does not fulfill the condition tag = len:36. Value Name = %v", s),
			})
		}
	}
	 //NameField = Age , Tag = min:18 , type =int
	for _, s := range _user.Age {
		if  s  < 18 {
			ve = append(ve, ValidationError{
				Field:  "Age",
				Err:	fmt.Errorf("Age does not fulfill the condition tag = min:18. Value Age = %v", s),
			})
		}
	}

	return ve, nil
}`
		s, err := getValidator("user", f)
		for _, e := range err {
			require.Equal(t, e.Err, ErrWriteTag)
		}
		require.Equal(t, s, text)
	})
	t.Run("Check Validator have error geneateFileWithError = false", func(t *testing.T) {
		f := []StField{
			{
				Name:      "Name",
				TypeField: "string",
				Tag:       `validate:"len:36"`,
				IsArray:   true,
			}, {
				Name:      "Age",
				TypeField: "int",
				Tag:       `validate:"midn:18"`,
				IsArray:   true,
			},
		}
		text := ""
		geneateFileWithError = false
		s, err := getValidator("user", f)
		for _, e := range err {
			require.Equal(t, e.Err, ErrWriteTag)
		}

		require.Equal(t, s, text)
	})

	t.Run("Check Validator have error geneateFileWithError = true", func(t *testing.T) {
		f := []StField{
			{
				Name:      "Name",
				TypeField: "string",
				Tag:       `validate:"len:36"`,
				IsArray:   true,
			}, {
				Name:      "Age",
				TypeField: "int",
				Tag:       `validate:"midn:18"`,
				IsArray:   true,
			},
		}
		geneateFileWithError = true
		text := `
func (user User) Validate() ([]ValidationError, error) {
	ve := []ValidationError{}

	 //NameField = Name , Tag = len:36 , type =string
	for _, s := range user.Name {
		if  len(s)  != 36 {
			ve = append(ve, ValidationError{
				Field:  "Name",
				Err:	fmt.Errorf("Name does not fulfill the condition tag = len:36. Value Name = %v", s),
			})
		}
	}
	 //NameField = Age , Tag = midn:18 , type =int
	ve = append(ve, ValidationError{
		Field:  "Age does not generate the condition tag = midn:18. Field = User.Age",
		Err:	fmt.Errorf("error with codegeneration"),
	})

	return ve, nil
}`
		s, err := getValidator("User", f)
		for _, e := range err {
			require.Equal(t, e.Err, ErrWriteTag)
		}
		require.Equal(t, s, text)
	})
}
