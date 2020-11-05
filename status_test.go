// tgbot_ping - status_test
// 2020-11-05 19:50
// Benny <benny.think@gmail.com>

package tgbot_ping

import (
	"testing"
)
import "github.com/stretchr/testify/assert"

func TestGetRuntime(t *testing.T) {
	var result string

	result = getRuntime("laughing_feistel", "display", "markdown")
	assert.Contains(t, result, "display")
	assert.Contains(t, result, "`")

	result = getRuntime("laughing_feistel", "hello_world", "html")
	assert.Contains(t, result, "hello_world")
	assert.Contains(t, result, "<pre>")

	result = getRuntime("no", "hello_world", "html")
	assert.Contains(t, result, "Runtime information")
}
