# Darknut

[![GoDev](https://img.shields.io/static/v1?label=godev&message=reference&color=00add8)][godev]
[![Build Status](https://github.com/ttab/darknut/actions/workflows/test.yml/badge.svg?branch=master)][actions]

![Image](docs/darknut.png?raw=true)

Darknut can be used to unmarshal a NewsDoc into a specialised struct. Mostly useful for getting strict typing of data attributes and flattening the data structure.

Slices of blocks and fields that are pointers will be treated as optional, all others will result in an error. External types, like UUIDs, are supported through [TextUnmarshaler](https://pkg.go.dev/encoding#TextUnmarshaler).

``` go
type planningItem struct {
	UUID                uuid.UUID         `newsdoc:"uuid"`
	Title               string            `newsdoc:"title"`
	Meta                planningItemBlock `newsdoc:"meta,type=core/planning-item"`
    InternalDescription *descriptionBlock `newsdoc:"meta,type=core/description,role=internal"`
	PublicDescription   *descriptionBlock `newsdoc:"meta,type=core/description,role=public"`
	Assignments         []assignmentBlock `newsdoc:"meta,type=core/assignment"`
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
```

See the [documentation][godev] for more information.

[godev]: https://pkg.go.dev/github.com/ttab/darknut
[actions]: https://github.com/ttab/darknut/actions
