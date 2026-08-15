package gobrightbox

import (
	"time"
)

// Account represents a Brightbox Cloud Account
// https://api.gb1.brightbox.com/1.0/#account
type Account struct {
	Id                    string
	Name                  string
	Status                string
	Address1              string `json:"address_1"`
	Address2              string `json:"address_2"`
	City                  string
	County                string
	Postcode              string
	CountryCode           string     `json:"country_code"`
	CountryName           string     `json:"country_name"`
	VatRegistrationNumber string     `json:"vat_registration_number"`
	TelephoneNumber       string     `json:"telephone_number"`
	TelephoneVerified     bool       `json:"telephone_verified"`
	VerifiedTelephone     string     `json:"verified_telephone"`
	VerifiedAt            *time.Time `json:"verified_at"`
	VerifiedIp            string     `json:"verified_ip"`
	ValidCreditCard       bool       `json:"valid_credit_card"`
	CreatedAt             *time.Time `json:"created_at"`
	RamLimit              int        `json:"ram_limit"`
	RamUsed               int        `json:"ram_used"`
	DbsRamLimit           int        `json:"dbs_ram_limit"`
	DbsRamUsed            int        `json:"dbs_ram_used"`
	CloudIpsLimit         int        `json:"cloud_ips_limit"`
	CloudIpsUsed          int        `json:"cloud_ips_used"`
	LoadBalancersLimit    int        `json:"load_balancers_limit"`
	LoadBalancersUsed     int        `json:"load_balancers_used"`
	LibraryFtpHost        string     `json:"library_ftp_host"`
	LibraryFtpUser        string     `json:"library_ftp_user"`
	LibraryFtpPassword    string     `json:"library_ftp_password"`
	Owner                 User
	Users                 []User
}

// Accounts retrieves a list of all accounts associated with the client.
//
// API Clients are only ever associated with one single account. User clients
// can have multiple accounts, through collaborations.
func (c *Client) Accounts() ([]Account, error) {
	var accounts []Account
	_, err := c.MakeApiRequest("GET", "/1.0/accounts?nested=false", nil, &accounts)
	if err != nil {
		return nil, err
	}
	return accounts, err
}

// Account retrieves a detailed view of one account
func (c *Client) Account(identifier string) (*Account, error) {
	account := new(Account)
	_, err := c.MakeApiRequest("GET", "/1.0/accounts/"+identifier, nil, account)
	if err != nil {
		return nil, err
	}
	return account, err
}
