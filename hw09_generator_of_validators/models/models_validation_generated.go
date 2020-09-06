/*
* CODE GENERATED AUTOMATICALLY WITH go-validate
* THIS FILE SHOULD NOT BE EDITED BY HAND
* Date and time of file generation 2020.09.06 15:11:01
 */
//nolint:gomnd,gofmt,goimports

package models

import (
	"fmt"
	"regexp"
)

type ValidationError struct {
	Field string
	Err   error
}

func (user User) Validate() ([]ValidationError, error) {
	ve := []ValidationError{}

	// NameField = ID , Tag = len:36 , type =string
	if len(user.ID) != 36 {
		ve = append(ve, ValidationError{
			Field: "ID",
			Err:   fmt.Errorf("ID does not fulfill the condition tag = len:36. Value ID = %v", user.ID),
		})
	}
	// NameField = Age , Tag = min:18|max:50 , type =int
	if user.Age < 18 {
		ve = append(ve, ValidationError{
			Field: "Age",
			Err:   fmt.Errorf("Age does not fulfill the condition tag = min:18. Value Age = %v", user.Age),
		})
	}
	if user.Age > 50 {
		ve = append(ve, ValidationError{
			Field: "Age",
			Err:   fmt.Errorf("Age does not fulfill the condition tag = max:50. Value Age = %v", user.Age),
		})
	}
	// NameField = Email , Tag = regexp:^\\w+@\\w+\\.\\w+$ , type =string
	if !regexp.MustCompile("^\\w+@\\w+\\.\\w+$").MatchString(user.Email) {
		ve = append(ve, ValidationError{
			Field: "Email",
			Err:   fmt.Errorf("Email does not fulfill the condition tag = regexp:^\\w+@\\w+\\.\\w+$. Value Email = %v", user.Email),
		})
	}
	// NameField = Role , Tag = in:admin,stuff , type =string
	if user.Role != "admin" && user.Role != "stuff" {
		ve = append(ve, ValidationError{
			Field: "Role",
			Err:   fmt.Errorf("Role does not fulfill the condition tag = in:admin,stuff. Value Role = %v", user.Role),
		})
	}
	// NameField = Phones , Tag = len:11 , type =string
	for _, s := range user.Phones {
		if len(s) != 11 {
			ve = append(ve, ValidationError{
				Field: "Phones",
				Err:   fmt.Errorf("Phones does not fulfill the condition tag = len:11. Value Phones = %v", s),
			})
		}
	}

	return ve, nil
}

func (app App) Validate() ([]ValidationError, error) {
	ve := []ValidationError{}

	// NameField = Version , Tag = len:5 , type =string
	if len(app.Version) != 5 {
		ve = append(ve, ValidationError{
			Field: "Version",
			Err:   fmt.Errorf("Version does not fulfill the condition tag = len:5. Value Version = %v", app.Version),
		})
	}

	return ve, nil
}

func (response Response) Validate() ([]ValidationError, error) {
	ve := []ValidationError{}

	// NameField = Code , Tag = in:200,404,500 , type =int
	if response.Code != 200 && response.Code != 404 && response.Code != 500 {
		ve = append(ve, ValidationError{
			Field: "Code",
			Err:   fmt.Errorf("Code does not fulfill the condition tag = in:200,404,500. Value Code = %v", response.Code),
		})
	}

	return ve, nil
}
