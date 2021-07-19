package gobrightbox

import (
	"time"
)

// Image represents a Machine Image
// https://api.gb1.brightbox.com/1.0/#image
type Image struct {
	Id                string
	Name              string
	Username          string
	Status            string
	Locked            bool
	Description       string
	Source            string
	Arch              string
	CreatedAt         time.Time `json:"created_at"`
	Official          bool
	Public            bool
	Owner             string
	SourceType        string `json:"source_type"`
	VirtualSize       int    `json:"virtual_size"`
	DiskSize          int    `json:"disk_size"`
	CompatibilityMode bool   `json:"compatibility_mode"`
	AncestorId        string `json:"ancestor_id"`
	LicenceName       string `json:"licence_name"`
}

// Images retrieves a list of all images
func (c *Client) Images() ([]Image, error) {
	var images []Image
	_, err := c.MakeApiRequest("GET", "/1.0/images", nil, &images)
	if err != nil {
		return nil, err
	}
	return images, err
}

// Image retrieves a detailed view of one image
func (c *Client) Image(identifier string) (*Image, error) {
	image := new(Image)
	_, err := c.MakeApiRequest("GET", "/1.0/images/"+identifier, nil, image)
	if err != nil {
		return nil, err
	}
	return image, err
}

// DestroyImage issues a request to destroy the image
func (c *Client) DestroyImage(identifier string) error {
	_, err := c.MakeApiRequest("DELETE", "/1.0/images/"+identifier, nil, nil)
	if err != nil {
		return err
	}
	return nil
}
