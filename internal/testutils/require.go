package testutils

import (
	"cmp"
	"slices"
	"testing"
)

func RequireNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("unexpected error: [%T] %v\n", err, err)
	}
}

func RequireEqual[T comparable](t *testing.T, expected T, actual T) {
	t.Helper()

	if expected != actual {
		t.Errorf("expected equal:\nexpected=%v\nactual=%v\n", expected, actual)
	}
}

func RequireElementsMatch[T cmp.Ordered](t *testing.T, expected []T, actual []T) {
	t.Helper()

	expectedCopy := slices.Clone(expected)
	actualCopy := slices.Clone(actual)

	slices.Sort(expectedCopy)
	slices.Sort(actualCopy)

	if slices.Compare(expectedCopy, actualCopy) != 0 {
		t.Errorf("expected same elements:\nexpected=%v\nactual=%v\n", expected, actual)
	}
}
