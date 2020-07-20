package queue

import (
	"container/list"
)

type Queue struct {
	l *list.List
}

func (q *Queue) Append(l *list.List) {
	q.l.PushFrontList(l)
}

func (q *Queue) Len() int {
	return q.l.Len()
}

func (q *Queue) Pop() interface{} {

	e := q.l.Front()
	if e != nil {
		return q.l.Remove(e)
	}
	return nil
}

func New() *Queue {
	return &Queue{l: list.New()}
}
