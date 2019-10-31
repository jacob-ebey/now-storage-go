package storage

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestUpload(t *testing.T) {
	testFile := "test.txt"

	if err := godotenv.Load(".env"); err != nil {
		t.Error("Failed to load dotenv file.")
	}

	client := Client{
		Token:          os.Getenv("NOW_TOKEN"),
		DeploymentName: os.Getenv("NOW_DEPLOYMENT"),
		Team:           os.Getenv("NOW_TEAMID"),
		Retries:        2,
	}

	f, err := os.Open(testFile)
	if err != nil {
		t.Error("Failed to open test file to upload.", err)
	}
	defer f.Close()
	ogBody, err := ioutil.ReadAll(f)

	deployment, err := client.Upload(bytes.NewReader(ogBody), testFile)
	if err != nil {
		t.Error("Failed to deploy file.", err)
	}

	if deployment == nil {
		t.Error("Deployment is nil without an error.")
	}

	if deployment.ID == "" {
		t.Error("Deployment does not contain an ID.")
	}

	if deployment.Url == "" {
		t.Error("Deployment does not contain a URL.")
	}

	if len(deployment.Files) != 1 {
		t.Error("Files length mismatch.")
	}

	client.WaitForReady(*deployment)

	res, err := http.Get(deployment.Files[0].Url)
	if err != nil {
		t.Error("Could not get uploaded file.", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error("Could not read uploaded file response.", err)
	}

	if err != nil {
		t.Error("Could not read file to compare to response.", err)
	}

	if string(body) != string(ogBody) {
		t.Error("Uploaded file response and origional file are not equal.")
	}
}

func TestUploadMany(t *testing.T) {
	testFile := "test.txt"
	testFile2 := "test2.txt"

	if err := godotenv.Load(".env"); err != nil {
		t.Error("Failed to load dotenv file.")
	}

	client := Client{
		Token:          os.Getenv("NOW_TOKEN"),
		DeploymentName: os.Getenv("NOW_DEPLOYMENT"),
		Team:           os.Getenv("NOW_TEAMID"),
		Retries:        2,
	}

	f, err := os.Open(testFile)
	if err != nil {
		t.Error("Failed to open test file to upload.", err)
	}
	defer f.Close()
	ogBody, err := ioutil.ReadAll(f)
	f2, err := os.Open(testFile2)
	if err != nil {
		t.Error("Failed to open test file 2 to upload.", err)
	}
	defer f2.Close()
	ogBody2, err := ioutil.ReadAll(f2)

	deployment, err := client.UploadMany([]FileToUpload{
		FileToUpload{
			Name:   testFile,
			Reader: bytes.NewReader(ogBody),
		},
		FileToUpload{
			Name:   testFile2,
			Reader: bytes.NewReader(ogBody2),
		},
	})
	if err != nil {
		t.Error("Failed to upload file.", err)
	}

	if err != nil {
		t.Errorf("Failed to create deployment, %v", err)
	}

	if deployment == nil {
		t.Error("Deployment is nil without an error.")
	}

	if deployment.ID == "" {
		t.Error("Deployment does not contain an ID.")
	}

	if deployment.Url == "" {
		t.Error("Deployment does not contain a URL.")
	}

	if len(deployment.Files) != 2 {
		t.Error("Files length mismatch.")
	}

	client.WaitForReady(*deployment)

	res, err := http.Get(deployment.Url + "/" + testFile)
	if err != nil {
		t.Error("Could not get uploaded file.", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error("Could not read uploaded file response.", err)
	}

	if err != nil {
		t.Error("Could not read file to compare to response.", err)
	}

	if string(body) != string(ogBody) {
		t.Error("Uploaded file response and origional file are not equal.")
	}

	res2, err := http.Get(deployment.Url + "/" + testFile2)
	if err != nil {
		t.Error("Could not get uploaded file.", err)
	}
	defer res2.Body.Close()

	body2, err := ioutil.ReadAll(res2.Body)
	if err != nil {
		t.Error("Could not read uploaded file response.", err)
	}

	if err != nil {
		t.Error("Could not read file to compare to response.", err)
	}

	if string(body2) != string(ogBody2) {
		t.Error("Uploaded file response and origional file are not equal.")
	}
}

func TestUploadFile(t *testing.T) {
	testFile := "test.txt"
	testFileSha := "16192dbe587d3d3556543bcc6401e2f0dbf06b69"
	testFileLen := int64(4)

	if err := godotenv.Load(".env"); err != nil {
		t.Error("Failed to load dotenv file.")
	}

	client := Client{
		Token:          os.Getenv("NOW_TOKEN"),
		DeploymentName: os.Getenv("NOW_DEPLOYMENT"),
		Team:           os.Getenv("NOW_TEAMID"),
		Retries:        2,
	}

	f, err := os.Open(testFile)
	if err != nil {
		t.Error("Failed to open test file to upload.", err)
	}
	defer f.Close()

	uploaded, err := client.UploadFile(f, testFile)
	if err != nil {
		t.Error("Failed to upload file.", err)
	}

	if uploaded == nil {
		t.Error("No uploaded file returned.")
	}

	if uploaded.Name != testFile {
		t.Errorf("Uploaded file name mismatch. Expected %s, got %s", testFile, uploaded.Name)
	}

	if uploaded.Sha != testFileSha {
		t.Errorf("Uploaded file sha mismatch. Expected %s, got %s", testFileSha, uploaded.Sha)
	}

	if uploaded.Len != testFileLen {
		t.Errorf("Uploaded file sha mismatch. Expected %d, got %d", testFileLen, uploaded.Len)
	}
}

func TestCreateDeployment(t *testing.T) {
	testFile := "test.txt"

	if err := godotenv.Load(".env"); err != nil {
		t.Error("Failed to load dotenv file.")
	}

	client := Client{
		Token:          os.Getenv("NOW_TOKEN"),
		DeploymentName: os.Getenv("NOW_DEPLOYMENT"),
		Team:           os.Getenv("NOW_TEAMID"),
		Retries:        2,
	}

	f, err := os.Open(testFile)
	if err != nil {
		t.Error("Failed to open test file to upload.", err)
	}
	defer f.Close()
	ogBody, err := ioutil.ReadAll(f)

	uploaded, err := client.UploadFile(bytes.NewReader(ogBody), testFile)
	if err != nil {
		t.Error("Failed to upload file.", err)
	}

	deployment, err := client.CreateDeployment([]*UploadedFile{uploaded})

	if err != nil {
		t.Errorf("Failed to create deployment, %v", err)
	}

	if deployment == nil {
		t.Error("Deployment is nil without an error.")
	}

	if deployment.ID == "" {
		t.Error("Deployment does not contain an ID.")
	}

	if deployment.Url == "" {
		t.Error("Deployment does not contain a URL.")
	}

	if len(deployment.Files) != 1 {
		t.Error("Files length mismatch.")
	}

	client.WaitForReady(DeployResponse{
		Url: deployment.Url,
	})

	res, err := http.Get(deployment.Files[0].Url)
	if err != nil {
		t.Error("Could not get uploaded file.", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error("Could not read uploaded file response.", err)
	}

	if err != nil {
		t.Error("Could not read file to compare to response.", err)
	}

	if string(body) != string(ogBody) {
		t.Error("Uploaded file response and origional file are not equal.")
	}
}
