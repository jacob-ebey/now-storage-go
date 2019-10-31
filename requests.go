package storage

type DeploymentRequest struct {
	Public  bool             `json:"public"`
	Version int              `json:"version"`
	Name    string           `json:"name"`
	Files   []DeploymentFile `json:"files"`
}

type DeploymentFile struct {
	File string `json:"file"`
	Sha  string `json:"sha"`
	Size int64  `json:"size"`
}
