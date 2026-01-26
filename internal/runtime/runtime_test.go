package runtime

import "testing"

func TestWasmExecDTSEmbed(t *testing.T) {
	if WasmExecDTS == "" {
		t.Error("WasmExecDTS should not be empty - embed failed")
	}

	// Verify it contains expected TypeScript content
	if len(WasmExecDTS) < 100 {
		t.Errorf("WasmExecDTS seems too short: %d bytes", len(WasmExecDTS))
	}
}
