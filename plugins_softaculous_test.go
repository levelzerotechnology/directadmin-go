package directadmin

import (
	"encoding/json"
	"testing"
)

// Softaculous answers with disable_wp_cron as an empty string, the number 1, or the
// boolean true depending on the call, so a plain bool field would fail to decode the
// response at all.
func TestSoftaculousSettingsDecodeMixedBooleans(t *testing.T) {
	tests := map[string]struct {
		payload string
		want    bool
	}{
		"empty string": {payload: `{"disable_wp_cron":""}`},
		"json true":    {payload: `{"disable_wp_cron":true}`, want: true},
		"number one":   {payload: `{"disable_wp_cron":1}`, want: true},
		"number zero":  {payload: `{"disable_wp_cron":0}`},
		"string one":   {payload: `{"disable_wp_cron":"1"}`, want: true},
		"absent":       {payload: `{}`},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var settings SoftaculousInstallationSettings
			if err := json.Unmarshal([]byte(tt.payload), &settings); err != nil {
				t.Fatalf("decode: %v", err)
			}

			if got := settings.WPCronDisabled(); got != tt.want {
				t.Errorf("WPCronDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

// The core level is written as auto_upgrade_core and read back as eu_auto_upgrade. Reading
// the wrong key makes every comparison fail and every site look unconfigured.
func TestSoftaculousAutoUpgradeLevelReadsTheReturnedKey(t *testing.T) {
	var settings SoftaculousInstallationSettings
	if err := json.Unmarshal([]byte(`{"eu_auto_upgrade":2}`), &settings); err != nil {
		t.Fatal(err)
	}

	if got := settings.AutoUpgradeLevel(); got != SoftaculousAutoUpgradeMinor {
		t.Errorf("AutoUpgradeLevel() = %q, want %q", got, SoftaculousAutoUpgradeMinor)
	}

	// auto_upgrade_core is write-only: a response carrying it must not be mistaken for the
	// current level.
	var wrong SoftaculousInstallationSettings
	if err := json.Unmarshal([]byte(`{"auto_upgrade_core":2}`), &wrong); err != nil {
		t.Fatal(err)
	}

	// Absent, empty and false all mean off.
	if wrong.AutoUpgradeLevel() != SoftaculousAutoUpgradeOff {
		t.Error("auto_upgrade_core must not be read as the current level")
	}
}
