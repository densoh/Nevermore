package data

import (
	"path/filepath"
	"testing"

	"github.com/ArcCS/Nevermore/config"
)

func TestMetricsFlushToSQLite(t *testing.T) {
	config.Server.MetricsDB = filepath.Join(t.TempDir(), "metrics.db")
	metricsDB = nil
	t.Cleanup(func() {
		if metricsDB != nil {
			_ = metricsDB.Close()
			metricsDB = nil
		}
	})

	StoreCombatMetric("bash", 0, 0, 30, 5, 25, 0, 7, 10, 1, 42)
	StoreCombatMetric("bash-miss", 0, 0, 0, 0, 0, 0, 7, 10, 1, 42)
	StoreChatLog(0, 7, 0, "hello")
	StoreItemSale(99, 7, 10, 500)

	if !FlushCombatMetrics() || !FlushChatLogs() || !FlushItemSales() {
		t.Fatal("flush failed")
	}
	if len(CombatMetricsCapture) != 0 || len(ChatLogsCapture) != 0 || len(ItemSalesCapture) != 0 {
		t.Fatal("buffers not cleared after a successful flush")
	}
	if FlushCombatMetrics() {
		t.Fatal("flushing an empty buffer should report false")
	}

	var n, dmg int
	if err := metricsDB.QueryRow("SELECT count(*), sum(final_damage) FROM combat_metrics WHERE attacker_id = 7").Scan(&n, &dmg); err != nil {
		t.Fatal(err)
	}
	if n != 2 || dmg != 25 {
		t.Fatalf("combat_metrics: got %d rows / %d dmg, want 2 / 25", n, dmg)
	}
	var msg string
	if err := metricsDB.QueryRow("SELECT contents FROM chat").Scan(&msg); err != nil || msg != "hello" {
		t.Fatalf("chat: got %q, %v", msg, err)
	}
	var value int
	if err := metricsDB.QueryRow("SELECT sell_value FROM item_sales WHERE item_id = 99 AND time > datetime('now', '-1 minute')").Scan(&value); err != nil || value != 500 {
		t.Fatalf("item_sales: got %d, %v", value, err)
	}
}

func TestMetricsFlushFailureKeepsBuffer(t *testing.T) {
	// A path inside a missing directory can't be opened.
	config.Server.MetricsDB = filepath.Join(t.TempDir(), "missing", "metrics.db")
	metricsDB = nil
	CombatMetricsCapture = CombatMetricsCapture[:0]

	StoreCombatMetric("bash", 0, 0, 30, 5, 25, 0, 7, 10, 1, 42)
	if FlushCombatMetrics() {
		t.Fatal("flush should fail when the file can't be opened")
	}
	if len(CombatMetricsCapture) != 1 {
		t.Fatalf("failed flush should keep the batch, have %d", len(CombatMetricsCapture))
	}
	CombatMetricsCapture = CombatMetricsCapture[:0]
}
