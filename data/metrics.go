package data

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/ArcCS/Nevermore/config"
	_ "modernc.org/sqlite"
)

type CombatMetric struct {
	Action       string
	ActionType   int // damage 0, heal 1
	Mode         int // melee 0, ranged 1, spell 2, item 3
	TotalDamage  int
	Resisted     int
	FinalDamage  int
	AttackerType int // player 0, mob 1, npc 2
	AttackerId   int
	AttackerTier int
	VictimType   int // player 0, mob 1, npc 2
	VictimId     int
	CombatTime   time.Time
}

type ChatLog struct {
	ChatType int // say 0 osay 1 sent 2 act 3	gmsay 4, ptell 5
	FromId   int
	ToId     int
	Message  string
	ChatTime time.Time
}

type ItemTotals struct {
	ItemId     int
	TotalSold  int
	TotalValue int
	LastSold   time.Time
}

type ItemSales struct {
	ItemId     int
	TimeSold   time.Time
	SellerId   int
	SellerTier int
	SellValue  int
}

// Times are stored as UTC text so SQLite's date functions and plain string
// comparisons both work on them.
const metricsTimeFormat = "2006-01-02 15:04:05.000"

const metricsSchema = `
CREATE TABLE IF NOT EXISTS combat_metrics (
	action        TEXT,
	type          INTEGER,
	mode          INTEGER,
	total_damage  INTEGER,
	resisted      INTEGER,
	final_damage  INTEGER,
	attacker_type INTEGER,
	attacker_id   INTEGER,
	attacker_tier INTEGER,
	victim_type   INTEGER,
	victim_id     INTEGER,
	time          TEXT
);
CREATE INDEX IF NOT EXISTS combat_metrics_time ON combat_metrics (time);
CREATE INDEX IF NOT EXISTS combat_metrics_action ON combat_metrics (action);

CREATE TABLE IF NOT EXISTS chat (
	type     INTEGER,
	from_id  INTEGER,
	to_id    INTEGER,
	contents TEXT,
	time     TEXT
);
CREATE INDEX IF NOT EXISTS chat_time ON chat (time);

CREATE TABLE IF NOT EXISTS item_sales (
	item_id     INTEGER,
	time        TEXT,
	seller_id   INTEGER,
	seller_tier INTEGER,
	sell_value  INTEGER
);
CREATE INDEX IF NOT EXISTS item_sales_item ON item_sales (item_id);
`

var (
	metricsDB *sql.DB
	metricsMu sync.Mutex // guards metricsDB
	captureMu sync.Mutex // guards the *Capture buffers
)

// openMetrics lazily opens the metrics SQLite file and creates its tables.
// It returns nil if the file can't be opened; callers keep their buffers and
// retry on the next flush.
func openMetrics() *sql.DB {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	if metricsDB != nil {
		return metricsDB
	}
	db, err := sql.Open("sqlite", config.Server.MetricsDB+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		log.Println("Error opening metrics db:", err)
		return nil
	}
	// SQLite allows one writer; a single connection avoids "database is locked".
	db.SetMaxOpenConns(1)
	if _, err = db.Exec(metricsSchema); err != nil {
		log.Println("Error creating metrics tables:", err)
		_ = db.Close()
		return nil
	}
	metricsDB = db
	return metricsDB
}

// insertBatch writes n rows with one prepared statement inside a single
// transaction, so a flush either lands completely or not at all.
func insertBatch(query string, n int, row func(i int) []any) bool {
	db := openMetrics()
	if db == nil {
		return false
	}
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return false
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Println(err)
		_ = tx.Rollback()
		return false
	}
	defer stmt.Close()
	for i := 0; i < n; i++ {
		if _, err = stmt.Exec(row(i)...); err != nil {
			log.Println(err)
			_ = tx.Rollback()
			return false
		}
	}
	if err = tx.Commit(); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func metricsTime(t time.Time) string {
	return t.UTC().Format(metricsTimeFormat)
}

func StoreChatLog(chatType int, fromId int, toId int, message string) {
	captureMu.Lock()
	defer captureMu.Unlock()
	ChatLogsCapture = append(ChatLogsCapture, ChatLog{ChatType: chatType, FromId: fromId, ToId: toId, Message: message, ChatTime: time.Now()})
}

func StoreItemSale(ItemId int, SellerId int, SellerTier int, SellValue int) {
	captureMu.Lock()
	defer captureMu.Unlock()
	ItemSalesCapture = append(ItemSalesCapture, ItemSales{ItemId: ItemId, TimeSold: time.Now(), SellerId: SellerId, SellerTier: SellerTier, SellValue: SellValue})
}

func StoreCombatMetric(Action string, ActionType int, Mode int, TotalDamage int, Resisted int, FinalDamage int, AttackerType int, AttackerId int, AttackerTier int, VictimType int, VictimId int) {
	captureMu.Lock()
	defer captureMu.Unlock()
	CombatMetricsCapture = append(CombatMetricsCapture, CombatMetric{Action: Action, ActionType: ActionType, Mode: Mode, TotalDamage: TotalDamage, Resisted: Resisted, FinalDamage: FinalDamage, AttackerType: AttackerType, AttackerId: AttackerId, AttackerTier: AttackerTier, VictimType: VictimType, VictimId: VictimId, CombatTime: time.Now()})
}

// Each flush swaps the buffer out under the lock so the game can keep
// recording while the write runs; on failure the batch is put back in front
// of anything recorded since, to retry next time.

func FlushCombatMetrics() bool {
	captureMu.Lock()
	batch := CombatMetricsCapture
	CombatMetricsCapture = make([]CombatMetric, 0)
	captureMu.Unlock()
	if len(batch) == 0 {
		return false
	}

	ok := insertBatch("INSERT INTO combat_metrics (action, type, mode, total_damage, resisted, final_damage, attacker_type, attacker_id, attacker_tier, victim_type, victim_id, time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		len(batch), func(i int) []any {
			m := batch[i]
			return []any{m.Action, m.ActionType, m.Mode, m.TotalDamage, m.Resisted, m.FinalDamage,
				m.AttackerType, m.AttackerId, m.AttackerTier, m.VictimType, m.VictimId, metricsTime(m.CombatTime)}
		})
	if !ok {
		captureMu.Lock()
		CombatMetricsCapture = append(batch, CombatMetricsCapture...)
		captureMu.Unlock()
	}
	return ok
}

func FlushChatLogs() bool {
	captureMu.Lock()
	batch := ChatLogsCapture
	ChatLogsCapture = make([]ChatLog, 0)
	captureMu.Unlock()
	if len(batch) == 0 {
		return false
	}

	log.Println("Running Chatlog Flush")
	ok := insertBatch("INSERT INTO chat (type, from_id, to_id, contents, time) VALUES (?, ?, ?, ?, ?)",
		len(batch), func(i int) []any {
			c := batch[i]
			return []any{c.ChatType, c.FromId, c.ToId, c.Message, metricsTime(c.ChatTime)}
		})
	if !ok {
		captureMu.Lock()
		ChatLogsCapture = append(batch, ChatLogsCapture...)
		captureMu.Unlock()
	}
	return ok
}

func FlushItemSales() bool {
	captureMu.Lock()
	batch := ItemSalesCapture
	ItemSalesCapture = make([]ItemSales, 0)
	captureMu.Unlock()
	if len(batch) == 0 {
		return false
	}

	ok := insertBatch("INSERT INTO item_sales (item_id, time, seller_id, seller_tier, sell_value) VALUES (?, ?, ?, ?, ?)",
		len(batch), func(i int) []any {
			s := batch[i]
			return []any{s.ItemId, metricsTime(s.TimeSold), s.SellerId, s.SellerTier, s.SellValue}
		})
	if !ok {
		captureMu.Lock()
		ItemSalesCapture = append(batch, ItemSalesCapture...)
		captureMu.Unlock()
	}
	return ok
}
