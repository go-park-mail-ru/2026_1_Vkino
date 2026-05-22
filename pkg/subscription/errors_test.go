package subscription

import (
	"errors"
	"testing"
)

var errPlainSubscription = errors.New("plain error")

func TestToGRPCErrorRoundTripFeatureForbidden(t *testing.T) {
	t.Parallel()

	currentLevel := int32(1)
	requiredLevel := int32(2)
	detail := mustSubscriptionDetail(t, NewFeatureForbidden(
		CodePaidContentForbidden,
		FeaturePaidContent,
		1,
		2,
		"Платный контент недоступен на вашем уровне подписки.",
	))

	assertSubscriptionDetail(t, detail, Detail{
		Code:          CodeSubscriptionFeatureForbidden,
		Feature:       FeaturePaidContent,
		Message:       "Платный контент недоступен на вашем уровне подписки.",
		CurrentLevel:  &currentLevel,
		RequiredLevel: &requiredLevel,
	})
}

func TestToGRPCErrorRoundTripLimitExceeded(t *testing.T) {
	t.Parallel()

	limit := int32(4)
	used := int32(4)
	remaining := int32(0)
	detail := mustSubscriptionDetail(t, NewLimitExceeded(
		CodeRoomMembersLimitExceeded,
		FeatureWatchPartyUsers,
		4,
		4,
		"Лимит активных участников комнаты по подписке владельца исчерпан.",
	))

	assertSubscriptionDetail(t, detail, Detail{
		Code:      CodeRoomMembersLimitExceeded,
		Feature:   FeatureWatchPartyUsers,
		Message:   "Лимит активных участников комнаты по подписке владельца исчерпан.",
		Limit:     &limit,
		Used:      &used,
		Remaining: &remaining,
	})
}

func TestDetailFromGRPCErrorRejectsPlainError(t *testing.T) {
	t.Parallel()

	if _, ok := DetailFromGRPCError(errPlainSubscription); ok {
		t.Fatal("expected plain error to be ignored")
	}
}

func mustSubscriptionDetail(t *testing.T, err error) Detail {
	t.Helper()

	grpcErr, ok := ToGRPCError(err)
	if !ok {
		t.Fatal("expected grpc conversion to succeed")
	}

	detail, ok := DetailFromGRPCError(grpcErr)
	if !ok {
		t.Fatal("expected detail extraction to succeed")
	}

	return detail
}

func assertSubscriptionDetail(t *testing.T, got, want Detail) {
	t.Helper()

	if got.Code != want.Code {
		t.Fatalf("unexpected code: %q", got.Code)
	}

	if got.Feature != want.Feature {
		t.Fatalf("unexpected feature: %q", got.Feature)
	}

	if got.Message != want.Message {
		t.Fatalf("unexpected message: %q", got.Message)
	}

	assertInt32PointerValue(t, "current level", got.CurrentLevel, want.CurrentLevel)
	assertInt32PointerValue(t, "required level", got.RequiredLevel, want.RequiredLevel)
	assertInt32PointerValue(t, "limit", got.Limit, want.Limit)
	assertInt32PointerValue(t, "used", got.Used, want.Used)
	assertInt32PointerValue(t, "remaining", got.Remaining, want.Remaining)
}

func assertInt32PointerValue(t *testing.T, label string, got, want *int32) {
	t.Helper()

	switch {
	case got == nil && want == nil:
		return
	case got == nil || want == nil:
		t.Fatalf("unexpected %s pointer: got %v, want %v", label, got, want)
	case *got != *want:
		t.Fatalf("unexpected %s value: got %d, want %d", label, *got, *want)
	}
}
