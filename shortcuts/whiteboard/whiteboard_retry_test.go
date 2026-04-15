// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package whiteboard

import "testing"

func TestIsRetryableWhiteboardReadNotReady(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{
			name: "doc applying",
			msg:  "doc is applying [@from@] resource error",
			want: true,
		},
		{
			name: "data not ready",
			msg:  "doc data is not ready",
			want: true,
		},
		{
			name: "case insensitive",
			msg:  "DOC IS APPLYING",
			want: true,
		},
		{
			name: "other error",
			msg:  "permission denied",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableWhiteboardReadNotReady(tt.msg)
			if got != tt.want {
				t.Fatalf("isRetryableWhiteboardReadNotReady(%q)=%v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}
