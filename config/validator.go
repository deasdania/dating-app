package config

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

type BaseValidator struct {
	Validator *validator.Validate
}

func NewValidator() *BaseValidator {
	bv := BaseValidator{
		Validator: validator.New(),
	}
	bv.Validator.RegisterValidation("alphanumwithspace", bv.ValidateAlphanumWithSpace)
	bv.Validator.RegisterValidation("fileAttachment", bv.ValidateFileAttachment)
	return &bv
}

func (b *BaseValidator) Validate(i interface{}) error {
	if err := b.Validator.Struct(i); err != nil {
		// return err
		// Return only the first validation error
		for _, err := range err.(validator.ValidationErrors) {
			return fmt.Errorf("%s", err.Field())
		}
	}
	return nil
}

func (b *BaseValidator) ValidateAlphanumWithSpace(fl validator.FieldLevel) bool {
	pattern := "^[a-zA-Z0-9 ]+$"
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(fl.Field().String())
}

func (b *BaseValidator) ValidateFileAttachment(fl validator.FieldLevel) bool {
	field := fl.Field().Interface().([]*multipart.FileHeader)

	for _, v := range field {
		file, err := v.Open()
		if err != nil {
			return false
		}
		defer file.Close()

		buffer := make([]byte, 512)
		if _, err := file.Read(buffer); err != nil {
			return false
		}

		contentType := http.DetectContentType(buffer)
		switch {
		case strings.HasPrefix(contentType, "image/jpeg"):
			continue
		case strings.HasPrefix(contentType, "image/png"):
			continue
		case strings.HasPrefix(contentType, "image/jpg"):
			continue
		case strings.HasPrefix(contentType, "application/pdf"):
			continue
		default:
			return false
		}
	}
	return true
}
