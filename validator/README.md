# Validator

The validator package is a wrapper on [go-playground/validator](https://github.com/go-playground/validator) for validation.

## Usage

```go
package main

import (
    "fmt"

    "github.com/digitalmonsters/go-common/validator"
)

type User struct {
    Name string `validate:"required"`
    Age  int    `validate:"gte=0,lte=130"`
    Email string `validate:"required,email"`
}

func main() {
    validator.Init(&validator.Option{})
    
    user := User{
        Name: "Badger",
        Age:  5,
        Email: "",
    }

    err := validator.Validate(user) 
    if err != nil {
        fmt.Println(err)
    }
}
```

## Custom Message Formatter

```go
package main

import (
    "fmt"

    "github.com/digitalmonsters/go-common/validator"
)

type User struct {
    Name string `validate:"required"`
    Age  int    `validate:"gte=0,lte=130"`
    Email string `validate:"required,email"`
}

func main() {
    validator.Init(&validator.Option{
        MessageFormatter: validator.MessageFormatter{
            "required": func(err *FieldError) string {
                return fmt.Sprintf("%s can not be empty.", err.FieldName)
            },
        },
    })
   
    user := User{
        Name: "Badger",
        Age:  5,
        Email: "",
    }

    err := validator.Validate(user) 
    if err != nil {
        fmt.Println(err)
    }
}
```
