package metrics

import "testing"

func TestLabelValue(t *testing.T) {
	t.Parallel()

	if labelValue("") != unknownLabelValue {
		t.Fatal("expected unknown label")
	}

	if labelValue("svc") != "svc" {
		t.Fatal("expected service label")
	}
}

func TestSetServiceInfo(t *testing.T) {
	t.Parallel()

	SetServiceInfo("svc")
}
