package storage

import (
	"bytes"
	"encoding/hex"
	"io"
	"io/ioutil"

	"crypto/sha1"
)

const DeploymentName = "now-storage"
const DeployUrl = "https://api.zeit.co/v10/now/deployments"
const DeploymentUrl = "https://api.zeit.co/v10/now/deployments/get?url="
const FileUrl = "https://api.zeit.co/v2/now/files"

func Len(reader io.Reader) (int64, io.Reader, error) {
	buf := &bytes.Buffer{}
	tee := io.TeeReader(reader, buf)

	tempBuff := &bytes.Buffer{}
	nRead, err := io.Copy(tempBuff, tee)
	if err != nil {
		return 0, reader, err
	}

	return nRead, tempBuff, nil
}

func GetSha1(reader io.Reader) (string, io.Reader, error) {
	bts, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", reader, err
	}

	h := sha1.New()
	h.Write(bts)

	return hex.EncodeToString(h.Sum(nil)), bytes.NewReader(bts), nil
}

func (client *Client) GetDeploymentName() string {
	deploymentName := client.DeploymentName

	if deploymentName == "" {
		return DeploymentName
	}

	return deploymentName
}

func (client *Client) GetDeployUrl() string {
	res := DeployUrl

	if client.Team != "" {
		res += "?teamId=" + client.Team
	}

	return res
}

func (client *Client) GetDeploymentUrl(deployment DeployResponse) string {
	return DeploymentName + deployment.Url
}

func (client *Client) GetFileUrl() string {
	res := FileUrl

	if client.Team != "" {
		res += "?teamId=" + client.Team
	}

	return res
}

func (client *Client) GetBearerToken() string {
	return "Bearer " + client.Token
}
