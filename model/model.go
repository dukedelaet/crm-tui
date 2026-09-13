package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dukedelaet/crm-tui/util"
)

// ─── Colors (pink & lavender) ────────────────────────────────────────────────

var (
	pink        = lipgloss.AdaptiveColor{Light: "#FF6B9D", Dark: "#FF6B9D"}
	lavender    = lipgloss.AdaptiveColor{Light: "#C8A2C8", Dark: "#B388EB"}
	bgDark      = lipgloss.AdaptiveColor{Light: "#FFF0F8", Dark: "#1A0A2E"}
	textPrimary = lipgloss.AdaptiveColor{Light: "#2D1B4E", Dark: "#F0E0FF"}
	textMuted   = lipgloss.AdaptiveColor{Light: "#8B6F9E", Dark: "#9B7FDA"}
	borderC     = lipgloss.AdaptiveColor{Light: "#E0B0FF", Dark: "#4A2070"}
	accentRed   = lipgloss.AdaptiveColor{Light: "#E91E63", Dark: "#FF5252"}
)

// ─── Types ────────────────────────────────────────────────────────────────────

type Status struct {
	Contacts          int     `json:"contacts"`
	Organizations     int     `json:"organizations"`
	OpenDeals         int     `json:"open_deals"`
	DealValue         float64 `json:"deal_value"`
	OpenTasks         int     `json:"open_tasks"`
	OverdueTasks      int     `json:"overdue_tasks"`
	Interactions7Days int     `json:"interactions_7days"`
}

type Tab int

const (
	TabDashboard Tab = iota
	TabPeople
	TabOrgs
	TabDeals
	TabTasks
	TabInteractions
)

type ModalType int

const (
	ModalNone ModalType = iota
	ModalPersonForm
	ModalOrgForm
	ModalDealForm
	ModalTaskForm
	ModalInteractionForm
	ModalConfirmDelete
	ModalQuickLaunch
	ModalDetail
)

type Person struct {
	ID    int
	Name  string
	Email string
	Phone string
	Title string
	OrgID int
	Tags  string
}

type Org struct {
	ID       int
	Name     string
	Domain   string
	Industry string
}

type Deal struct {
	ID       int
	Title    string
	Value    float64
	PersonID int
	Stage    string
}

type Task struct {
	ID       int
	Title    string
	PersonID int
	Due      string
	Priority string
	Done     bool
}

type Interaction struct {
	ID       int
	Type     string
	PersonID int
	Subject  string
	Content  string
	Date     string
}

type QuickLaunchEntry struct {
	Label    string
	Shortcut string
	Command  string
	Category string
}

type modalField struct {
	label    string
	value    string
	input    textinput.Model
	required bool
}

type listItem struct {
	title string
	id    int
}

func (i listItem) FilterValue() string { return i.title }

type qItem struct {
	entry QuickLaunchEntry
}

func (i qItem) FilterValue() string { return i.entry.Label + " " + i.entry.Command }

// ─── Model ────────────────────────────────────────────────────────────────────

type Model struct {
	tab              Tab
	modal            ModalType
	width            int
	height           int
	loading          bool
	lastErr          string
	people           []Person
	orgs             []Org
	deals            []Deal
	tasks            []Task
	interactions     []Interaction
	dashStats        string
	personList       list.Model
	orgList          list.Model
	dealList         list.Model
	taskList         list.Model
	logList          list.Model
	peopleTable      table.Model
	orgsTable        table.Model
	dealsTable       table.Model
	tasksTable       table.Model
	logsTable        table.Model
	quickLaunch      list.Model
	quickLaunchItems []qItem
	quickLaunchQuery textinput.Model
	searchInput      textinput.Model
	searchFocused    bool
	modalFields      []modalField
	modalTitle       string
	help             help.Model
	delegate         list.ItemDelegate
}

// ─── Init ─────────────────────────────────────────────────────────────────────

func NewModel() Model {
	d := list.NewDefaultDelegate()
	m := Model{
		tab:           TabDashboard,
		modal:         ModalNone,
		searchFocused: false,
		help:          help.New(),
		loading:       true,
		delegate:      d,
	}
	m.searchInput = newTI("Search...")
	m.searchInput.Focus()
	m.searchInput.Prompt = "> "
	m.quickLaunchQuery = newTI("Filter commands...")
	m.quickLaunchQuery.Focus()
	m.populateQL()
	m.refreshAll()
	return m
}

func (m *Model) populateQL() {
	m.quickLaunchItems = []qItem{
		{QuickLaunchEntry{"Dashboard", "g d", "crm status", "View"}},
		{QuickLaunchEntry{"List People", "g p", "crm person list", "People"}},
		{QuickLaunchEntry{"Add Person", "n p", "crm person add", "People"}},
		{QuickLaunchEntry{"Search People", "/ p", "crm person list --search", "People"}},
		{QuickLaunchEntry{"List Orgs", "g o", "crm org list", "Orgs"}},
		{QuickLaunchEntry{"Add Org", "n o", "crm org add", "Orgs"}},
		{QuickLaunchEntry{"List Deals", "g de", "crm deal list", "Deals"}},
		{QuickLaunchEntry{"Add Deal", "n de", "crm deal add", "Deals"}},
		{QuickLaunchEntry{"Pipeline", "g pl", "crm deal pipeline", "Deals"}},
		{QuickLaunchEntry{"List Tasks", "g t", "crm task list", "Tasks"}},
		{QuickLaunchEntry{"Add Task", "n t", "crm task add", "Tasks"}},
		{QuickLaunchEntry{"Overdue Tasks", "g to", "crm task list --overdue", "Tasks"}},
		{QuickLaunchEntry{"Log Call", "n c", "crm log call", "Logs"}},
		{QuickLaunchEntry{"Log Email", "n e", "crm log email", "Logs"}},
		{QuickLaunchEntry{"Log Meeting", "n m", "crm log meeting", "Logs"}},
		{QuickLaunchEntry{"Log Note", "n n", "crm log note", "Logs"}},
		{QuickLaunchEntry{"Search All", "/", "crm search", "Search"}},
		{QuickLaunchEntry{"Context Briefing", "g c", "crm context", "Briefing"}},
	}
	items := make([]list.Item, len(m.quickLaunchItems))
	for i, it := range m.quickLaunchItems {
		items[i] = it
	}
	m.quickLaunch = list.New(items, m.delegate, m.contentWidth()+2, 18)
}

// ─── Refresh ──────────────────────────────────────────────────────────────────

func (m *Model) refreshAll() tea.Cmd {
	m.loading = true
	return tea.Batch(
		m.doRefresh(TabDashboard),
		m.doRefresh(TabPeople),
		m.doRefresh(TabOrgs),
		m.doRefresh(TabDeals),
		m.doRefresh(TabTasks),
		m.doRefresh(TabInteractions),
	)
}

func (m *Model) refreshTabCmd() tea.Cmd { return m.doRefresh(m.tab) }

func (m *Model) doRefresh(tab Tab) tea.Cmd {
	switch tab {
	case TabDashboard:
		return func() tea.Msg {
			out, _, err := util.RunCRM("status", "-f", "json")
			if err != nil {
				return DashboardMsg{Err: err.Error()}
			}
			return DashboardMsg{Content: out}
		}
	case TabPeople:
		return func() tea.Msg {
			out, err := util.RunCRMJSON("person", "list")
			if err != nil {
				return PeopleMsg{Err: err}
			}
			var p []Person
			if out == "" || out == "[]" {
				return PeopleMsg{People: []Person{}}
			}
			json.Unmarshal([]byte(out), &p)
			return PeopleMsg{People: p}
		}
	case TabOrgs:
		return func() tea.Msg {
			out, err := util.RunCRMJSON("org", "list")
			if err != nil {
				return OrgsMsg{Err: err}
			}
			var o []Org
			if out == "" || out == "[]" {
				return OrgsMsg{Orgs: []Org{}}
			}
			json.Unmarshal([]byte(out), &o)
			return OrgsMsg{Orgs: o}
		}
	case TabDeals:
		return func() tea.Msg {
			out, err := util.RunCRMJSON("deal", "list")
			if err != nil {
				return DealsMsg{Err: err}
			}
			var d []Deal
			if out == "" || out == "[]" {
				return DealsMsg{Deals: []Deal{}}
			}
			json.Unmarshal([]byte(out), &d)
			return DealsMsg{Deals: d}
		}
	case TabTasks:
		return func() tea.Msg {
			out, err := util.RunCRMJSON("task", "list")
			if err != nil {
				return TasksMsg{Err: err}
			}
			var t []Task
			if out == "" || out == "[]" {
				return TasksMsg{Tasks: []Task{}}
			}
			json.Unmarshal([]byte(out), &t)
			return TasksMsg{Tasks: t}
		}
	case TabInteractions:
		return func() tea.Msg {
			out, err := util.RunCRMJSON("log", "list", "--limit", "20")
			if err != nil {
				return InteractionsMsg{Err: err}
			}
			var l []Interaction
			if out == "" || out == "[]" {
				return InteractionsMsg{Interactions: []Interaction{}}
			}
			json.Unmarshal([]byte(out), &l)
			return InteractionsMsg{Interactions: l}
		}
	}
	return nil
}

// ─── Messages ─────────────────────────────────────────────────────────────────

type DashboardMsg struct{ Content, Err string }
type PeopleMsg struct{ People []Person; Err error }
type OrgsMsg struct{ Orgs []Org; Err error }
type DealsMsg struct{ Deals []Deal; Err error }
type TasksMsg struct{ Tasks []Task; Err error }
type InteractionsMsg struct{ Interactions []Interaction; Err error }

// ─── Update ───────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd { return m.refreshAll() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m.resize(), nil

	case DashboardMsg:
		m.dashStats = msg.Content
		m.loading = false
		if msg.Err != "" {
			m.lastErr = msg.Err
		}
		return m, nil

	case PeopleMsg:
		m.loading = false
		if msg.Err != nil {
			m.lastErr = msg.Err.Error()
			return m, nil
		}
		m.people = msg.People
		return m.rebuildPeople(), nil

	case OrgsMsg:
		m.loading = false
		if msg.Err != nil {
			m.lastErr = msg.Err.Error()
			return m, nil
		}
		m.orgs = msg.Orgs
		return m.rebuildOrgs(), nil

	case DealsMsg:
		m.loading = false
		if msg.Err != nil {
			m.lastErr = msg.Err.Error()
			return m, nil
		}
		m.deals = msg.Deals
		return m.rebuildDeals(), nil

	case TasksMsg:
		m.loading = false
		if msg.Err != nil {
			m.lastErr = msg.Err.Error()
			return m, nil
		}
		m.tasks = msg.Tasks
		return m.rebuildTasks(), nil

	case InteractionsMsg:
		m.loading = false
		if msg.Err != nil {
			m.lastErr = msg.Err.Error()
			return m, nil
		}
		m.interactions = msg.Interactions
		return m.rebuildInteractions(), nil

	case tea.KeyMsg:
		if m.modal == ModalQuickLaunch {
			return m.updateQL(msg)
		}
		if m.isModal() {
			return m.updateModal(msg)
		}
		if m.searchFocused {
			m.searchInput, _ = m.searchInput.Update(msg)
			if msg.String() == "esc" || msg.String() == "ctrl+c" {
				m.searchFocused = false
				m.searchInput.Blur()
			}
			return m, nil
		}
		return m.updateMain(msg)
	}
	return m, nil
}

func (m Model) resize() Model {
	w := m.contentWidth()
	m.peopleTable = table.New(
		table.WithColumns(m.peopleTable.Columns()),
		table.WithRows(m.peopleTable.Rows()),
		table.WithWidth(w),
		table.WithHeight(15),
		table.WithFocused(true),
	)
	initTableStyles(&m.peopleTable)
	m.orgsTable = table.New(
		table.WithColumns(m.orgsTable.Columns()),
		table.WithRows(m.orgsTable.Rows()),
		table.WithWidth(w),
		table.WithHeight(15),
		table.WithFocused(true),
	)
	initTableStyles(&m.orgsTable)
	m.dealsTable = table.New(
		table.WithColumns(m.dealsTable.Columns()),
		table.WithRows(m.dealsTable.Rows()),
		table.WithWidth(w),
		table.WithHeight(15),
		table.WithFocused(true),
	)
	initTableStyles(&m.dealsTable)
	m.tasksTable = table.New(
		table.WithColumns(m.tasksTable.Columns()),
		table.WithRows(m.tasksTable.Rows()),
		table.WithWidth(w),
		table.WithHeight(15),
		table.WithFocused(true),
	)
	initTableStyles(&m.tasksTable)
	m.logsTable = table.New(
		table.WithColumns(m.logsTable.Columns()),
		table.WithRows(m.logsTable.Rows()),
		table.WithWidth(w),
		table.WithHeight(15),
		table.WithFocused(true),
	)
	initTableStyles(&m.logsTable)
	m.quickLaunch = list.New(m.quickLaunch.Items(), m.delegate, m.contentWidth()+2, 18)
	return m
}

func (m Model) updateMain(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "r":
		return m, m.refreshAll()
	case "ctrl+p":
		m.modal = ModalQuickLaunch
		return m, nil
	case "1":
		m.tab, m.lastErr = TabPeople, ""
		return m, m.refreshTabCmd()
	case "2":
		m.tab, m.lastErr = TabOrgs, ""
		return m, m.refreshTabCmd()
	case "3":
		m.tab, m.lastErr = TabDeals, ""
		return m, m.refreshTabCmd()
	case "4":
		m.tab, m.lastErr = TabTasks, ""
		return m, m.refreshTabCmd()
	case "5":
		m.tab, m.lastErr = TabInteractions, ""
		return m, m.refreshTabCmd()
	case "6":
		m.tab, m.lastErr = TabDashboard, ""
		return m, m.refreshTabCmd()
	case "l", "tab":
		m.tab = (m.tab + 1) % 6
		m.lastErr = ""
		return m, m.refreshTabCmd()
	case "h", "shift+tab":
		m.tab = (m.tab + 5) % 6
		m.lastErr = ""
		return m, m.refreshTabCmd()
	case "/":
		m.searchFocused = true
		m.searchInput.Focus()
		return m, nil
	case "n":
		m.openNewModal()
		return m, nil
	case "e":
		m.openEditModal()
		return m, nil
	case "d":
		m.modal = ModalConfirmDelete
		return m, nil
	case "v":
		m.modal = ModalDetail
		return m, nil
	}
	switch m.tab {
	case TabPeople:
		m.personList, _ = m.personList.Update(msg)
		m.peopleTable, _ = m.peopleTable.Update(msg)
	case TabOrgs:
		m.orgList, _ = m.orgList.Update(msg)
		m.orgsTable, _ = m.orgsTable.Update(msg)
	case TabDeals:
		m.dealList, _ = m.dealList.Update(msg)
		m.dealsTable, _ = m.dealsTable.Update(msg)
	case TabTasks:
		m.taskList, _ = m.taskList.Update(msg)
		m.tasksTable, _ = m.tasksTable.Update(msg)
	case TabInteractions:
		m.logList, _ = m.logList.Update(msg)
		m.logsTable, _ = m.logsTable.Update(msg)
	}
	return m, nil
}

func (m Model) updateQL(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.modal = ModalNone
		return m, nil
	case "enter":
		idx := m.quickLaunch.Index()
		if idx >= 0 && idx < len(m.quickLaunchItems) {
			m.modal = ModalNone
			m.openModalForCommand(m.quickLaunchItems[idx].entry)
		}
		return m, nil
	case "/":
		m.quickLaunchQuery.Focus()
		return m, nil
	}
	if m.quickLaunchQuery.Focused() {
		m.quickLaunchQuery, _ = m.quickLaunchQuery.Update(msg)
		return m, nil
	}
	m.quickLaunch, _ = m.quickLaunch.Update(msg)
	return m, nil
}

func (m Model) updateModal(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.modal = ModalNone
		m.modalFields = nil
		return m, nil
	case "tab", "ctrl+i":
		m.cycleField(1)
		return m, nil
	case "shift+tab":
		m.cycleField(-1)
		return m, nil
	case "enter", "ctrl+s":
		return m.submitModal()
	}
	for i := range m.modalFields {
		if m.modalFields[i].input.Focused() {
			newInput, cmd := m.modalFields[i].input.Update(msg)
			m.modalFields[i].input = newInput
			m.modalFields[i].value = newInput.Value()
			return m, cmd
		}
	}
	return m, nil
}

func (m *Model) cycleField(dir int) {
	for i, f := range m.modalFields {
		if f.input.Focused() {
			next := (i + dir + len(m.modalFields)) % len(m.modalFields)
			f.input.Blur()
			m.modalFields[i].input = f.input
			m.modalFields[next].input.Focus()
			m.modalFields[next].input.SetValue(m.modalFields[next].value)
			return
		}
	}
}

func (m *Model) isModal() bool {
	return m.modal != ModalNone && m.modal != ModalQuickLaunch
}

// ─── Modal openers ────────────────────────────────────────────────────────────

func (m *Model) openNewModal() {
	switch m.tab {
	case TabPeople:
		m.openPersonModal(nil)
	case TabOrgs:
		m.openOrgModal(nil)
	case TabDeals:
		m.openDealModal(nil)
	case TabTasks:
		m.openTaskModal(nil)
	case TabInteractions:
		m.openInteractionModal(nil)
	}
}

func (m *Model) openEditModal() {
	switch m.tab {
	case TabPeople:
		if idx := m.personList.Index(); idx >= 0 && idx < len(m.people) {
			m.openPersonModal(&m.people[idx])
		}
	case TabOrgs:
		if idx := m.orgList.Index(); idx >= 0 && idx < len(m.orgs) {
			m.openOrgModal(&m.orgs[idx])
		}
	case TabDeals:
		if idx := m.dealList.Index(); idx >= 0 && idx < len(m.deals) {
			m.openDealModal(&m.deals[idx])
		}
	case TabTasks:
		if idx := m.taskList.Index(); idx >= 0 && idx < len(m.tasks) {
			m.openTaskModal(&m.tasks[idx])
		}
	}
}

func (m *Model) openPersonModal(p *Person) {
	m.modal = ModalPersonForm
	if p != nil {
		m.modalTitle = "Edit Person"
		m.modalFields = []modalField{
			{"Name", p.Name, newTI(p.Name), true},
			{"Email", p.Email, newTI(p.Email), false},
			{"Phone", p.Phone, newTI(p.Phone), false},
			{"Title", p.Title, newTI(p.Title), false},
			{"Org ID", strconv.Itoa(p.OrgID), newTI(strconv.Itoa(p.OrgID)), false},
			{"Tags", p.Tags, newTI(p.Tags), false},
		}
	} else {
		m.modalTitle = "Add Person"
		m.modalFields = []modalField{
			{"Name", "", newTI(""), true},
			{"Email", "", newTI(""), false},
			{"Phone", "", newTI(""), false},
			{"Title", "", newTI(""), false},
			{"Org ID", "", newTI(""), false},
			{"Tags", "", newTI(""), false},
		}
	}
}

func (m *Model) openOrgModal(o *Org) {
	m.modal = ModalOrgForm
	if o != nil {
		m.modalTitle = "Edit Organization"
		m.modalFields = []modalField{
			{"Name", o.Name, newTI(o.Name), true},
			{"Domain", o.Domain, newTI(o.Domain), false},
			{"Industry", o.Industry, newTI(o.Industry), false},
		}
	} else {
		m.modalTitle = "Add Organization"
		m.modalFields = []modalField{
			{"Name", "", newTI(""), true},
			{"Domain", "", newTI(""), false},
			{"Industry", "", newTI(""), false},
		}
	}
}

func (m *Model) openDealModal(d *Deal) {
	m.modal = ModalDealForm
	if d != nil {
		m.modalTitle = "Edit Deal"
		m.modalFields = []modalField{
			{"Title", d.Title, newTI(d.Title), true},
			{"Value", fmt.Sprintf("%.0f", d.Value), newTI(fmt.Sprintf("%.0f", d.Value)), false},
			{"Person ID", strconv.Itoa(d.PersonID), newTI(strconv.Itoa(d.PersonID)), false},
			{"Stage", d.Stage, newTI(d.Stage), false},
		}
	} else {
		m.modalTitle = "Add Deal"
		m.modalFields = []modalField{
			{"Title", "", newTI(""), true},
			{"Value", "0", newTI("0"), false},
			{"Person ID", "", newTI(""), false},
			{"Stage", "proposal", newTI("proposal"), false},
		}
	}
}

func (m *Model) openTaskModal(t *Task) {
	m.modal = ModalTaskForm
	if t != nil {
		m.modalTitle = "Edit Task"
		m.modalFields = []modalField{
			{"Title", t.Title, newTI(t.Title), true},
			{"Person ID", strconv.Itoa(t.PersonID), newTI(strconv.Itoa(t.PersonID)), false},
			{"Due", t.Due, newTI(t.Due), false},
			{"Priority", t.Priority, newTI(t.Priority), false},
		}
	} else {
		m.modalTitle = "Add Task"
		m.modalFields = []modalField{
			{"Title", "", newTI(""), true},
			{"Person ID", "", newTI(""), false},
			{"Due", time.Now().AddDate(0, 0, 7).Format("2006-01-02"), newTI(time.Now().AddDate(0, 0, 7).Format("2006-01-02")), false},
			{"Priority", "medium", newTI("medium"), false},
		}
	}
}

func (m *Model) openInteractionModal(l *Interaction) {
	m.modal = ModalInteractionForm
	if l != nil {
		m.modalTitle = "Edit Interaction"
		m.modalFields = []modalField{
			{"Type", l.Type, newTI(l.Type), true},
			{"Person ID", strconv.Itoa(l.PersonID), newTI(strconv.Itoa(l.PersonID)), true},
			{"Subject", l.Subject, newTI(l.Subject), true},
			{"Content", l.Content, newTI(l.Content), false},
			{"Date", l.Date, newTI(l.Date), false},
		}
	} else {
		m.modalTitle = "Log Interaction"
		m.modalFields = []modalField{
			{"Type", "call", newTI("call"), true},
			{"Person ID", "", newTI(""), true},
			{"Subject", "", newTI(""), true},
			{"Content", "", newTI(""), false},
			{"Date", time.Now().Format("2006-01-02 15:04"), newTI(time.Now().Format("2006-01-02 15:04")), false},
		}
	}
}

func (m *Model) openModalForCommand(entry QuickLaunchEntry) {
	switch entry.Command {
	case "crm person add":
		m.openPersonModal(nil)
	case "crm org add":
		m.openOrgModal(nil)
	case "crm deal add":
		m.openDealModal(nil)
	case "crm task add":
		m.openTaskModal(nil)
	case "crm log call", "crm log email", "crm log meeting", "crm log note":
		m.openInteractionModal(nil)
	default:
		parts := strings.Fields(entry.Command)
		if len(parts) > 1 {
			args := parts[1:]
			m.runCRM(args)
		}
	}
}

func (m *Model) runCRM(args []string) tea.Cmd {
	return func() tea.Msg {
		_, _, err := util.RunCRM(args...)
		if err != nil {
			return DashboardMsg{Err: err.Error()}
		}
		return DashboardMsg{Content: "Command executed"}
	}
}

// ─── Modal submit ─────────────────────────────────────────────────────────────

func (m *Model) submitModal() (Model, tea.Cmd) {
	switch m.modal {
	case ModalPersonForm:
		return m.submitPerson()
	case ModalOrgForm:
		return m.submitOrg()
	case ModalDealForm:
		return m.submitDeal()
	case ModalTaskForm:
		return m.submitTask()
	case ModalInteractionForm:
		return m.submitInteraction()
	case ModalConfirmDelete:
		return m.doDelete()
	}
	m.modal = ModalNone
	return *m, nil
}

func (m *Model) submitPerson() (Model, tea.Cmd) {
	if m.modalFields[0].value == "" {
		m.lastErr = "Name is required"
		return *m, nil
	}
	args := []string{"person", "add", m.modalFields[0].value}
	for i, keys := range []string{"email", "phone", "title", "org", "tags"} {
		if m.modalFields[i+1].value != "" {
			args = append(args, "--"+keys, m.modalFields[i+1].value)
		}
	}
	m.modal = ModalNone
	return *m, func() tea.Msg {
		_, _, err := util.RunCRM(args...)
		if err != nil {
			return DashboardMsg{Err: err.Error()}
		}
		return DashboardMsg{Content: "Person added"}
	}
}

func (m *Model) submitOrg() (Model, tea.Cmd) {
	if m.modalFields[0].value == "" {
		m.lastErr = "Name is required"
		return *m, nil
	}
	args := []string{"org", "add", m.modalFields[0].value}
	for i, keys := range []string{"domain", "industry"} {
		if m.modalFields[i+1].value != "" {
			args = append(args, "--"+keys, m.modalFields[i+1].value)
		}
	}
	m.modal = ModalNone
	return *m, func() tea.Msg {
		_, _, err := util.RunCRM(args...)
		if err != nil {
			return DashboardMsg{Err: err.Error()}
		}
		return DashboardMsg{Content: "Organization added"}
	}
}

func (m *Model) submitDeal() (Model, tea.Cmd) {
	if m.modalFields[0].value == "" {
		m.lastErr = "Title is required"
		return *m, nil
	}
	args := []string{"deal", "add", m.modalFields[0].value}
	for i, keys := range []string{"value", "person", "stage"} {
		if m.modalFields[i+1].value != "" {
			args = append(args, "--"+keys, m.modalFields[i+1].value)
		}
	}
	m.modal = ModalNone
	return *m, func() tea.Msg {
		_, _, err := util.RunCRM(args...)
		if err != nil {
			return DashboardMsg{Err: err.Error()}
		}
		return DashboardMsg{Content: "Deal added"}
	}
}

func (m *Model) submitTask() (Model, tea.Cmd) {
	if m.modalFields[0].value == "" {
		m.lastErr = "Title is required"
		return *m, nil
	}
	args := []string{"task", "add", m.modalFields[0].value}
	for i, keys := range []string{"person", "due", "priority"} {
		if m.modalFields[i+1].value != "" {
			args = append(args, "--"+keys, m.modalFields[i+1].value)
		}
	}
	m.modal = ModalNone
	return *m, func() tea.Msg {
		_, _, err := util.RunCRM(args...)
		if err != nil {
			return DashboardMsg{Err: err.Error()}
		}
		return DashboardMsg{Content: "Task added"}
	}
}

func (m *Model) submitInteraction() (Model, tea.Cmd) {
	ltype := m.modalFields[0].value
	if m.modalFields[1].value == "" {
		m.lastErr = "Person ID is required"
		return *m, nil
	}
	args := []string{"log", ltype, m.modalFields[1].value}
	for i, keys := range []string{"subject", "content", "at"} {
		if m.modalFields[i+2].value != "" {
			args = append(args, "--"+keys, m.modalFields[i+2].value)
		}
	}
	m.modal = ModalNone
	return *m, func() tea.Msg {
		_, _, err := util.RunCRM(args...)
		if err != nil {
			return DashboardMsg{Err: err.Error()}
		}
		return DashboardMsg{Content: "Interaction logged"}
	}
}

func (m *Model) doDelete() (Model, tea.Cmd) {
	var args []string
	switch m.tab {
	case TabPeople:
		if idx := m.personList.Index(); idx >= 0 && idx < len(m.people) {
			args = []string{"person", "delete", strconv.Itoa(m.people[idx].ID)}
		}
	case TabOrgs:
		if idx := m.orgList.Index(); idx >= 0 && idx < len(m.orgs) {
			args = []string{"org", "delete", strconv.Itoa(m.orgs[idx].ID)}
		}
	case TabDeals:
		if idx := m.dealList.Index(); idx >= 0 && idx < len(m.deals) {
			args = []string{"deal", "delete", strconv.Itoa(m.deals[idx].ID)}
		}
	case TabTasks:
		if idx := m.taskList.Index(); idx >= 0 && idx < len(m.tasks) {
			args = []string{"task", "delete", strconv.Itoa(m.tasks[idx].ID)}
		}
	}
	m.modal = ModalNone
	if len(args) == 0 {
		m.lastErr = "Nothing selected"
		return *m, nil
	}
	return *m, func() tea.Msg {
		_, _, err := util.RunCRM(args...)
		if err != nil {
			return DashboardMsg{Err: err.Error()}
		}
		return DashboardMsg{Content: "Deleted"}
	}
}

// ─── Rebuild from data ────────────────────────────────────────────────────────

func (m Model) rebuildPeople() Model {
	items := make([]list.Item, len(m.people))
	for i, p := range m.people {
		items[i] = listItem{title: p.Name + compactTag(p.Tags), id: p.ID}
	}
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = lipgloss.NewStyle().Background(lavender).Foreground(bgDark).Bold(true)
	m.delegate = d
	m.personList = list.New(items, m.delegate, m.contentWidth(), 15)
	m.peopleTable = newTable(
		[]table.Column{{Title: "ID", Width: 6}, {Title: "Name", Width: 20}, {Title: "Email", Width: 25}, {Title: "Phone", Width: 15}, {Title: "Title", Width: 15}, {Title: "Tags", Width: 15}},
		personRows(m.people),
	)
	return m
}

func (m Model) rebuildOrgs() Model {
	items := make([]list.Item, len(m.orgs))
	for i, o := range m.orgs {
		items[i] = listItem{title: o.Name, id: o.ID}
	}
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = lipgloss.NewStyle().Background(lavender).Foreground(bgDark).Bold(true)
	m.delegate = d
	m.orgList = list.New(items, m.delegate, m.contentWidth(), 15)
	m.orgsTable = newTable(
		[]table.Column{{Title: "ID", Width: 6}, {Title: "Name", Width: 25}, {Title: "Domain", Width: 25}, {Title: "Industry", Width: 20}},
		orgRows(m.orgs),
	)
	return m
}

func (m Model) rebuildDeals() Model {
	items := make([]list.Item, len(m.deals))
	for i, d := range m.deals {
		items[i] = listItem{title: fmt.Sprintf("%s  $%.0f [%s]", d.Title, d.Value, d.Stage), id: d.ID}
	}
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = lipgloss.NewStyle().Background(lavender).Foreground(bgDark).Bold(true)
	m.delegate = d
	m.dealList = list.New(items, m.delegate, m.contentWidth(), 15)
	m.dealsTable = newTable(
		[]table.Column{{Title: "ID", Width: 6}, {Title: "Title", Width: 25}, {Title: "Value", Width: 12}, {Title: "Person", Width: 10}, {Title: "Stage", Width: 12}},
		dealRows(m.deals),
	)
	return m
}

func (m Model) rebuildTasks() Model {
	items := make([]list.Item, len(m.tasks))
	for i, t := range m.tasks {
		done := "✗"
		if t.Done {
			done = "✓"
		}
		items[i] = listItem{title: fmt.Sprintf("%s [%s] due %s  %s", t.Title, t.Priority, t.Due, done), id: t.ID}
	}
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = lipgloss.NewStyle().Background(lavender).Foreground(bgDark).Bold(true)
	m.delegate = d
	m.taskList = list.New(items, m.delegate, m.contentWidth(), 15)
	m.tasksTable = newTable(
		[]table.Column{{Title: "ID", Width: 6}, {Title: "Title", Width: 30}, {Title: "Person", Width: 10}, {Title: "Due", Width: 12}, {Title: "Priority", Width: 10}, {Title: "Done", Width: 6}},
		taskRows(m.tasks),
	)
	return m
}

func (m Model) rebuildInteractions() Model {
	items := make([]list.Item, len(m.interactions))
	for i, l := range m.interactions {
		items[i] = listItem{title: fmt.Sprintf("[%s] %s — %s", strings.ToUpper(l.Type), l.Subject, l.Date), id: l.ID}
	}
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = lipgloss.NewStyle().Background(lavender).Foreground(bgDark).Bold(true)
	m.delegate = d
	m.logList = list.New(items, m.delegate, m.contentWidth(), 15)
	m.logsTable = newTable(
		[]table.Column{{Title: "ID", Width: 6}, {Title: "Type", Width: 8}, {Title: "Subject", Width: 30}, {Title: "Person", Width: 10}, {Title: "Date", Width: 18}},
		logRows(m.interactions),
	)
	return m
}

// ─── Constants ────────────────────────────────────────────────────────────────

const sidebarWidth = 16

func (m Model) contentWidth() int {
	return max(m.width-sidebarWidth-4, 10)
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	contentW := max(m.width-sidebarWidth-2, 10)

	sidebar := m.renderSidebar()
	content := m.renderContent(contentW)
	footer := m.footer()

	return lipgloss.JoinHorizontal(lipgloss.Left, sidebar, content) +
		lipgloss.NewStyle().Width(m.width).Render(footer) + "\n"
}

func (m Model) renderSidebar() string {
	title := lipgloss.NewStyle().
		Background(pink).Foreground(bgDark).Bold(true).Padding(0, 1).Width(sidebarWidth).Render(" crm-tui ")

	divider := lipgloss.NewStyle().Width(sidebarWidth).Foreground(borderC).Render("─")

	tabs := []struct {
		tab  Tab
		name string
	}{
		{TabDashboard, "Status"},
		{TabPeople, "People"},
		{TabOrgs, "Orgs"},
		{TabDeals, "Deals"},
		{TabTasks, "Tasks"},
		{TabInteractions, "Logs"},
	}

	var lines []string
	lines = append(lines, title, divider)
	for _, t := range tabs {
		active := m.tab == t.tab
		s := lipgloss.NewStyle().Width(sidebarWidth).Padding(0, 1)
		if active {
			s = s.Background(lavender).Foreground(bgDark).Bold(true)
		} else {
			s = s.Foreground(textMuted)
		}
		prefix := " "
		if active {
			prefix = "▸ "
		}
		lines = append(lines, s.Render(prefix+t.name))
	}

	return lipgloss.NewStyle().
		Background(bgDark).
		Width(sidebarWidth).
		Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderContent(w int) string {
	if m.lastErr != "" {
		m.lastErr = "" // handled inside content renderers now
	}

	var main string
	switch m.modal {
	case ModalQuickLaunch:
		main = m.renderQL(w)
	case ModalPersonForm, ModalOrgForm, ModalDealForm, ModalTaskForm, ModalInteractionForm, ModalConfirmDelete:
		main = m.renderModal(w)
	case ModalDetail:
		main = m.renderDetail(w)
	default:
		switch m.tab {
		case TabDashboard:
			main = m.renderDashboard(w)
		case TabPeople:
			main = m.renderTable("people", w, m.peopleTable.View(), len(m.people) == 0)
		case TabOrgs:
			main = m.renderTable("orgs", w, m.orgsTable.View(), len(m.orgs) == 0)
		case TabDeals:
			main = m.renderTable("deals", w, m.dealsTable.View(), len(m.deals) == 0)
		case TabTasks:
			main = m.renderTable("tasks", w, m.tasksTable.View(), len(m.tasks) == 0)
		case TabInteractions:
			main = m.renderTable("logs", w, m.logsTable.View(), len(m.interactions) == 0)
		default:
			main = m.renderDashboard(w)
		}
	}

	if m.lastErr != "" {
		errBanner := lipgloss.NewStyle().Background(accentRed).Foreground(bgDark).Padding(0, 1).Width(w).Render(" !  " + m.lastErr)
		m.lastErr = ""
		return lipgloss.JoinVertical(lipgloss.Left, errBanner, main)
	}
	return main
}

func (m Model) renderTable(name string, w int, content string, empty bool) string {
	if m.loading {
		return stylePanel(w).Render(lipgloss.NewStyle().Foreground(textMuted).Italic(true).Render("  Loading "+name+"..."))
	}
	if empty {
		hint := "Press 'n' to add one."
		if name == "logs" {
			hint = "Press 'n' to log one."
		}
		return stylePanel(w).Render(lipgloss.NewStyle().Foreground(textMuted).Italic(true).Render("  No "+name+". "+hint))
	}
	return stylePanel(w).Render(content)
}

func (m Model) renderDashboard(w int) string {
	if m.loading && m.dashStats == "" {
		return stylePanel(w).Render(lipgloss.NewStyle().Foreground(textMuted).Italic(true).Render("  Loading dashboard..."))
	}
	var s Status
	if m.dashStats != "" {
		json.Unmarshal([]byte(m.dashStats), &s)
	}
	return stylePanel(w).Render(formatStatus(s, w))
}

func formatStatus(s Status, w int) string {
	title := lipgloss.NewStyle().
		Background(pink).Foreground(bgDark).Bold(true).Padding(0, 1).Render(" Dashboard ")

	contacts := statBox("Contacts", s.Contacts)
	orgs := statBox("Orgs", s.Organizations)
	deals := statBox("Deals", s.OpenDeals)
	value := statBox("$ Value", s.DealValue)
	tasks := statBox("Tasks", s.OpenTasks)
	overdue := statBox("Overdue", s.OverdueTasks)
	recent := statBox("This Week", s.Interactions7Days)

	row1 := lipgloss.JoinHorizontal(lipgloss.Left, contacts, orgs, deals, value)
	row2 := lipgloss.JoinHorizontal(lipgloss.Left, tasks, overdue, recent)

	divider := lipgloss.NewStyle().Foreground(borderC).Render(strings.Repeat("─", w-4))

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		lipgloss.NewStyle().Height(1).Render(""),
		divider,
		lipgloss.NewStyle().Height(1).Render(""),
		row1,
		lipgloss.NewStyle().Height(1).Render(""),
		row2,
	)
}

func statBox(label string, value any) string {
	l := lipgloss.NewStyle().
		Foreground(lavender).Bold(true).Width(14).Padding(0, 1).BorderLeft(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(borderC)
	v := lipgloss.NewStyle().
		Foreground(pink).Bold(true).Width(10).Padding(0, 1).BorderRight(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(borderC)
	return lipgloss.JoinHorizontal(lipgloss.Left, l.Render(label), v.Render(fmt.Sprintf("%v", value)))
}

func (m Model) renderQL(w int) string {
	title := lipgloss.NewStyle().Background(pink).Foreground(bgDark).Bold(true).Padding(0, 1).Render(" Command Palette  (ctrl+p)")
	search := m.quickLaunchQuery.View()
	lv := m.quickLaunch.View()
	hint := lipgloss.NewStyle().Foreground(textMuted).Render("  ↑↓ Navigate  Enter: Open  Esc: Close  /: Filter")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Background(bgDark).
		BorderForeground(borderC).
		Padding(1, 2).
		Width(min(w, 70)).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", search, "", lv, "", hint))
}

func (m Model) renderModal(w int) string {
	if len(m.modalFields) == 0 {
		return ""
	}
	var lines []string
	lines = append(lines,
		lipgloss.NewStyle().Background(pink).Foreground(bgDark).Bold(true).Padding(0, 1).Render(" "+m.modalTitle),
		"",
	)
	for _, f := range m.modalFields {
		var rendered string
		if f.input.Focused() {
			rendered = lipgloss.JoinHorizontal(lipgloss.Left,
				lipgloss.NewStyle().Foreground(lavender).Bold(true).Width(14).Padding(0, 1).Render(f.label+": "),
				f.input.View(),
			)
		} else {
			rendered = lipgloss.JoinHorizontal(lipgloss.Left,
				lipgloss.NewStyle().Foreground(lavender).Bold(true).Width(14).Padding(0, 1).Render(f.label+": "),
				lipgloss.NewStyle().Foreground(textPrimary).Background(bgDark).Padding(0, 1).Render(f.value),
			)
		}
		lines = append(lines, rendered)
	}
	lines = append(lines, "",
		lipgloss.NewStyle().Foreground(textMuted).Italic(true).Render("  Tab: switch fields · Enter/Ctrl+S: save · Esc: cancel"),
	)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Background(bgDark).
		BorderForeground(borderC).
		Padding(1, 2).
		Width(min(w, 60)).
		Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderDetail(w int) string {
	var content string
	switch m.tab {
	case TabPeople:
		if idx := m.personList.Index(); idx >= 0 && idx < len(m.people) {
			p := m.people[idx]
			content = fmt.Sprintf("  Name:     %s\n  Email:    %s\n  Phone:    %s\n  Title:    %s\n  Org ID:   %d\n  Tags:     %s", p.Name, p.Email, p.Phone, p.Title, p.OrgID, p.Tags)
		}
	case TabOrgs:
		if idx := m.orgList.Index(); idx >= 0 && idx < len(m.orgs) {
			o := m.orgs[idx]
			content = fmt.Sprintf("  Name:      %s\n  Domain:    %s\n  Industry:  %s", o.Name, o.Domain, o.Industry)
		}
	case TabDeals:
		if idx := m.dealList.Index(); idx >= 0 && idx < len(m.deals) {
			d := m.deals[idx]
			content = fmt.Sprintf("  Title:   %s\n  Value:   $%.0f\n  Person:  %d\n  Stage:   %s", d.Title, d.Value, d.PersonID, d.Stage)
		}
	case TabTasks:
		if idx := m.taskList.Index(); idx >= 0 && idx < len(m.tasks) {
			t := m.tasks[idx]
			done := "No"
			if t.Done {
				done = "Yes"
			}
			content = fmt.Sprintf("  Title:   %s\n  Person:  %d\n  Due:     %s\n  Priority:%s\n  Done:    %s", t.Title, t.PersonID, t.Due, t.Priority, done)
		}
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Background(bgDark).
		BorderForeground(borderC).
		Padding(1, 2).
		Width(min(w, 50)).
		Render(content)
}

func (m Model) footer() string {
	return lipgloss.NewStyle().
		Foreground(textMuted).
		Width(m.width).
		Render(lipgloss.JoinHorizontal(lipgloss.Center,
			lipgloss.NewStyle().Foreground(pink).Bold(true).Render("↑↓")+" Navigate  ",
			lipgloss.NewStyle().Foreground(pink).Bold(true).Render("H/L")+" Tabs  ",
			lipgloss.NewStyle().Foreground(pink).Bold(true).Render("N")+" New  ",
			lipgloss.NewStyle().Foreground(pink).Bold(true).Render("E")+" Edit  ",
			lipgloss.NewStyle().Foreground(pink).Bold(true).Render("D")+" Delete  ",
			lipgloss.NewStyle().Foreground(pink).Bold(true).Render("V")+" View  ",
			lipgloss.NewStyle().Foreground(pink).Bold(true).Render("/")+" Search  ",
			lipgloss.NewStyle().Foreground(pink).Bold(true).Render("R")+" Refresh  ",
			lipgloss.NewStyle().Foreground(pink).Bold(true).Render("Ctrl+P")+" Palette  ",
			lipgloss.NewStyle().Foreground(pink).Bold(true).Render("Q")+" Quit",
		))
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func stylePanel(w int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 1).
		BorderForeground(borderC).
		Width(w)
}

func tabStyle(selected bool) lipgloss.Style {
	s := lipgloss.NewStyle().Padding(0, 2).BorderBottom(true).BorderStyle(lipgloss.NormalBorder())
	if selected {
		s = s.Background(pink).Foreground(bgDark).Bold(true).BorderBackground(pink)
	} else {
		s = s.Foreground(textMuted).BorderForeground(borderC)
	}
	return s
}

func newTI(placeholder string) textinput.Model {
	i := textinput.New()
	i.Placeholder = placeholder
	i.Prompt = ""
	i.CharLimit = 100
	i.PlaceholderStyle = lipgloss.NewStyle().Foreground(textMuted)
	i.Cursor.Style = lipgloss.NewStyle().Background(pink).Foreground(bgDark)
	i.TextStyle = lipgloss.NewStyle().Foreground(textPrimary)
	return i
}

func compactTag(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func newTable(cols []table.Column, rows []table.Row) table.Model {
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(15),
	)
	t.SetStyles(table.Styles{
		Header:   lipgloss.NewStyle().Foreground(pink).Bold(true).Padding(0, 1),
		Cell:     lipgloss.NewStyle().Foreground(textPrimary).Padding(0, 1),
		Selected: lipgloss.NewStyle().Background(lavender).Foreground(bgDark).Bold(true),
	})
	return t
}

func initTableStyles(t *table.Model) {
	t.SetStyles(table.Styles{
		Header:   lipgloss.NewStyle().Foreground(pink).Bold(true).Padding(0, 1),
		Cell:     lipgloss.NewStyle().Foreground(textPrimary).Padding(0, 1),
		Selected: lipgloss.NewStyle().Background(lavender).Foreground(bgDark).Bold(true),
	})
}

func personRows(ps []Person) []table.Row {
	rows := []table.Row{{"ID", "Name", "Email", "Phone", "Title", "Tags"}}
	for _, p := range ps {
		rows = append(rows, table.Row{strconv.Itoa(p.ID), p.Name, p.Email, p.Phone, p.Title, compactTag(p.Tags)})
	}
	return rows
}

func orgRows(os []Org) []table.Row {
	rows := []table.Row{{"ID", "Name", "Domain", "Industry"}}
	for _, o := range os {
		rows = append(rows, table.Row{strconv.Itoa(o.ID), o.Name, o.Domain, o.Industry})
	}
	return rows
}

func dealRows(ds []Deal) []table.Row {
	rows := []table.Row{{"ID", "Title", "Value", "Person", "Stage"}}
	for _, d := range ds {
		rows = append(rows, table.Row{strconv.Itoa(d.ID), d.Title, fmt.Sprintf("$%.0f", d.Value), strconv.Itoa(d.PersonID), d.Stage})
	}
	return rows
}

func taskRows(ts []Task) []table.Row {
	rows := []table.Row{{"ID", "Title", "Person", "Due", "Priority", "Done"}}
	for _, t := range ts {
		done := "✗"
		if t.Done {
			done = "✓"
		}
		rows = append(rows, table.Row{strconv.Itoa(t.ID), t.Title, strconv.Itoa(t.PersonID), t.Due, t.Priority, done})
	}
	return rows
}

func logRows(is []Interaction) []table.Row {
	rows := []table.Row{{"ID", "Type", "Subject", "Person", "Date"}}
	for _, l := range is {
		rows = append(rows, table.Row{strconv.Itoa(l.ID), strings.ToUpper(l.Type), l.Subject, strconv.Itoa(l.PersonID), l.Date})
	}
	return rows
}
