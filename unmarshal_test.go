package darknut_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/ttab/darknut"
	"github.com/ttab/newsdoc"
)

type planningItem struct {
	UUID                uuid.UUID         `newsdoc:"uuid"`
	Title               string            `newsdoc:"title"`
	Meta                planningItemBlock `newsdoc:"meta,type=core/planning-item"`
	InternalDescription *descriptionBlock `newsdoc:"meta,type=core/description,role=internal"`
	PublicDescription   *descriptionBlock `newsdoc:"meta,type=core/description,role=public"`
	Assignments         []assignmentBlock `newsdoc:"meta,type=core/assignment"`
	Deliverables        []deliverableLink `newsdoc:"links,rel=deliverable"`
}

type descriptionBlock struct {
	Role string `newsdoc:"role"`
	Text string `newsdoc:"data.text"`
}

type planningItemBlock struct {
	Date        *time.Time `newsdoc:"data.date,format=2006-01-02"`
	Publish     time.Time  `newsdoc:"data.publish"`
	PublishSlot *int       `newsdoc:"data.publish_slot"`
	Public      bool       `newsdoc:"data.public"`
	Tentative   bool       `newsdoc:"data.tentative"`
	Urgency     int        `newsdoc:"data.urgency"`
}

type assignmentBlock struct {
	Starts    time.Time        `newsdoc:"data.starts"`
	Ends      *time.Time       `newsdoc:"data.ends"`
	Status    string           `newsdoc:"data.status"`
	FullDay   bool             `newsdoc:"data.full_day"`
	Kind      []assignmentKind `newsdoc:"meta,type=core/assignment-kind"`
	Assignees []assigneeLink   `newsdoc:"links,rel=assignee"`
}

type assignmentKind struct {
	Value string `newsdoc:"value"`
}

type assigneeLink struct {
	UUID uuid.UUID `newsdoc:"uuid"`
}

type deliverableLink struct {
	UUID uuid.UUID `newsdoc:"uuid"`
	Type string    `newsdoc:"type"`
}

func TestUnmarshalDocument(t *testing.T) {
	docData, err := os.ReadFile("testdata/planning.json")
	must(t, err, "read planning data")

	goldenData, err := os.ReadFile("testdata/planning.golden.json")
	must(t, err, "read planning golden data")

	var (
		doc    newsdoc.Document
		item   planningItem
		golden planningItem
	)

	err = json.Unmarshal(docData, &doc)
	must(t, err, "unmarshal NewsDoc data")

	err = json.Unmarshal(goldenData, &golden)
	must(t, err, "unmarshal NewsDoc data")

	err = darknut.UnmarshalDocument(doc, &item)
	must(t, err, "unmarshal newsdoc document")

	if diff := cmp.Diff(golden, item); diff != "" {
		t.Errorf("UnmarshalDocument() mismatch (-want +got):\n%s", diff)
	}
}

func must(t *testing.T, err error, msg string) {
	t.Helper()

	if err == nil {
		return
	}

	t.Fatalf("failed to %s: %v", msg, err)
}
