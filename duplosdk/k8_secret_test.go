package duplosdk

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Integers above 2^53 in a decoded secret value come back with their digits intact,
// for a single secret and for each secret in a list.
func TestDuploK8sSecret_UnmarshalKeepsLargeIntegers(t *testing.T) {
	body := `{"SecretName":"s","SecretType":"Opaque","SecretData":{"k":{"n":9007199254740993}}}`

	var one DuploK8sSecret
	assert.NoError(t, json.Unmarshal([]byte(body), &one))
	out, err := json.Marshal(one.SecretData)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"k":{"n":9007199254740993}}`, string(out))
	assert.Contains(t, string(out), "9007199254740993")
	assert.Equal(t, "s", one.SecretName)

	var list []DuploK8sSecret
	assert.NoError(t, json.Unmarshal([]byte("["+body+"]"), &list))
	out, err = json.Marshal(list[0].SecretData)
	assert.NoError(t, err)
	assert.Contains(t, string(out), "9007199254740993")
}
