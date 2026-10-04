package objects

import (
	"encoding/json"
	"errors"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/data"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/utils"
	"github.com/jinzhu/copier"
)

// A QuestEvent is a time-limited realm event a DM sets up with the EVENT
// command. It can be started by hand or at Starts, and runs until Expires.
// While it runs:
//
//   - its Rooms are in quest mode (see Room.InQuestMode),
//   - its spawn groups add their own mobs on top of each room's normal
//     encounters (see Room.EventEncounter),
//   - everyone who logs in is shown the message of the latest stage reached
//     that has one, or Message.
//
// An event runs in stages, starting at 1. A stage can bring more rooms into
// quest mode, unseal exits and switch on spawn groups set to it. A stage is
// reached at its At time, by a DM, or by a script ($EVENTSTAGE, including a
// mob's @DEATH script) - whichever comes first.
//
// A stage can also redirect exits: while the event is at that stage or later
// the exit leads somewhere else, and it leads back where it always did the
// moment the event ends. Nothing about the exit itself is changed or saved,
// so there is nothing to put back. A later stage can re-point the same exit.
//
// In preview the event runs for staff only: spawns roll only where staff
// stand, only staff see its messages, quest mode is off and its exits stay
// sealed to players. It lets a DM walk the event before players can.
//
// When it ends - Expires passes or a DM stops it - every copy of its Mobs
// and every mob its spawn groups made is removed from the world, along with
// permanent ground items (fixtures, not loot) whose id is in Items.
type QuestEvent struct {
	Name    string
	Message string
	Starts  time.Time // start automatically at this time; cleared once started
	Expires time.Time
	Active  bool // started and not yet ended; Running also checks Expires
	Preview bool // running for staff only
	Stage   int  // the stage reached while running

	Rooms  []int
	Mobs   []int
	Items  []int
	Spawns map[string]*SpawnGroup
	Stages map[int]*Stage
}

// A Stage is what reaching a stage of an event changes.
type Stage struct {
	At      time.Time  // reach this stage automatically at this time
	Message string     // broadcast when reached; the login message from then on
	Rooms   []int      // rooms that join quest mode from this stage
	Exits   []ExitGate // exits sealed to players until this stage

	Redirects []ExitRedirect // exits that lead elsewhere from this stage
}

// An ExitRedirect sends one exit to a different room while it applies.
type ExitRedirect struct {
	Room int
	Exit string
	To   int
}

// An ExitGate names one exit, by its room and lowercased exit name.
type ExitGate struct {
	Room int
	Exit string
}

// A SpawnGroup rolls for an event mob in each of its rooms every Interval
// seconds while a player is there, with a Rate% chance, from Stage on. Mobs
// maps mob id to a weight for picking which one.
type SpawnGroup struct {
	Rooms    []int
	Mobs     map[int]int
	Rate     int
	Interval int
	Stage    int
}

// DefaultSpawnInterval is the seconds between a spawn group's rolls in a
// room when the group doesn't set one.
const DefaultSpawnInterval = 60

func (g *SpawnGroup) interval() time.Duration {
	if g.Interval <= 0 {
		return DefaultSpawnInterval * time.Second
	}
	return time.Duration(g.Interval) * time.Second
}

// Running reports whether the event is in effect at now, for players or,
// in preview, for staff.
func (e *QuestEvent) Running(now time.Time) bool {
	return e.Active && now.Before(e.Expires)
}

// liveFor reports whether the event is in effect for c: running, and if it
// is only a preview, c is staff.
func (e *QuestEvent) liveFor(c *Character, now time.Time) bool {
	return e.Running(now) && (!e.Preview || isStaff(c))
}

// livePublic reports whether the event is in effect for players.
func (e *QuestEvent) livePublic(now time.Time) bool {
	return e.Running(now) && !e.Preview
}

// stageNumbers lists the defined stages in order.
func (e *QuestEvent) stageNumbers() []int {
	numbers := make([]int, 0, len(e.Stages))
	for n := range e.Stages {
		numbers = append(numbers, n)
	}
	sort.Ints(numbers)
	return numbers
}

// message is the latest reached stage's message, or the event's.
func (e *QuestEvent) message() string {
	msg := e.Message
	for _, n := range e.stageNumbers() {
		if n <= e.Stage && e.Stages[n].Message != "" {
			msg = e.Stages[n].Message
		}
	}
	return msg
}

func isStaff(c *Character) bool {
	return c != nil && c.Permission.HasAnyFlags(permissions.Builder, permissions.Dungeonmaster, permissions.Gamemaster)
}

// Lock order: a room lock may be held when taking questEventsMu, never the
// other way round. Room ticks check spawns and combat checks quest mode with
// their room locked, so anything that locks rooms (EndQuestEvent's sweep)
// must have let go of questEventsMu first. Broadcasts also wait until it is
// released.
var (
	questEventsMu sync.RWMutex
	questEvents   = map[string]*QuestEvent{}
	// questRooms maps each room in quest mode through a public event to the
	// latest expiry among the events that put it there.
	questRooms = map[int]time.Time{}
	// exitGates maps "room/exit" to the events and stages sealing it.
	exitGates = map[string][]exitGateRef{}
	// exitRedirects maps "room/exit" to the events and stages redirecting
	// it, in event name order.
	exitRedirects = map[string][]exitRedirectRef{}
	// timedExits maps "room/exit" to a destination a script gave it for a
	// while. They live in memory only, so a reboot puts every exit back.
	timedExits = map[string]timedExit{}
	// eventSpawnRolls is when each event/group/room last rolled to spawn.
	eventSpawnRolls = map[string]time.Time{}
)

type exitRedirectRef struct {
	event string
	stage int
	to    int
}

type timedExit struct {
	to    int
	until time.Time
}

type exitGateRef struct {
	event string
	stage int
}

func gateKey(room int, exit string) string {
	return strconv.Itoa(room) + "/" + strings.ToLower(exit)
}

// reindexQuestEvents rebuilds questRooms and exitGates. Callers hold
// questEventsMu for writing.
func reindexQuestEvents() {
	questRooms = map[int]time.Time{}
	exitGates = map[string][]exitGateRef{}
	exitRedirects = map[string][]exitRedirectRef{}
	for _, name := range sortedEventNames() {
		for n, stage := range questEvents[name].Stages {
			for _, rd := range stage.Redirects {
				key := gateKey(rd.Room, rd.Exit)
				exitRedirects[key] = append(exitRedirects[key], exitRedirectRef{name, n, rd.To})
			}
		}
	}
	for name, e := range questEvents {
		for n, stage := range e.Stages {
			for _, gate := range stage.Exits {
				key := gateKey(gate.Room, gate.Exit)
				exitGates[key] = append(exitGates[key], exitGateRef{name, n})
			}
		}
		if !e.Active || e.Preview {
			continue
		}
		rooms := append([]int(nil), e.Rooms...)
		for n, stage := range e.Stages {
			if n <= e.Stage {
				rooms = append(rooms, stage.Rooms...)
			}
		}
		for _, r := range rooms {
			if e.Expires.After(questRooms[r]) {
				questRooms[r] = e.Expires
			}
		}
	}
}

// questEventSave and questEventDelete persist events. Tests swap them out
// with SetQuestEventStore.
var (
	questEventSave   = data.SaveQuestEvent
	questEventDelete = data.DeleteQuestEvent
)

// SetQuestEventStore replaces where events are saved and deleted, and
// returns a function that puts the database back. It is for tests.
func SetQuestEventStore(save func(name, definition string) bool, del func(name string) bool) (restore func()) {
	oldSave, oldDelete := questEventSave, questEventDelete
	questEventSave, questEventDelete = save, del
	return func() { questEventSave, questEventDelete = oldSave, oldDelete }
}

func saveQuestEvent(e *QuestEvent) error {
	definition, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if !questEventSave(e.Name, string(definition)) {
		return errors.New("the database refused the save, see the log")
	}
	return nil
}

// LoadQuestEvents reads every event from the database. Anything due while
// the server was down - an expiry, a start, a stage - happens on the next
// CheckQuestEvents.
func LoadQuestEvents() {
	loaded := map[string]*QuestEvent{}
	for name, definition := range data.LoadQuestEvents() {
		e := &QuestEvent{}
		if err := json.Unmarshal([]byte(definition), e); err != nil {
			log.Println("Error loading quest event "+name+": ", err)
			continue
		}
		e.Name = name
		if e.Spawns == nil {
			e.Spawns = map[string]*SpawnGroup{}
		}
		if e.Stages == nil {
			e.Stages = map[int]*Stage{}
		}
		loaded[name] = e
	}
	questEventsMu.Lock()
	questEvents = loaded
	reindexQuestEvents()
	questEventsMu.Unlock()
	log.Printf("Finished loading %d quest events.", len(loaded))
}

// QuestEventNames lists every event, sorted.
func QuestEventNames() []string {
	questEventsMu.RLock()
	defer questEventsMu.RUnlock()
	return sortedEventNames()
}

// sortedEventNames lists events in a fixed order. Callers hold questEventsMu.
func sortedEventNames() []string {
	names := make([]string, 0, len(questEvents))
	for name := range questEvents {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// QuestEventCopy returns a snapshot of the named event for display.
func QuestEventCopy(name string) (QuestEvent, bool) {
	questEventsMu.RLock()
	defer questEventsMu.RUnlock()
	e, ok := questEvents[strings.ToLower(name)]
	if !ok {
		return QuestEvent{}, false
	}
	// A JSON round trip, not copier: copier's deep copy zeroes time.Time.
	snapshot := QuestEvent{}
	raw, err := json.Marshal(e)
	if err != nil || json.Unmarshal(raw, &snapshot) != nil {
		return QuestEvent{}, false
	}
	return snapshot, true
}

// CreateQuestEvent adds a new, inactive event.
func CreateQuestEvent(name string) error {
	name = strings.ToLower(name)
	questEventsMu.Lock()
	defer questEventsMu.Unlock()
	if _, ok := questEvents[name]; ok {
		return errors.New("an event with that name already exists")
	}
	e := &QuestEvent{Name: name, Spawns: map[string]*SpawnGroup{}, Stages: map[int]*Stage{}}
	if err := saveQuestEvent(e); err != nil {
		return err
	}
	questEvents[name] = e
	return nil
}

// EditQuestEvent applies edit to the named event and saves it. If edit
// returns an error nothing is saved, but edit must not have half-changed
// the event before failing.
func EditQuestEvent(name string, edit func(e *QuestEvent) error) error {
	questEventsMu.Lock()
	defer questEventsMu.Unlock()
	e, ok := questEvents[strings.ToLower(name)]
	if !ok {
		return errors.New("there is no event by that name")
	}
	if err := edit(e); err != nil {
		return err
	}
	reindexQuestEvents()
	return saveQuestEvent(e)
}

// DeleteQuestEvent removes an event that isn't running.
func DeleteQuestEvent(name string) error {
	name = strings.ToLower(name)
	questEventsMu.Lock()
	defer questEventsMu.Unlock()
	e, ok := questEvents[name]
	if !ok {
		return errors.New("there is no event by that name")
	}
	if e.Active {
		return errors.New("stop the event before deleting it")
	}
	if !questEventDelete(name) {
		return errors.New("the database refused the delete, see the log")
	}
	delete(questEvents, name)
	reindexQuestEvents()
	return nil
}

// broadcastQuestEvent sends msg to everyone, or only to staff in a preview.
// Never call it holding questEventsMu.
func broadcastQuestEvent(preview bool, msg string) {
	if msg == "" {
		return
	}
	if preview {
		ActiveCharacters.MessageGM("[EVENT PREVIEW] " + msg)
		return
	}
	ActiveCharacters.MessageAll(msg, config.BroadcastChannel)
}

// StartQuestEvent switches an event on at stage 1, for everyone or, in
// preview, for staff only, and announces it.
func StartQuestEvent(name string, preview bool) error {
	var msg string
	err := EditQuestEvent(name, func(e *QuestEvent) error {
		if e.Active {
			if e.Preview {
				return errors.New("the event is in preview; stop it before starting it for real")
			}
			return errors.New("the event is already running")
		}
		if !time.Now().Before(e.Expires) {
			return errors.New("set an expiration date in the future first")
		}
		e.Active, e.Preview, e.Stage = true, preview, 1
		if !preview {
			e.Starts = time.Time{}
		}
		msg = e.message()
		return nil
	})
	if err == nil {
		broadcastQuestEvent(preview, msg)
	}
	return err
}

// AdvanceQuestEvent moves a running event to stage and announces every
// stage newly reached on the way. A DM may move it back (force); otherwise
// only forward moves happen, and only when the event is live for actor -
// so a player's script can't advance a preview. It reports whether the
// stage changed.
func AdvanceQuestEvent(name string, stage int, actor *Character, force bool) (bool, error) {
	var preview bool
	var announce []string
	changed := false
	err := EditQuestEvent(name, func(e *QuestEvent) error {
		now := time.Now()
		if !e.Running(now) || (!force && !e.liveFor(actor, now)) {
			return errors.New("the event is not running")
		}
		if stage < 1 {
			return errors.New("stages start at 1")
		}
		if stage == e.Stage || (!force && stage < e.Stage) {
			return nil
		}
		for _, n := range e.stageNumbers() {
			if n > e.Stage && n <= stage {
				announce = append(announce, e.Stages[n].Message)
			}
		}
		e.Stage, preview, changed = stage, e.Preview, true
		return nil
	})
	for _, msg := range announce {
		broadcastQuestEvent(preview, msg)
	}
	return changed, err
}

// EndQuestEvent switches an event off and removes its mobs and fixtures from
// the world. It locks every room in turn, so call it without any room lock
// held - from a command, run it on its own goroutine.
func EndQuestEvent(name string) error {
	var mobIds, itemIds []int
	var preview bool
	err := EditQuestEvent(name, func(e *QuestEvent) error {
		if !e.Active {
			return errors.New("the event is not running")
		}
		preview = e.Preview
		e.Active, e.Preview, e.Stage = false, false, 0
		mobIds = append(mobIds, e.Mobs...)
		itemIds = append(itemIds, e.Items...)
		for key := range eventSpawnRolls {
			if strings.HasPrefix(key, e.Name+"/") {
				delete(eventSpawnRolls, key)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	sweepQuestEvent(strings.ToLower(name), mobIds, itemIds)
	broadcastQuestEvent(preview, "The "+name+" event has come to an end.")
	return nil
}

// sweepQuestEvent removes an ended event's mobs and fixtures from every room.
func sweepQuestEvent(name string, mobIds, itemIds []int) {
	for id, r := range Rooms {
		if r == nil || r.Mobs == nil || r.Items == nil {
			continue
		}
		r.LockRoom("QuestEventEnd", true)
		changed := false
		for _, m := range append([]*Mob(nil), r.Mobs.Contents...) {
			if m.QuestEvent == name || utils.IntIn(m.MobId, mobIds) {
				r.Mobs.Remove(m)
				changed = true
			}
		}
		for _, i := range append([]*Item(nil), r.Items.Contents...) {
			if i.Flags["permanent"] && utils.IntIn(i.ItemId, itemIds) {
				if err := r.Items.Remove(i); err != nil {
					log.Println("Error removing quest event item: ", err)
				}
				changed = true
			}
		}
		r.UnlockRoom("QuestEventEnd", true)
		if changed {
			log.Println("Quest event " + name + " cleaned up room " + strconv.Itoa(id))
			AddRoomUpdate(id)
		}
	}
}

// CheckQuestEvents does whatever is due: ends expired events, starts
// scheduled ones and reaches stages whose time has come. The world ticker
// calls it once a minute.
func CheckQuestEvents() {
	now := time.Now()
	var expired, starting []string
	advance := map[string]int{}
	questEventsMu.RLock()
	for name, e := range questEvents {
		switch {
		case e.Active && !e.Running(now):
			expired = append(expired, name)
		case !e.Active && !e.Starts.IsZero() && !now.Before(e.Starts) && now.Before(e.Expires):
			starting = append(starting, name)
		case e.Running(now):
			for _, n := range e.stageNumbers() {
				if at := e.Stages[n].At; n > e.Stage && !at.IsZero() && !now.Before(at) {
					advance[name] = n
				}
			}
		}
	}
	questEventsMu.RUnlock()

	for _, name := range expired {
		if err := EndQuestEvent(name); err != nil {
			log.Println("Error ending quest event "+name+": ", err)
		}
	}
	for _, name := range starting {
		if err := StartQuestEvent(name, false); err != nil {
			log.Println("Error starting quest event "+name+": ", err)
		}
	}
	for name, stage := range advance {
		if _, err := AdvanceQuestEvent(name, stage, nil, true); err != nil {
			log.Println("Error advancing quest event "+name+": ", err)
		}
	}
}

// QuestEventRoom reports whether a public event puts the room in quest mode.
func QuestEventRoom(roomId int) bool {
	questEventsMu.RLock()
	defer questEventsMu.RUnlock()
	expires, ok := questRooms[roomId]
	return ok && time.Now().Before(expires)
}

// QuestExitSealed reports whether an event keeps this exit shut to players:
// it is sealed until its event is running for players and has reached its
// stage, and sealed again once the event ends.
func QuestExitSealed(roomId int, exit string) bool {
	questEventsMu.RLock()
	defer questEventsMu.RUnlock()
	now := time.Now()
	for _, gate := range exitGates[gateKey(roomId, exit)] {
		e := questEvents[gate.event]
		if e == nil || !e.livePublic(now) || e.Stage < gate.stage {
			return true
		}
	}
	return false
}

// QuestExitTo returns where this exit temporarily leads for c, if anywhere
// but its own destination. A timed redirect from a script wins; otherwise
// it is the latest stage reached of the first live event, by name, that
// redirects it. A nil c sees what players see.
func QuestExitTo(roomId int, exit string, c *Character) (int, bool) {
	key := gateKey(roomId, exit)
	now := time.Now()
	questEventsMu.RLock()
	defer questEventsMu.RUnlock()

	to, found := 0, false
	if timed, ok := timedExits[key]; ok && now.Before(timed.until) {
		to, found = timed.to, true
	} else {
		event, stage := "", 0
		for _, ref := range exitRedirects[key] {
			e := questEvents[ref.event]
			if e == nil || !e.liveFor(c, now) || ref.stage > e.Stage {
				continue
			}
			if event != "" && ref.event != event {
				break
			}
			if ref.stage > stage {
				event, stage, to, found = ref.event, ref.stage, ref.to, true
			}
		}
	}
	if !found {
		return 0, false
	}
	// A destination deleted since falls back to the exit's own.
	if _, ok := Rooms[to]; !ok {
		return 0, false
	}
	return to, true
}

// RedirectExitFor makes an exit lead to another room for d, after which it
// leads where it always did. A d of zero or less puts it back now.
func RedirectExitFor(roomId int, exit string, to int, d time.Duration) {
	questEventsMu.Lock()
	defer questEventsMu.Unlock()
	now := time.Now()
	for key, timed := range timedExits {
		if !now.Before(timed.until) {
			delete(timedExits, key)
		}
	}
	if d <= 0 {
		delete(timedExits, gateKey(roomId, exit))
		return
	}
	timedExits[gateKey(roomId, exit)] = timedExit{to, now.Add(d)}
}

// ParseExitTo reads the arguments of $EXITTO: room_id exit to_room_id
// seconds, where the exit name may be several words. Seconds of 0 puts the
// exit back.
func ParseExitTo(args []string) (roomId int, exit string, to int, d time.Duration, ok bool) {
	if len(args) < 4 {
		return
	}
	roomId, err1 := strconv.Atoi(args[0])
	to, err2 := strconv.Atoi(args[len(args)-2])
	secs, err3 := strconv.Atoi(args[len(args)-1])
	exit = strings.ToLower(strings.Join(args[1:len(args)-2], " "))
	if err1 != nil || err2 != nil || err3 != nil || secs < 0 || secs > MaxExitToSeconds {
		return 0, "", 0, 0, false
	}
	room, exists := Rooms[roomId]
	if !exists {
		return 0, "", 0, 0, false
	}
	if _, exists = room.Exits[exit]; !exists {
		return 0, "", 0, 0, false
	}
	if _, exists = Rooms[to]; !exists {
		return 0, "", 0, 0, false
	}
	return roomId, exit, to, time.Duration(secs) * time.Second, true
}

// MaxExitToSeconds is the longest a script may redirect an exit: a day.
const MaxExitToSeconds = 86400

// QuestEventStage returns an event's stage and whether it is live for c.
func QuestEventStage(name string, c *Character) (int, bool) {
	questEventsMu.RLock()
	defer questEventsMu.RUnlock()
	e, ok := questEvents[strings.ToLower(name)]
	if !ok || !e.liveFor(c, time.Now()) {
		return 0, false
	}
	return e.Stage, true
}

// QuestEventStageIn returns an event's stage and whether it is live in a
// room: running, and if only a preview, with staff standing there. It is how
// mob scripts, which have no player to ask, see events.
func QuestEventStageIn(name string, room *Room) (int, bool) {
	staffHere := false
	if room != nil && room.Chars != nil {
		for _, c := range room.Chars.Contents {
			staffHere = staffHere || isStaff(c)
		}
	}
	questEventsMu.RLock()
	defer questEventsMu.RUnlock()
	e, ok := questEvents[strings.ToLower(name)]
	if !ok || !e.Running(time.Now()) || (e.Preview && !staffHere) {
		return 0, false
	}
	return e.Stage, true
}

// QuestEventMessages returns the login messages of every event live for c.
func QuestEventMessages(c *Character) []string {
	now := time.Now()
	questEventsMu.RLock()
	defer questEventsMu.RUnlock()
	var messages []string
	for _, name := range sortedEventNames() {
		e := questEvents[name]
		if !e.liveFor(c, now) {
			continue
		}
		if msg := e.message(); msg != "" {
			if e.Preview {
				msg = "[EVENT PREVIEW] " + msg
			}
			messages = append(messages, msg)
		}
	}
	return messages
}

// eventSpawn is one mob an event spawn group has decided to put in a room.
type eventSpawn struct {
	event string
	mobId int
}

// EventEncounter rolls each running event's spawn groups that include this
// room. Event mobs come on top of the room's own encounters, but the room's
// mob cap and crowding still apply. Call it with the room locked.
func (r *Room) EventEncounter() {
	for _, spawn := range r.eventSpawnRolls(time.Now()) {
		newMob := Mob{}
		if err := copier.CopyWithOption(&newMob, Mobs[spawn.mobId], copier.Option{DeepCopy: true}); err != nil {
			log.Println("Error copying mob during event encounter: ", err)
			continue
		}
		if newMob.Placement <= 0 {
			newMob.Placement = 5
		} else if newMob.Placement >= 6 {
			newMob.Placement = utils.Roll(5, 1, 0)
		}
		newMob.QuestEvent = spawn.event
		r.Mobs.Add(&newMob, false)
		newMob.StartTicking()
	}
}

// eventSpawnRolls makes every spawn roll due in this room at now and returns
// the mobs that came up.
func (r *Room) eventSpawnRolls(now time.Time) []eventSpawn {
	if len(r.Mobs.Contents) >= 10 {
		return nil
	}
	_, crowdPct := EncounterCrowding(len(r.Chars.Contents), len(r.Mobs.ListHostile()))
	staffHere := false
	for _, c := range r.Chars.Contents {
		staffHere = staffHere || isStaff(c)
	}

	var spawns []eventSpawn
	questEventsMu.Lock()
	defer questEventsMu.Unlock()
	for _, name := range sortedEventNames() {
		e := questEvents[name]
		if !e.Running(now) || (e.Preview && !staffHere) {
			continue
		}
		for groupName, g := range e.Spawns {
			if !utils.IntIn(r.RoomId, g.Rooms) || len(g.Mobs) == 0 || g.Rate <= 0 || g.Stage > e.Stage {
				continue
			}
			key := e.Name + "/" + groupName + "/" + strconv.Itoa(r.RoomId)
			if now.Sub(eventSpawnRolls[key]) < g.interval() {
				continue
			}
			eventSpawnRolls[key] = now
			if utils.Roll(100, 1, 0) > g.Rate*crowdPct/100 {
				continue
			}
			if mobId, ok := pickEventMob(g.Mobs); ok {
				spawns = append(spawns, eventSpawn{e.Name, mobId})
			}
		}
	}
	return spawns
}

// pickEventMob picks a mob by weight from those out at this time of day.
func pickEventMob(weights map[int]int) (int, bool) {
	var ids []int
	total := 0
	for id, weight := range weights {
		m, ok := Mobs[id]
		if !ok || weight <= 0 {
			continue
		}
		if (DayTime && m.Flags["night_only"]) || (!DayTime && m.Flags["day_only"]) {
			continue
		}
		ids = append(ids, id)
		total += weight
	}
	if total == 0 {
		return 0, false
	}
	sort.Ints(ids)
	pick := utils.Roll(total, 1, 0)
	for _, id := range ids {
		pick -= weights[id]
		if pick <= 0 {
			return id, true
		}
	}
	return ids[len(ids)-1], true
}
