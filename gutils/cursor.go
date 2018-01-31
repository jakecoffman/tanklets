package gutils

type Cursor struct {
	next, start, max int
}

func NewCursor(start, max int) *Cursor {
	return &Cursor{next: start, start: start, max: max}
}

func (c *Cursor) Next() int {
	r := c.next
	c.next++
	if c.next > c.max {
		c.next = c.start
	}
	return r
}
