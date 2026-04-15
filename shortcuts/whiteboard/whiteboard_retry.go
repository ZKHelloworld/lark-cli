// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package whiteboard

import (
	"strings"
	"time"
)

// Tunable in tests.
var whiteboardReadRetryMax = 5
var whiteboardReadRetryInterval = 300 * time.Millisecond

func isRetryableWhiteboardReadNotReady(msg string) bool {
	m := strings.ToLower(msg)
	return strings.Contains(m, "doc is applying") || strings.Contains(m, "data is not ready")
}
