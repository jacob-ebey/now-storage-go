# Now Storage Go

Use Now static deployments to upload and store files.

## Usage

```bash
> go get github.com/jacob-ebey/now-storage-go
```

### Simple

Upload a file and create a deployment in a single step.

```golang
package main

import (
  "fmt"

  "github.com/jacob-ebey/now-storage-go"
)

func main() {
  // Get a os.Reader for the file we want to upload
  f, err := os.Open(testFile)
	if err != nil {
    fmt.Println("Failed to open test file to upload.", err)
    return
	}
  defer f.Close()
  
  // Create our storage client
  client := Client{
		Token:          os.Getenv("NOW_TOKEN"),
		DeploymentName: os.Getenv("NOW_DEPLOYMENT"),
		Team:           os.Getenv("NOW_TEAMID"),
		Retries:        2,
  }
  
  // Upload the file and create a deployment
  deployment, err := client.Upload(bytes.NewReader(ogBody), testFile)
	if err != nil {
    fmt.Println("Failed to deploy file.", err)
    return
  }

  // Wait for the deployment to be ready
  if err := client.WaitForReady(deployment); err != nil {
    fmt.Println("Failed to wait for deployment.", err)
    return
  }
  
  fmt.Println(deployment.Url)          // Print the base URL of the deployment
  fmt.Println(deployment.Files[0].Url) // Print the URL of the uploaded file
}
```

### Advanced

You can also upload files and manually create deployments if applicable to your application.

```golang
package main

import (
  "fmt"

  "github.com/jacob-ebey/now-storage-go"
)

func main() {
  // Get a os.Reader for the file we want to upload
  f, err := os.Open(testFile)
	if err != nil {
    fmt.Println("Failed to open test file to upload.", err)
    return
	}
  defer f.Close()
  
  // Create our storage client
  client := Client{
		Token:          os.Getenv("NOW_TOKEN"),
		DeploymentName: os.Getenv("NOW_DEPLOYMENT"),
		Team:           os.Getenv("NOW_TEAMID"),
		Retries:        2,
  }
  
  // Upload the file without creating a deployment
  uploadedFile, err := client.UploadFile(bytes.NewReader(ogBody), testFile)
	if err != nil {
    fmt.Println("Failed to upload file.", err)
    return
  }
  
  // Create a deployment consisting of the uploaded file
  deployment, err := client.CreateDeployment([]*UploadedFile{uploadedFile})
	if err != nil {
    fmt.Println("Failed to create deployment.", err)
    return
  }
  
  // Wait for the deployment to be ready
  if err := client.WaitForReady(deployment); err != nil {
    fmt.Println("Failed to wait for deployment.", err)
    return
  }
  
  fmt.Println(deployment.Url)          // Print the base URL of the deployment
  fmt.Println(deployment.Files[0].Url) // Print the URL of the uploaded file
}
```
