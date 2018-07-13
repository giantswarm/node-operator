package drainer

import (
	"testing"
	"time"

	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Resource_Drainer_drainingTimedOut(t *testing.T) {
	testCases := []struct {
		Name           string
		DrainerConfig  corev1alpha1.DrainerConfig
		Now            time.Time
		Timeout        time.Duration
		ExpectedResult bool
	}{
		{
			Name: "case 0: one second before the timeout should happen the drainer config should not be timed out",
			DrainerConfig: corev1alpha1.DrainerConfig{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(time.Unix(10, 0)),
				},
			},
			Now:            time.Unix(19, 0),
			Timeout:        10 * time.Second,
			ExpectedResult: false,
		},
		{
			Name: "case 1: at the time the drainer config hits the timeout boundary it should be timed out",
			DrainerConfig: corev1alpha1.DrainerConfig{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(time.Unix(10, 0)),
				},
			},
			Now:            time.Unix(20, 0),
			Timeout:        10 * time.Second,
			ExpectedResult: true,
		},
		{
			Name: "case 2: one second after the timeout should happen the drainer config should be timed out",
			DrainerConfig: corev1alpha1.DrainerConfig{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(time.Unix(10, 0)),
				},
			},
			Now:            time.Unix(21, 0),
			Timeout:        10 * time.Second,
			ExpectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			result := drainingTimedOut(tc.DrainerConfig, tc.Now, tc.Timeout)

			if result != tc.ExpectedResult {
				t.Fatalf("expected %#v got %#v", tc.ExpectedResult, result)
			}
		})
	}
}
