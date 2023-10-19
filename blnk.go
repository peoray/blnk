package blnk

import (
	"encoding/json"
	"errors"
	"time"
)

type Filter struct {
	Status string
}

type Transaction struct {
	ID                  int64                  `json:"-"`
	TransactionID       string                 `json:"id"`
	Tag                 string                 `json:"tag"`
	Reference           string                 `json:"reference" binding:"required" error:"Please provide a reference."`
	Amount              int64                  `json:"amount" binding:"required" error:"Please provide an amount."`
	Currency            string                 `json:"currency" binding:"required" error:"Please provide a valid currency."`
	DRCR                string                 `json:"drcr" binding:"required" error:"Please provide a 'drcr' value. It must be either 'Credit' or 'Debit'."`
	Status              string                 `json:"status"`
	LedgerID            string                 `json:"ledger_id"`
	BalanceID           string                 `json:"balance_id" binding:"required" error:"amount is required: please provide an amount."`
	CreditBalanceBefore int64                  `json:"credit_balance_before"`
	DebitBalanceBefore  int64                  `json:"debit_balance_before"`
	CreditBalanceAfter  int64                  `json:"credit_balance_after"`
	DebitBalanceAfter   int64                  `json:"debit_balance_after"`
	BalanceBefore       int64                  `json:"balance_before"`
	BalanceAfter        int64                  `json:"balance_after"`
	CreatedAt           time.Time              `json:"created_at"`
	ScheduledFor        time.Time              `json:"scheduled_for,omitempty"`
	SkipBalanceUpdate   bool                   `json:"-"`
	MetaData            map[string]interface{} `json:"meta_data,omitempty"`
}

type TransactionFilter struct {
	ID                       int64     `json:"id"`
	Tag                      string    `json:"tag"`
	DRCR                     string    `json:"drcr"`
	AmountRange              int64     `json:"amount_range"`
	CreditBalanceBeforeRange int64     `json:"credit_balance_before_range"`
	DebitBalanceBeforeRange  int64     `json:"debit_balance_before_range"`
	CreditBalanceAfterRange  int64     `json:"credit_balance_after_range"`
	DebitBalanceAfterRange   int64     `json:"debit_balance_after_range"`
	BalanceBeforeRange       int64     `json:"balance_before"`
	BalanceAfterRange        int64     `json:"balance_after"`
	From                     time.Time `json:"from"`
	To                       time.Time `json:"to"`
}

type Balance struct {
	ID                 int64                  `json:"-"`
	BalanceID          string                 `json:"id"`
	Balance            int64                  `json:"balance"`
	CreditBalance      int64                  `json:"credit_balance"`
	DebitBalance       int64                  `json:"debit_balance"`
	Currency           string                 `json:"currency"`
	CurrencyMultiplier int64                  `json:"currency_multiplier"`
	LedgerID           string                 `json:"ledger_id" binding:"required" error:"Please provide a ledger ID."`
	IdentityID         string                 `json:"identity_id"`
	Identity           *Identity              `json:"identity,omitempty"`
	Ledger             *Ledger                `json:"ledger,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	MetaData           map[string]interface{} `json:"meta_data"`
}

type BalanceFilter struct {
	ID                 int64     `json:"id"`
	BalanceRange       string    `json:"balance_range"`
	CreditBalanceRange string    `json:"credit_balance_range"`
	DebitBalanceRange  string    `json:"debit_balance_range"`
	Currency           string    `json:"currency"`
	LedgerID           string    `json:"ledger_id"`
	From               time.Time `json:"from"`
	To                 time.Time `json:"to"`
}

type Ledger struct {
	ID        int64                  `json:"-"`
	LedgerID  string                 `json:"id"`
	Name      string                 `json:"name" binding:"required" error:"Please provide a name for your ledger."`
	CreatedAt time.Time              `json:"created_at"`
	MetaData  map[string]interface{} `json:"meta_data,omitempty"`
}

type LedgerFilter struct {
	ID   int64     `json:"id"`
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type Policy struct {
	ID        int64     `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	Operator  string    `json:"operator,omitempty"`
	Field     string    `json:"field,omitempty"`
	Value     string    `json:"value"`
	Action    string    `json:"action,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Identity struct {
	IdentityID   string                 `json:"identity_id"`
	IdentityType string                 `json:"identity_type"` // "individual" or "organization"
	Individual   Individual             `json:"individual"`
	Organization Organization           `json:"organization"`
	Street       string                 `json:"street"`
	Country      string                 `json:"country"`
	State        string                 `json:"state"`
	PostCode     string                 `json:"post_code"`
	City         string                 `json:"city"`
	CreatedAt    time.Time              `json:"created_at"`
	MetaData     map[string]interface{} `json:"meta_data"`
}

type Account struct {
	AccountID string                 `json:"account_id"`
	Name      string                 `json:"name"`
	Number    string                 `json:"number"`
	BankName  string                 `json:"bank_name"`
	MetaData  map[string]interface{} `json:"meta_data"`
}

type Individual struct {
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	OtherNames   string    `json:"other_names"`
	Gender       string    `json:"gender"`
	DOB          time.Time `json:"dob"`
	EmailAddress string    `json:"email_address"`
	PhoneNumber  string    `json:"phone_number"`
	Nationality  string    `json:"nationality"`
}

type Organization struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

func (balance *Balance) AddCredit(amount int64) {
	balance.CreditBalance += amount
}

func (balance *Balance) AddDebit(amount int64) {
	balance.DebitBalance += amount
}

func (balance *Balance) ComputeBalance() {
	balance.Balance = balance.CreditBalance - balance.DebitBalance
}

func (balance *Balance) AttachBalanceBefore(transaction *Transaction) {
	transaction.DebitBalanceBefore = balance.DebitBalance
	transaction.CreditBalanceBefore = balance.CreditBalance
	transaction.BalanceBefore = balance.Balance
}

func (balance *Balance) AttachBalanceAfter(transaction *Transaction) {
	transaction.DebitBalanceAfter = balance.DebitBalance
	transaction.CreditBalanceAfter = balance.CreditBalance
	transaction.BalanceAfter = balance.Balance
}

func (balance *Balance) applyMultiplier(transaction *Transaction) {
	if balance.CurrencyMultiplier == 0 {
		balance.CurrencyMultiplier = 1
	}
	transaction.Amount = transaction.Amount * balance.CurrencyMultiplier
}

func (balance *Balance) UpdateBalances(transaction *Transaction) error {
	// Validate transaction
	err := transaction.validate()
	if err != nil {
		return err
	}

	balance.applyMultiplier(transaction)
	balance.AttachBalanceBefore(transaction)
	if transaction.DRCR == "Credit" {
		balance.AddCredit(transaction.Amount)
	} else {
		balance.AddDebit(transaction.Amount)
	}

	balance.ComputeBalance()
	balance.AttachBalanceAfter(transaction)
	return nil
}

func (transaction *Transaction) validate() error {
	if transaction.Amount <= 0 {
		return errors.New("transaction amount must be positive")
	}
	if transaction.DRCR != "Credit" && transaction.DRCR != "Debit" {
		return errors.New("transaction DRCR must be 'Credit' or 'Debit'")
	}
	return nil
}

func (transaction *Transaction) ToJSON() ([]byte, error) {
	return json.Marshal(transaction)
}

//balance, balance before, balance after, credit balance, debit balance, credit balance before, credit balance after, debit balance before, debit balance after
//amount
//apply_on: credit, debit
//action: allow, deny
//if amount > 1000 && credit balance < 500000 allow
//if credit_balance <= 4000 deny
//