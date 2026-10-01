package duplocloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// DUPLO-44791.  The backend writes duplocloud.net/skip-encoding onto a secret the first
// time it is updated, so a configuration that sets secret_annotations explicitly would
// plan the key's removal on every run.

func TestK8sSecretAnnotationsForState_DropsSkipEncodingTheConfigDoesNotMention(t *testing.T) {
	backend := map[string]string{"duplocloud.net/skip-encoding": "False"}

	assert.Equal(t, map[string]string{}, k8sSecretAnnotationsForState(map[string]interface{}{}, backend), "empty map")
	assert.Equal(t, map[string]string{}, k8sSecretAnnotationsForState(nil, backend), "omitted")

	backend = map[string]string{"n": "1", "duplocloud.net/skip-encoding": "False"}
	assert.Equal(t, map[string]string{"n": "1"}, k8sSecretAnnotationsForState(map[string]interface{}{"n": "1"}, backend), "other keys kept")
}

func TestK8sSecretAnnotationsForState_KeepsSkipEncodingTheConfigSets(t *testing.T) {
	backend := map[string]string{"duplocloud.net/skip-encoding": "true"}
	configured := map[string]interface{}{"duplocloud.net/skip-encoding": "true"}

	assert.Equal(t, backend, k8sSecretAnnotationsForState(configured, backend))
}

func TestK8sSecretAnnotationsForState_LeavesTheBackendMapAlone(t *testing.T) {
	backend := map[string]string{"duplocloud.net/skip-encoding": "False"}

	k8sSecretAnnotationsForState(nil, backend)

	assert.Contains(t, backend, "duplocloud.net/skip-encoding")
}

// The backend rewrites the annotation as bool.ToString() on every update, so a configured
// "true" comes back as "True".  It parses the value case-insensitively, so the two agree.
func TestK8sSecretAnnotationsForState_KeepsTheConfiguredSpellingOfSkipEncoding(t *testing.T) {
	backend := map[string]string{"duplocloud.net/skip-encoding": "True"}
	configured := map[string]interface{}{"duplocloud.net/skip-encoding": "true"}

	assert.Equal(t, map[string]string{"duplocloud.net/skip-encoding": "true"}, k8sSecretAnnotationsForState(configured, backend))
}

func TestK8sSecretAnnotationsForState_ReportsADifferentSkipEncodingValue(t *testing.T) {
	backend := map[string]string{"duplocloud.net/skip-encoding": "False"}
	configured := map[string]interface{}{"duplocloud.net/skip-encoding": "true"}

	assert.Equal(t, backend, k8sSecretAnnotationsForState(configured, backend))
}
