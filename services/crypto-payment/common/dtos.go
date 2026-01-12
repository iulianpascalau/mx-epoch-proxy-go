package common

// BalanceEntry represents a record for a certain ID with an addres
type BalanceEntry struct {
	ID             int
	Address        string
	CurrentBalance float64
	TotalRequests  int
}
