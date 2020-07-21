package queue

import (
	"container/list"
)

type Queue struct {
	l *list.List
}

func (q *Queue) IsEmpty() bool {
	return q.l.Len() == 0
}

func (q *Queue) Dequeue() interface{} {
	e := q.l.Front()
	if e != nil {
		return q.l.Remove(e)
	}
	return nil
}

func (q *Queue) Enqueue(e interface{}) {
	q.l.PushFront(e)

}

func New() *Queue {
	return &Queue{l: list.New()}
}
