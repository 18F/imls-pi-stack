package state

import "errors"

type QueueRow struct {
	Rowid int
	Item  string
}
type Queue struct {
	name string
	fifo []string
}

func NewQueue(name string) (q *Queue) {
	q = &Queue{name: name, fifo: make([]string, 0)}
	return q
}

func (queue *Queue) Enqueue(item string) {
	queue.fifo = append(queue.fifo, item)
}

func (queue *Queue) Peek() (string, error) {
	if len(queue.fifo) > 0 {
		return queue.fifo[0], nil
	} else {
		return "", errors.New("queue is empty in peek")
	}

}

func (queue *Queue) Dequeue() (string, error) {
	if len(queue.fifo) > 0 {
		var s string
		s, queue.fifo = queue.fifo[0], queue.fifo[1:]
		return s, nil
	} else {
		return "", errors.New("queue is empty in dequeue")
	}

}
