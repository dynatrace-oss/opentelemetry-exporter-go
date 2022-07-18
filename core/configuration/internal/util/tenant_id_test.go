package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateTenantId(t *testing.T) {
	testCases := []struct {
		tenant     string
		expectedId int32
	}{
		{"tenant", 1238414539},
		{"jmw13303", 1292943070},
		{"äpfel", 1997722269},
		{"elo29571", -136051656},
		{"200082", 160529318},
		{"200083", 1760453844},
		{"200084", 618677635},
		{"200085", -253261892},
		{"200086", 1020956405},
		{"200087", 1709497622},
		{"200088", 1552768655},
		{"200089", -1763495057},
		{"200090", 1366648315},
	}

	for _, tc := range testCases {
		t.Run(tc.tenant, func(t *testing.T) {
			tenantId := CalculateTenantId(tc.tenant)
			assert.Equal(t, tenantId, tc.expectedId)
		})
	}
}
