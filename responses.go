package storage

import (
	"fmt"
)

type UploadResponse struct {
	Error *UploadError `json:"error"`
}

type UploadError struct {
	Code string `json:"code"`
}

func (err *UploadError) Error() string {
	return fmt.Sprintf("File upload failed, code: %s.", err.Code)
}

type DeployResponse struct {
	ID         string        `json:"id"`
	Url        string        `json:"url"`
	ReadyState string        `json:"readyState"`
	Error      *DeployErrors `json:"error"`
	Files      []DeployedFile
}

type DeployedFile struct {
	Name string
	Url  string
	Sha  string
}

type DeployErrors struct {
	Message  string        `json:"message"`
	Warnings []DeployError `json:"warning"`
	Missing  []DeployError `json:"missing"`
}

type DeployError struct {
	Message string `json:"message"`
}

func (err *DeployErrors) Error() string {
	return err.Message
}
