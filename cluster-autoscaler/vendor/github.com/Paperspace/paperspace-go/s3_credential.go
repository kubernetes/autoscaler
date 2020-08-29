package paperspace

type S3Credential struct {
	AccessKey string `json:"accessKey"`
	Bucket    string `json:"bucket"`
	SecretKey string `json:"secretKey,omitempty"`
}
