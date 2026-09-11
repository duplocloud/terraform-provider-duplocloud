package duplocloud

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
)

// ecacheSettleServer serves a canned ECache instance, advancing through the supplied
// readings one poll at a time and repeating the last one forever.
func ecacheSettleServer(t *testing.T, readings []string) (*httptest.Server, *duplosdk.Client) {
	t.Helper()
	var n int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := readings[n]
		if n < len(readings)-1 {
			n++
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	c, err := duplosdk.NewClient(srv.URL, "fake-token")
	if err != nil {
		t.Fatalf("NewClient: %s", err)
	}
	return srv, c
}

func ecacheReading(status string, replicas int, failover, multiAz bool) string {
	return fmt.Sprintf(`{"Identifier":"duplo-qagrpg","Name":"qagrpg","InstanceStatus":%q,`+
		`"Replicas":%d,"AutomaticFailoverEnabled":%t,"MultiAZEnabled":%t}`,
		status, replicas, failover, multiAz)
}

// TestEcacheInstanceWaitUntilSettled covers the DUPLO-44491 regression: the replica
// endpoint answers "accepted", not "applied", so an update that never reached the cache
// used to be reported as a successful apply.
func TestEcacheInstanceWaitUntilSettled(t *testing.T) {
	prevTimeout, prevPoll := ecacheSettleTimeout, ecacheSettlePollInterval
	ecacheSettleTimeout, ecacheSettlePollInterval = time.Second, 50*time.Millisecond
	defer func() { ecacheSettleTimeout, ecacheSettlePollInterval = prevTimeout, prevPoll }()

	wantReplicas := func(n int) func(*duplosdk.DuploEcacheInstance) bool {
		return func(i *duplosdk.DuploEcacheInstance) bool { return i.Replicas == n }
	}

	t.Run("settles once the count lands", func(t *testing.T) {
		srv, c := ecacheSettleServer(t, []string{
			ecacheReading("modifying", 4, false, false),
			ecacheReading("available", 1, false, false),
		})
		defer srv.Close()

		last, err := ecacheInstanceWaitUntilSettled(context.Background(), c, "t1", "qagrpg", "replicas", wantReplicas(1))
		if err != nil {
			t.Fatalf("expected the wait to settle, got: %s", err)
		}
		if last == nil || last.Replicas != 1 {
			t.Fatalf("expected the last reading to report 1 replica, got %+v", last)
		}
	})

	t.Run("fails when the count never changes", func(t *testing.T) {
		srv, c := ecacheSettleServer(t, []string{ecacheReading("available", 4, false, false)})
		defer srv.Close()

		last, err := ecacheInstanceWaitUntilSettled(context.Background(), c, "t1", "qagrpg", "replicas", wantReplicas(1))
		if err == nil {
			t.Fatal("expected an error when the requested replica count never lands")
		}
		// The caller reports this value back to the user, so it has to survive the timeout.
		if last == nil || last.Replicas != 4 {
			t.Fatalf("expected the last reading to report 4 replicas, got %+v", last)
		}
	})

	t.Run("does not trust a reading taken mid-modification", func(t *testing.T) {
		srv, c := ecacheSettleServer(t, []string{ecacheReading("modifying", 1, false, false)})
		defer srv.Close()

		if _, err := ecacheInstanceWaitUntilSettled(context.Background(), c, "t1", "qagrpg", "replicas", wantReplicas(1)); err == nil {
			t.Fatal("expected the wait to keep polling while the instance is still modifying")
		}
	})

	t.Run("keeps polling when the instance cannot be read", func(t *testing.T) {
		// EcacheInstanceGet reports a missing instance as (nil, nil); the wait must not panic.
		srv, c := ecacheSettleServer(t, []string{`{}`})
		defer srv.Close()

		if _, err := ecacheInstanceWaitUntilSettled(context.Background(), c, "t1", "qagrpg", "replicas", wantReplicas(1)); err == nil {
			t.Fatal("expected the wait to time out rather than settle on an unreadable instance")
		}
	})
}
