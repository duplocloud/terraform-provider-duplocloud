package duplocloud

import (
	"reflect"
	"testing"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Test_expandTrustedKeyGroups(t *testing.T) {
	cases := []struct {
		name     string
		given    []interface{}
		expected *duplosdk.DuploCFDTrustedKeyGroups
	}{
		{
			name:  "empty list disables trusted key groups",
			given: []interface{}{},
			expected: &duplosdk.DuploCFDTrustedKeyGroups{
				Enabled:  false,
				Quantity: 0,
			},
		},
		{
			name:  "non-empty list enables trusted key groups",
			given: []interface{}{"key-group-1", "key-group-2"},
			expected: &duplosdk.DuploCFDTrustedKeyGroups{
				Enabled:  true,
				Quantity: 2,
				Items:    []string{"key-group-1", "key-group-2"},
			},
		},
		{
			name:  "quantity matches filtered items when list contains empty strings",
			given: []interface{}{"key-group-1", "", "key-group-2"},
			expected: &duplosdk.DuploCFDTrustedKeyGroups{
				Enabled:  true,
				Quantity: 2,
				Items:    []string{"key-group-1", "key-group-2"},
			},
		},
		{
			name:  "list of only empty strings disables trusted key groups",
			given: []interface{}{""},
			expected: &duplosdk.DuploCFDTrustedKeyGroups{
				Enabled:  false,
				Quantity: 0,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := expandTrustedKeyGroups(c.given)
			if !reflect.DeepEqual(actual, c.expected) {
				t.Errorf("expected %+v, got %+v", c.expected, actual)
			}
		})
	}
}

func Test_flattenTrustedKeyGroups(t *testing.T) {
	cases := []struct {
		name     string
		given    *duplosdk.DuploCFDTrustedKeyGroups
		expected []interface{}
	}{
		{
			name:     "nil items flattens to empty list",
			given:    &duplosdk.DuploCFDTrustedKeyGroups{Enabled: false, Quantity: 0},
			expected: []interface{}{},
		},
		{
			name: "items flatten to string list",
			given: &duplosdk.DuploCFDTrustedKeyGroups{
				Enabled:  true,
				Quantity: 2,
				Items:    []string{"key-group-1", "key-group-2"},
			},
			expected: []interface{}{"key-group-1", "key-group-2"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := flattenTrustedKeyGroups(c.given)
			if !reflect.DeepEqual(actual, c.expected) {
				t.Errorf("expected %+v, got %+v", c.expected, actual)
			}
		})
	}
}

func Test_cacheBehaviorConfiguresAttr(t *testing.T) {
	behaviorWithValues := cty.ObjectVal(map[string]cty.Value{
		"trusted_key_groups": cty.ListVal([]cty.Value{cty.StringVal("kg-1")}),
	})
	behaviorWithEmptyList := cty.ObjectVal(map[string]cty.Value{
		"trusted_key_groups": cty.ListValEmpty(cty.String),
	})
	behaviorWithoutAttr := cty.ObjectVal(map[string]cty.Value{
		"trusted_key_groups": cty.NullVal(cty.List(cty.String)),
	})
	behaviorObjType := cty.Object(map[string]cty.Type{"trusted_key_groups": cty.List(cty.String)})
	behaviorWithUnknownAttr := cty.ObjectVal(map[string]cty.Value{
		"trusted_key_groups": cty.UnknownVal(cty.List(cty.String)),
	})

	behaviorWithSigners := cty.ObjectVal(map[string]cty.Value{
		"trusted_key_groups": cty.NullVal(cty.List(cty.String)),
		"trusted_signers":    cty.ListVal([]cty.Value{cty.StringVal("111122223333")}),
	})

	cases := []struct {
		name     string
		block    cty.Value
		index    int
		attr     string
		expected bool
	}{
		{
			name:     "trusted_signers checked independently of trusted_key_groups",
			block:    cty.ListVal([]cty.Value{behaviorWithSigners}),
			index:    0,
			attr:     "trusted_signers",
			expected: true,
		},
		{
			name:     "trusted_key_groups null while trusted_signers set",
			block:    cty.ListVal([]cty.Value{behaviorWithSigners}),
			index:    0,
			attr:     "trusted_key_groups",
			expected: false,
		},
		{
			name:     "explicit values present",
			block:    cty.ListVal([]cty.Value{behaviorWithValues}),
			index:    0,
			expected: true,
		},
		{
			name:     "explicit empty list is still configured",
			block:    cty.ListVal([]cty.Value{behaviorWithEmptyList}),
			index:    0,
			expected: true,
		},
		{
			name:     "attribute omitted from config",
			block:    cty.ListVal([]cty.Value{behaviorWithoutAttr}),
			index:    0,
			expected: false,
		},
		{
			name:     "block itself absent",
			block:    cty.NullVal(cty.List(behaviorObjType)),
			index:    0,
			expected: false,
		},
		{
			name:     "index out of range",
			block:    cty.ListVal([]cty.Value{behaviorWithValues}),
			index:    1,
			expected: false,
		},
		{
			name:     "empty block list",
			block:    cty.ListValEmpty(behaviorObjType),
			index:    0,
			expected: false,
		},
		{
			// e.g. trusted_key_groups = [some_resource.x.id] where some_resource.x
			// hasn't resolved yet - must be treated as managed, not absent, or the
			// caller would clobber the pending value with stale existing data.
			name:     "attribute value itself is unknown",
			block:    cty.ListVal([]cty.Value{behaviorWithUnknownAttr}),
			index:    0,
			expected: true,
		},
		{
			name:     "behavior object itself is unknown",
			block:    cty.ListVal([]cty.Value{cty.UnknownVal(behaviorObjType)}),
			index:    0,
			expected: true,
		},
		{
			name:     "block itself is unknown",
			block:    cty.UnknownVal(cty.List(behaviorObjType)),
			index:    0,
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.attr == "" {
				c.attr = "trusted_key_groups"
			}
			raw := cty.ObjectVal(map[string]cty.Value{"default_cache_behavior": c.block})
			actual := cacheBehaviorConfiguresAttr(raw, "default_cache_behavior", c.index, c.attr)
			if actual != c.expected {
				t.Errorf("expected %v, got %v", c.expected, actual)
			}
		})
	}
}

func Test_preserveOrderedBehaviorsTrust(t *testing.T) {
	protectedTKG := &duplosdk.DuploCFDTrustedKeyGroups{Enabled: true, Quantity: 1, Items: []string{"kg-protected"}}
	protectedTS := &duplosdk.DuploCFDTrustedSigners{Enabled: true, Quantity: 1, Items: []string{"111122223333"}}
	disabledTKG := func() *duplosdk.DuploCFDTrustedKeyGroups {
		return &duplosdk.DuploCFDTrustedKeyGroups{Enabled: false}
	}
	disabledTS := func() *duplosdk.DuploCFDTrustedSigners {
		return &duplosdk.DuploCFDTrustedSigners{Enabled: false}
	}

	t.Run("a new behavior inserted at index 0 does not inherit the old index-0 behavior's key groups", func(t *testing.T) {
		// Before: only "/protected/*" existed, with trusted key groups.
		existing := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/protected/*", TrustedKeyGroups: protectedTKG},
		}
		// After: user inserts "/public/*" ahead of it. Neither behavior manages
		// either trust attribute in config.
		updated := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/public/*", TrustedKeyGroups: disabledTKG(), TrustedSigners: disabledTS()},
			{PathPattern: "/protected/*", TrustedKeyGroups: disabledTKG(), TrustedSigners: disabledTS()},
		}

		preserveOrderedBehaviorsTrust(updated, existing, []bool{false, false}, []bool{false, false})

		if trustedKeyGroupsEnabled(updated[0].TrustedKeyGroups) {
			t.Errorf("expected /public/* to remain disabled, got %+v", updated[0].TrustedKeyGroups)
		}
		if !reflect.DeepEqual(updated[1].TrustedKeyGroups, protectedTKG) {
			t.Errorf("expected /protected/* to keep its existing key groups, got %+v", updated[1].TrustedKeyGroups)
		}
	})

	t.Run("reordering behaviors still preserves by path, not position", func(t *testing.T) {
		existing := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/a/*", TrustedKeyGroups: disabledTKG()},
			{PathPattern: "/b/*", TrustedKeyGroups: protectedTKG},
		}
		// Reordered: /b/* is now first.
		updated := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/b/*", TrustedKeyGroups: disabledTKG(), TrustedSigners: disabledTS()},
			{PathPattern: "/a/*", TrustedKeyGroups: disabledTKG(), TrustedSigners: disabledTS()},
		}

		preserveOrderedBehaviorsTrust(updated, existing, []bool{false, false}, []bool{false, false})

		if !reflect.DeepEqual(updated[0].TrustedKeyGroups, protectedTKG) {
			t.Errorf("expected /b/* to keep its key groups after reorder, got %+v", updated[0].TrustedKeyGroups)
		}
		if trustedKeyGroupsEnabled(updated[1].TrustedKeyGroups) {
			t.Errorf("expected /a/* to remain disabled, got %+v", updated[1].TrustedKeyGroups)
		}
	})

	t.Run("an explicitly configured attribute is never overwritten", func(t *testing.T) {
		existing := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/a/*", TrustedKeyGroups: protectedTKG},
		}
		explicit := &duplosdk.DuploCFDTrustedKeyGroups{Enabled: true, Quantity: 1, Items: []string{"kg-explicit"}}
		updated := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/a/*", TrustedKeyGroups: explicit, TrustedSigners: disabledTS()},
		}

		preserveOrderedBehaviorsTrust(updated, existing, []bool{true}, []bool{false})

		if !reflect.DeepEqual(updated[0].TrustedKeyGroups, explicit) {
			t.Errorf("expected explicitly configured value to be kept, got %+v", updated[0].TrustedKeyGroups)
		}
	})

	t.Run("trusted_signers is preserved on omission just like key groups", func(t *testing.T) {
		existing := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/a/*", TrustedSigners: protectedTS},
		}
		updated := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/a/*", TrustedKeyGroups: disabledTKG(), TrustedSigners: disabledTS()},
		}

		preserveOrderedBehaviorsTrust(updated, existing, []bool{false}, []bool{false})

		if !reflect.DeepEqual(updated[0].TrustedSigners, protectedTS) {
			t.Errorf("expected /a/* to keep its existing signers, got %+v", updated[0].TrustedSigners)
		}
	})

	t.Run("switching a behavior from key groups to signers disables the stale key groups", func(t *testing.T) {
		// The DUPLO-44632 headline case: key groups exist on the backend, the user's
		// config now sets trusted_signers and omits trusted_key_groups. The stale
		// key groups (whether carried via computed state into expand or preserved
		// from existing) must lose to the explicitly configured signers.
		existing := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/a/*", TrustedKeyGroups: protectedTKG},
		}
		updated := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/a/*", TrustedKeyGroups: protectedTKG, TrustedSigners: protectedTS},
		}

		preserveOrderedBehaviorsTrust(updated, existing, []bool{false}, []bool{true})

		if trustedKeyGroupsEnabled(updated[0].TrustedKeyGroups) {
			t.Errorf("expected stale key groups to be disabled, got %+v", updated[0].TrustedKeyGroups)
		}
		if !reflect.DeepEqual(updated[0].TrustedSigners, protectedTS) {
			t.Errorf("expected configured signers to be kept, got %+v", updated[0].TrustedSigners)
		}
	})

	t.Run("switching a behavior from signers to key groups disables the stale signers", func(t *testing.T) {
		existing := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/a/*", TrustedSigners: protectedTS},
		}
		updated := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/a/*", TrustedKeyGroups: protectedTKG, TrustedSigners: protectedTS},
		}

		preserveOrderedBehaviorsTrust(updated, existing, []bool{true}, []bool{false})

		if trustedSignersEnabled(updated[0].TrustedSigners) {
			t.Errorf("expected stale signers to be disabled, got %+v", updated[0].TrustedSigners)
		}
		if !reflect.DeepEqual(updated[0].TrustedKeyGroups, protectedTKG) {
			t.Errorf("expected configured key groups to be kept, got %+v", updated[0].TrustedKeyGroups)
		}
	})

	t.Run("renaming a behavior's path while swapping to signers still disables stale key groups", func(t *testing.T) {
		// The renamed PathPattern has no existing match, so nothing is preserved -
		// but the expanded behavior still carries stale key groups from computed
		// state alongside the newly configured signers. Conflict resolution must
		// run for unmatched behaviors too.
		existing := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/old/*", TrustedKeyGroups: protectedTKG},
		}
		updated := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/new/*", TrustedKeyGroups: protectedTKG, TrustedSigners: protectedTS},
		}

		preserveOrderedBehaviorsTrust(updated, existing, []bool{false}, []bool{true})

		if trustedKeyGroupsEnabled(updated[0].TrustedKeyGroups) {
			t.Errorf("expected stale key groups on renamed behavior to be disabled, got %+v", updated[0].TrustedKeyGroups)
		}
		if !reflect.DeepEqual(updated[0].TrustedSigners, protectedTS) {
			t.Errorf("expected configured signers to be kept, got %+v", updated[0].TrustedSigners)
		}
	})

	t.Run("both explicitly configured and enabled is left for CloudFront to reject", func(t *testing.T) {
		existing := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/a/*"},
		}
		updated := []duplosdk.DuploAwsCloudfrontCacheBehavior{
			{PathPattern: "/a/*", TrustedKeyGroups: protectedTKG, TrustedSigners: protectedTS},
		}

		preserveOrderedBehaviorsTrust(updated, existing, []bool{true}, []bool{true})

		if !reflect.DeepEqual(updated[0].TrustedKeyGroups, protectedTKG) || !reflect.DeepEqual(updated[0].TrustedSigners, protectedTS) {
			t.Errorf("expected both configured values untouched, got kg=%+v ts=%+v", updated[0].TrustedKeyGroups, updated[0].TrustedSigners)
		}
	})
}

func Test_trustedKeyGroupsEnabled(t *testing.T) {
	cases := []struct {
		name     string
		tkg      *duplosdk.DuploCFDTrustedKeyGroups
		expected bool
	}{
		{name: "nil", tkg: nil, expected: false},
		{name: "disabled", tkg: &duplosdk.DuploCFDTrustedKeyGroups{Enabled: false}, expected: false},
		{name: "enabled", tkg: &duplosdk.DuploCFDTrustedKeyGroups{Enabled: true, Quantity: 1, Items: []string{"kg-1"}}, expected: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if actual := trustedKeyGroupsEnabled(c.tkg); actual != c.expected {
				t.Errorf("expected %v, got %v", c.expected, actual)
			}
		})
	}
}

func Test_expandTrustedSigners(t *testing.T) {
	cases := []struct {
		name     string
		given    []interface{}
		expected *duplosdk.DuploCFDTrustedSigners
	}{
		{
			name:     "empty list disables trusted signers",
			given:    []interface{}{},
			expected: &duplosdk.DuploCFDTrustedSigners{Enabled: false, Quantity: 0},
		},
		{
			name:  "quantity matches filtered items when list contains empty strings",
			given: []interface{}{"111122223333", ""},
			expected: &duplosdk.DuploCFDTrustedSigners{
				Enabled:  true,
				Quantity: 1,
				Items:    []string{"111122223333"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := expandTrustedSigners(c.given)
			if !reflect.DeepEqual(actual, c.expected) {
				t.Errorf("expected %+v, got %+v", c.expected, actual)
			}
		})
	}
}

func Test_trustedSignersEnabled(t *testing.T) {
	cases := []struct {
		name     string
		ts       *duplosdk.DuploCFDTrustedSigners
		expected bool
	}{
		{name: "nil", ts: nil, expected: false},
		{name: "disabled", ts: &duplosdk.DuploCFDTrustedSigners{Enabled: false}, expected: false},
		{name: "enabled", ts: &duplosdk.DuploCFDTrustedSigners{Enabled: true, Quantity: 1, Items: []string{"111122223333"}}, expected: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if actual := trustedSignersEnabled(c.ts); actual != c.expected {
				t.Errorf("expected %v, got %v", c.expected, actual)
			}
		})
	}
}

func Test_suppressViewerCertificateManagedByAws(t *testing.T) {
	cases := []struct {
		name     string
		vc       map[string]interface{}
		expected bool
	}{
		{
			name:     "default certificate suppresses",
			vc:       map[string]interface{}{"cloudfront_default_certificate": true},
			expected: true,
		},
		{
			name:     "acm certificate does not suppress",
			vc:       map[string]interface{}{"acm_certificate_arn": "arn:aws:acm:us-east-1:111122223333:certificate/abc"},
			expected: false,
		},
		{
			name:     "iam certificate does not suppress",
			vc:       map[string]interface{}{"iam_certificate_id": "ASCAexample"},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, duploAwsCloudfrontDistributionSchema(), map[string]interface{}{
				"viewer_certificate": []interface{}{c.vc},
			})
			actual := suppressViewerCertificateManagedByAws("viewer_certificate.0.ssl_support_method", "", "sni-only", d)
			if actual != c.expected {
				t.Errorf("expected %v, got %v", c.expected, actual)
			}
		})
	}
}
