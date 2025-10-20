package account

type AccountType string

const (
	AccountTypeUser   AccountType = "user"
	AccountTypeClient AccountType = "client"
	AccountTypeSystem AccountType = "system"
)

type Account struct {
	ID   string      `json:"id"`
	Type AccountType `json:"type"`
}
