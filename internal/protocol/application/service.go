package application

import (
	"crypto/ed25519"
	"fmt"
	aggregate "github.com/example/federated-analytics-coordinator/internal/aggregation/domain"
	audit "github.com/example/federated-analytics-coordinator/internal/audit/domain"
	federation "github.com/example/federated-analytics-coordinator/internal/federation/domain"
	protocol "github.com/example/federated-analytics-coordinator/internal/protocol/domain"
	"sort"
	"sync"
	"time"
)

type Repository interface {
	CreateParticipant(federation.Participant) error
	FindParticipant(string, string) (federation.Participant, error)
	CreateTask(protocol.Task) error
	FindTask(string, string) (protocol.Task, error)
	SaveTask(protocol.Task) error
}
type Result struct {
	TaskID     string          `json:"task_id"`
	Metric     protocol.Metric `json:"metric"`
	Value      float64         `json:"value"`
	Count      int64           `json:"count"`
	Epsilon    float64         `json:"epsilon"`
	Precision  string          `json:"precision"`
	ReleasedAt time.Time       `json:"released_at"`
}
type Service struct {
	repo    Repository
	mu      sync.Mutex
	shares  map[string][]aggregate.Share
	nonces  map[string]bool
	results map[string]Result
	audits  map[string][]audit.Entry
	ids     func() string
	now     func() time.Time
}

func New(repo Repository, ids func() string) *Service {
	return &Service{repo: repo, shares: map[string][]aggregate.Share{}, nonces: map[string]bool{}, results: map[string]Result{}, audits: map[string][]audit.Entry{}, ids: ids, now: func() time.Time { return time.Now().UTC() }}
}
func (s *Service) RegisterParticipant(tenant string, p federation.Participant) (federation.Participant, error) {
	p.ID = s.ids()
	p.OrganizationID = tenant
	p.CreatedAt = s.now()
	p.UpdatedAt = p.CreatedAt
	p.Revision = 1
	if p.Status == "" {
		p.Status = federation.ParticipantActive
	}
	p.Capabilities = federation.NormalizeCapabilities(p.Capabilities)
	if e := p.Validate(); e != nil {
		return federation.Participant{}, e
	}
	if e := s.repo.CreateParticipant(p); e != nil {
		return federation.Participant{}, e
	}
	s.audit(tenant, "", "participant_registered", p.ID, "participant registration")
	return p, nil
}

type CreateTask struct {
	TemplateID          string          `json:"template_id"`
	Metric              protocol.Metric `json:"metric"`
	Participants        []string        `json:"participants"`
	MinimumParticipants int             `json:"minimum_participants"`
	Deadline            time.Time       `json:"deadline"`
	Epsilon             float64         `json:"epsilon"`
	Delta               float64         `json:"delta"`
}

func (s *Service) CreateTask(tenant string, c CreateTask) (protocol.Task, error) {
	if c.MinimumParticipants == 0 {
		c.MinimumParticipants = 2
	}
	for _, id := range c.Participants {
		p, e := s.repo.FindParticipant(tenant, id)
		if e != nil {
			return protocol.Task{}, e
		}
		if p.Status != federation.ParticipantActive {
			return protocol.Task{}, fmt.Errorf("participant %s not active", id)
		}
		if !p.Supports(string(c.Metric)) {
			return protocol.Task{}, fmt.Errorf("participant %s lacks metric capability", id)
		}
	}
	sort.Strings(c.Participants)
	t := protocol.Task{ID: s.ids(), TenantID: tenant, TemplateID: c.TemplateID, Metric: c.Metric, State: protocol.Draft, Participants: c.Participants, MinimumParticipants: c.MinimumParticipants, Deadline: c.Deadline, Budget: protocol.Budget{Epsilon: c.Epsilon, Delta: c.Delta}, Revision: 1, CreatedAt: s.now(), UpdatedAt: s.now()}
	if e := t.Validate(); e != nil {
		return protocol.Task{}, e
	}
	if e := s.repo.CreateTask(t); e != nil {
		return protocol.Task{}, e
	}
	s.audit(tenant, t.ID, "task_created", "system", "draft task")
	return t, nil
}
func (s *Service) Start(tenant, id string) (protocol.Task, error) {
	t, e := s.repo.FindTask(tenant, id)
	if e != nil {
		return t, e
	}
	if s.now().After(t.Deadline) {
		return t, fmt.Errorf("task deadline passed")
	}
	if e = t.Budget.Reserve(0.1); e != nil {
		return t, e
	}
	if e = t.Transition(protocol.Collecting, s.now()); e != nil {
		return t, e
	}
	if e = s.repo.SaveTask(t); e != nil {
		return t, e
	}
	s.audit(tenant, id, "task_started", "system", "collection started")
	return t, nil
}
func (s *Service) UploadShare(tenant string, share aggregate.Share) (aggregate.Share, error) {
	t, e := s.repo.FindTask(tenant, share.TaskID)
	if e != nil {
		return aggregate.Share{}, e
	}
	if t.State != protocol.Collecting && t.State != protocol.Masked {
		return aggregate.Share{}, fmt.Errorf("task is not accepting shares")
	}
	if s.now().After(t.Deadline) {
		return aggregate.Share{}, fmt.Errorf("task deadline passed")
	}
	included := false
	for _, id := range t.Participants {
		if id == share.ParticipantID {
			included = true
		}
	}
	if !included {
		return aggregate.Share{}, fmt.Errorf("participant is not assigned to task")
	}
	p, e := s.repo.FindParticipant(tenant, share.ParticipantID)
	if e != nil {
		return aggregate.Share{}, e
	}
	if p.Status != federation.ParticipantActive {
		return aggregate.Share{}, fmt.Errorf("participant is not active")
	}
	if e = share.Validate(); e != nil {
		return aggregate.Share{}, e
	}
	if !ed25519.Verify(ed25519.PublicKey(p.PublicKey), share.Canonical(), share.Signature) {
		return aggregate.Share{}, fmt.Errorf("invalid share signature")
	}
	s.mu.Lock()
	nonceKey := share.TaskID + "/" + share.ParticipantID + "/" + share.Nonce
	if s.nonces[nonceKey] {
		s.mu.Unlock()
		return aggregate.Share{}, fmt.Errorf("replayed share nonce")
	}
	for _, v := range s.shares[share.TaskID] {
		if v.ParticipantID == share.ParticipantID && v.Round == share.Round {
			s.mu.Unlock()
			return aggregate.Share{}, fmt.Errorf("duplicate round share")
		}
	}
	share.ID = s.ids()
	share.ReceivedAt = s.now()
	s.nonces[nonceKey] = true
	s.shares[share.TaskID] = append(s.shares[share.TaskID], share)
	s.mu.Unlock()
	s.audit(tenant, t.ID, "share_uploaded", share.ParticipantID, "signed share accepted")
	return share, nil
}
func (s *Service) Aggregate(tenant, id string) (Result, error) {
	t, e := s.repo.FindTask(tenant, id)
	if e != nil {
		return Result{}, e
	}
	if t.State != protocol.Collecting && t.State != protocol.Masked {
		return Result{}, fmt.Errorf("task cannot aggregate in state %s", t.State)
	}
	s.mu.Lock()
	shares := append([]aggregate.Share(nil), s.shares[id]...)
	s.mu.Unlock()
	if len(shares) < t.MinimumParticipants {
		return Result{}, fmt.Errorf("minimum participant threshold not met")
	}
	if t.State == protocol.Collecting {
		if e = t.Transition(protocol.Masked, s.now()); e != nil {
			return Result{}, e
		}
	}
	if e = t.Transition(protocol.Aggregating, s.now()); e != nil {
		return Result{}, e
	}
	a, e := aggregate.Sum(id, shares, s.now())
	if e != nil {
		return Result{}, e
	}
	if e = t.Transition(protocol.Revealing, s.now()); e != nil {
		return Result{}, e
	}
	r := Result{TaskID: id, Metric: t.Metric, Count: a.Count, Epsilon: 0.1, Precision: a.Precision, ReleasedAt: s.now()}
	switch t.Metric {
	case protocol.Mean:
		if a.Count == 0 {
			return Result{}, fmt.Errorf("mean has zero count")
		}
		r.Value = float64(a.Value) / float64(a.Count)
	default:
		r.Value = float64(a.Value)
	}
	if e = t.Budget.Commit(r.Epsilon); e != nil {
		return Result{}, e
	}
	if e = t.Transition(protocol.Completed, s.now()); e != nil {
		return Result{}, e
	}
	if e = s.repo.SaveTask(t); e != nil {
		return Result{}, e
	}
	s.mu.Lock()
	s.results[id] = r
	s.mu.Unlock()
	s.audit(tenant, id, "result_released", "system", "k-threshold and budget gate passed")
	return r, nil
}
func (s *Service) Abort(tenant, id, reason string) (protocol.Task, error) {
	t, e := s.repo.FindTask(tenant, id)
	if e != nil {
		return t, e
	}
	if e = t.Transition(protocol.Aborted, s.now()); e != nil {
		return t, e
	}
	t.Budget.Release(t.Budget.Reserved)
	t.AbortReason = reason
	if e = s.repo.SaveTask(t); e != nil {
		return t, e
	}
	s.audit(tenant, id, "task_aborted", "operator", reason)
	return t, nil
}
func (s *Service) GetTask(tenant, id string) (protocol.Task, error) {
	return s.repo.FindTask(tenant, id)
}
func (s *Service) Result(tenant, id string) (Result, error) {
	t, e := s.repo.FindTask(tenant, id)
	if e != nil {
		return Result{}, e
	}
	if t.State != protocol.Completed {
		return Result{}, fmt.Errorf("result not releasable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.results[id]
	if !ok {
		return Result{}, fmt.Errorf("result missing")
	}
	return r, nil
}
func (s *Service) Audit(tenant string) []audit.Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]audit.Entry(nil), s.audits[tenant]...)
}
func (s *Service) audit(tenant, task, action, actor, detail string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries := s.audits[tenant]
	previous := ""
	if len(entries) > 0 {
		previous = entries[len(entries)-1].Hash
	}
	e := audit.Entry{ID: s.ids(), TenantID: tenant, TaskID: task, Action: action, Actor: actor, Detail: detail, PreviousHash: previous, At: s.now()}
	e.Seal()
	s.audits[tenant] = append(entries, e)
}

func (s *Service) CompleteTask(tenant, taskID string) error {
	t, err := s.repo.FindTask(tenant, taskID)
	if err != nil {
		return err
	}
	if !s.CanComplete(t) {
		return fmt.Errorf("%s", s.CompletionReason(t))
	}
	if err = t.Transition(protocol.Completed, s.now()); err != nil {
		return err
	}
	if err = s.repo.SaveTask(t); err != nil {
		return err
	}
	return nil
}
