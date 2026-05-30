package postgres

import (
	"strings"
	"testing"
)

func TestSQLGetVKinoCoinsBalance_IncludesCoinsPurchaseIncome(t *testing.T) {
	t.Parallel()

	if !strings.Contains(sqlGetVKinoCoinsBalance, "'coins_purchase'") {
		t.Fatalf("sqlGetVKinoCoinsBalance must count coins_purchase operations as income")
	}
}
