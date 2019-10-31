package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/avast/retry-go"
)

var CouldNotGetDeploymentError = errors.New("Could not get deployment.")
var CouldNotUploadError = errors.New("One or more files failed to upload.")
var DeploymentReadyStateError = errors.New("Deployment ready state is set to ERROR.")
var DeploymentUrlMismatchError = errors.New("Deployment url does not match the requested deployment.")
var NoDeploymentUrlError = errors.New("No deployment URL.")
var NotAuthroizedError = errors.New("Not authorized.")

type SingleFileUploader interface {
	UploadFile(file io.Reader) (interface{}, error)
}

type Client struct {
	Token          string
	DeploymentName string
	Team           string
	Retries        uint
	HttpClient     *http.Client
}

type UploadedFile struct {
	Name string
	Sha  string
	Len  int64
}

type FileToUpload struct {
	Name   string
	Reader io.Reader
}

// Upload a file and create a deployment that contains the single file.
func (client *Client) Upload(file io.Reader, name string) (*DeployResponse, error) {
	var uploaded *UploadedFile
	if err := retry.Do(func() error {
		newUploaded, err := client.UploadFile(file, name)
		uploaded = newUploaded

		return err
	}, retry.Attempts(client.Retries)); err != nil {
		return nil, err
	}

	var deployment *DeployResponse
	if err := retry.Do(func() error {
		newDeployment, err := client.CreateDeployment([]*UploadedFile{uploaded})
		deployment = newDeployment
		return err
	}, retry.Attempts(client.Retries)); err != nil {
		return nil, err
	}

	return deployment, nil
}

// Upload many files and create a single deployment that contains the files.
func (client *Client) UploadMany(files []FileToUpload) (*DeployResponse, error) {
	c := make(chan *UploadedFile, len(files))

	for _, file := range files {
		go func(file FileToUpload, c chan *UploadedFile) {
			var retryCount uint = 0

			retry.Do(func() error {
				uploaded, err := client.UploadFile(file.Reader, file.Name)

				if err == nil {
					c <- uploaded
				}

				retryCount++

				return err
			}, retry.Attempts(client.Retries), retry.RetryIf(func(err error) bool {
				if err != nil && retryCount < client.Retries {
					return true
				}

				c <- nil

				return false
			}))
		}(file, c)
	}

	uploaded := []*UploadedFile{}

	for {
		r := <-c
		if r == nil {
			return nil, CouldNotUploadError
		}

		uploaded = append(uploaded, r)

		if len(uploaded) == len(files) {
			break
		}
	}

	var deployment *DeployResponse
	if err := retry.Do(func() error {
		newDeployment, err := client.CreateDeployment(uploaded)
		deployment = newDeployment
		return err
	}, retry.Attempts(client.Retries)); err != nil {
		return nil, err
	}

	return deployment, nil
}

// Wait for a deployment to be ready for public access.
func (client *Client) WaitForReady(deployment DeployResponse) error {
	if deployment.Url == "" {
		return NoDeploymentUrlError
	}

	for {
		if deployment.ReadyState == "READY" {
			return nil
		}

		if deployment.ReadyState == "ERROR" {
			return DeploymentReadyStateError
		}

		httpClient := client.HttpClient

		if httpClient == nil {
			httpClient = &http.Client{}
		}

		deploymentUrl := client.GetDeploymentUrl(deployment)

		request, err := http.NewRequest("GET", deploymentUrl, nil)
		if err != nil {
			return err
		}

		request.Header.Add("Authorization", client.GetBearerToken())

		res, err := httpClient.Do(request)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return CouldNotGetDeploymentError
		}

		read, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		body := DeployResponse{}
		if err := json.Unmarshal(read, &body); err != nil {
			return err
		}

		if "https://"+body.Url != deployment.Url {
			return DeploymentUrlMismatchError
		}

		deployment.ReadyState = body.ReadyState
	}
}

// Upload a file for use in a deployment.
func (client *Client) UploadFile(file io.Reader, name string) (*UploadedFile, error) {
	httpClient := client.HttpClient

	if httpClient == nil {
		httpClient = &http.Client{}
	}

	fileUrl := client.GetFileUrl()

	sha, file, err := GetSha1(file)
	if err != nil {
		return nil, err
	}

	length, file, err := Len(file)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", fileUrl, file)
	if err != nil {
		return nil, err
	}

	lengthStr := strconv.FormatInt(length, 10)
	request.Header.Add("Authorization", client.GetBearerToken())
	request.Header.Add("Content-Type", "application/octet-stream")
	request.Header.Add("Content-Length", lengthStr)
	request.Header.Add("x-now-size", lengthStr)
	request.Header.Add("x-now-digest", sha)

	res, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		read, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		body := UploadResponse{}
		if err := json.Unmarshal(read, &body); err != nil {
			return nil, err
		}

		if body.Error != nil {
			switch body.Error.Code {
			case "forbidden":
				return nil, NotAuthroizedError
			default:
				return nil, body.Error
			}
		}

		return nil, fmt.Errorf("An unknown error occured while uploading a file.")
	}

	return &UploadedFile{
		Name: name,
		Sha:  sha,
		Len:  length,
	}, nil
}

// Create a deployment that contains the uploaded files.
func (client *Client) CreateDeployment(files []*UploadedFile) (*DeployResponse, error) {
	httpClient := client.HttpClient

	if httpClient == nil {
		httpClient = &http.Client{}
	}

	deployUrl := client.GetDeployUrl()
	deploymentName := client.GetDeploymentName()

	deploymentFiles := make([]DeploymentFile, len(files))

	for index, file := range files {
		deploymentFiles[index] = DeploymentFile{
			File: file.Name,
			Sha:  file.Sha,
			Size: file.Len,
		}
	}

	requestBody, err := json.Marshal(DeploymentRequest{
		Version: 2,
		Public:  true,
		Name:    deploymentName,
		Files:   deploymentFiles,
	})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", deployUrl, bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", client.GetBearerToken())
	request.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	read, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	body := DeployResponse{}
	if err := json.Unmarshal(read, &body); err != nil {
		return nil, err
	}

	if body.Error != nil {
		return nil, body.Error
	}

	if body.Url != "" {
		body.Url = "https://" + body.Url
	}

	resultFiles := make([]DeployedFile, len(files))

	for index, file := range files {
		resultFiles[index] = DeployedFile{
			Name: file.Name,
			Sha:  file.Sha,
			Url:  body.Url + "/" + file.Name,
		}
	}

	body.Files = resultFiles

	return &body, nil
}
