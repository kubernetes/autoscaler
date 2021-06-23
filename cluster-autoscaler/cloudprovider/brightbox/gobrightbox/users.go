package gobrightbox

// User represents a Brightbox User
// https://api.gb1.brightbox.com/1.0/#user
type User struct {
	Id             string
	Name           string
	EmailAddress   string `json:"email_address"`
	EmailVerified  bool   `json:"email_verified"`
	SshKey         string `json:"ssh_key"`
	MessagingPref  bool   `json:"messaging_pref"`
	Accounts       []*Account
	DefaultAccount *Account `json:"default_account"`
}
