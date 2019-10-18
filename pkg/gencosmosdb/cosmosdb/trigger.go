package cosmosdb

import (
	"net/http"
)

// Trigger represents a trigger
type Trigger struct {
	ID               string           `json:"id,omitempty"`
	ResourceID       string           `json:"_rid,omitempty"`
	Timestamp        int              `json:"_ts,omitempty"`
	Self             string           `json:"_self,omitempty"`
	ETag             string           `json:"_etag,omitempty"`
	Body             string           `json:"body,omitempty"`
	TriggerOperation TriggerOperation `json:"triggerOperation,omitempty"`
	TriggerType      TriggerType      `json:"triggerType,omitempty"`
}

// TriggerOperation represents a trigger operation
type TriggerOperation string

// TriggerOperation constants
const (
	TriggerOperationAll     TriggerOperation = "All"
	TriggerOperationCreate  TriggerOperation = "Create"
	TriggerOperationReplace TriggerOperation = "Replace"
	TriggerOperationDelete  TriggerOperation = "Delete"
)

// TriggerType represents a trigger type
type TriggerType string

// TriggerType constants
const (
	TriggerTypePre  TriggerType = "Pre"
	TriggerTypePost TriggerType = "Post"
)

// Triggers represents triggers
type Triggers struct {
	Count      int        `json:"_count,omitempty"`
	ResourceID string     `json:"_rid,omitempty"`
	Triggers   []*Trigger `json:"Triggers,omitempty"`
}

type triggerClient struct {
	*databaseClient
	path string
}

// TriggerClient is a trigger client
type TriggerClient interface {
	Create(*Trigger) (*Trigger, error)
	List() TriggerIterator
	ListAll() (*Triggers, error)
	Get(string) (*Trigger, error)
	Delete(*Trigger) error
	Replace(*Trigger) (*Trigger, error)
}

type triggerListIterator struct {
	*triggerClient
	continuation string
	done         bool
}

// TriggerIterator is a trigger iterator
type TriggerIterator interface {
	Next() (*Triggers, error)
}

// NewTriggerClient returns a new trigger client
func NewTriggerClient(collc CollectionClient, collid string) TriggerClient {
	return &triggerClient{
		databaseClient: collc.(*collectionClient).databaseClient,
		path:           collc.(*collectionClient).path + "/colls/" + collid,
	}
}

func (c *triggerClient) all(i TriggerIterator) (*Triggers, error) {
	alltriggers := &Triggers{}

	for {
		triggers, err := i.Next()
		if err != nil {
			return nil, err
		}
		if triggers == nil {
			break
		}

		alltriggers.Count += triggers.Count
		alltriggers.ResourceID = triggers.ResourceID
		alltriggers.Triggers = append(alltriggers.Triggers, triggers.Triggers...)
	}

	return alltriggers, nil
}

func (c *triggerClient) Create(newtrigger *Trigger) (trigger *Trigger, err error) {
	err = c.do(http.MethodPost, c.path+"/triggers", "triggers", c.path, http.StatusCreated, &newtrigger, &trigger, nil)
	return
}

func (c *triggerClient) List() TriggerIterator {
	return &triggerListIterator{triggerClient: c}
}

func (c *triggerClient) ListAll() (*Triggers, error) {
	return c.all(c.List())
}

func (c *triggerClient) Get(triggerid string) (trigger *Trigger, err error) {
	err = c.do(http.MethodGet, c.path+"/triggers/"+triggerid, "triggers", c.path+"/triggers/"+triggerid, http.StatusOK, nil, &trigger, nil)
	return
}

func (c *triggerClient) Delete(trigger *Trigger) error {
	if trigger.ETag == "" {
		return ErrETagRequired
	}
	headers := http.Header{}
	headers.Set("If-Match", trigger.ETag)
	return c.do(http.MethodDelete, c.path+"/triggers/"+trigger.ID, "triggers", c.path+"/triggers/"+trigger.ID, http.StatusNoContent, nil, nil, headers)
}

func (c *triggerClient) Replace(newtrigger *Trigger) (trigger *Trigger, err error) {
	err = c.do(http.MethodPost, c.path+"/triggers/"+newtrigger.ID, "triggers", c.path+"/triggers/"+newtrigger.ID, http.StatusCreated, &newtrigger, &trigger, nil)
	return
}

func (i *triggerListIterator) Next() (triggers *Triggers, err error) {
	if i.done {
		return
	}

	headers := http.Header{}
	if i.continuation != "" {
		headers.Set("X-Ms-Continuation", i.continuation)
	}

	err = i.do(http.MethodGet, i.path+"/triggers", "triggers", i.path, http.StatusOK, nil, &triggers, headers)
	if err != nil {
		return
	}

	i.continuation = headers.Get("X-Ms-Continuation")
	i.done = i.continuation == ""

	return
}
