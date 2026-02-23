package types

import "github.com/Egot3/Zhao/queues"

type Queue struct {
	Name           string                 `toml:"name" json:"name"`
	Enabled        bool                   `toml:"enabled" json:"-"`
	Durable        bool                   `toml:"durable" json:"durable"`
	DeleteOnUnused bool                   `toml:"delete-when-unused" json:"auto_delete"`
	Exclusive      bool                   `toml:"exclusive" json:"exclusive"`
	NoWait         bool                   `toml:"no-wait" json:"-"`
	Args           map[string]interface{} `toml:"arguments" json:"-"`
}

func Equal(q1, q2 Queue) bool {
	return q1.Name == q2.Name &&
		q1.Durable == q2.Durable &&
		q1.DeleteOnUnused == q2.DeleteOnUnused &&
		q1.Exclusive == q2.Exclusive
}

func (q Queue) Canonical() queues.QueueStruct {
	return queues.QueueStruct{
		Name:           q.Name,
		Durable:        q.Durable,
		DeleteOnUnused: q.DeleteOnUnused,
		Exclusive:      q.Exclusive,
		NoWait:         q.NoWait,
		Args:           q.Args,
	}
}
