package mediator

type Updatable interface {
	Update(key string, msg map[string]any)
}

type Coordinator struct {
	updatable Updatable
}

func NewCoordinator(up Updatable) *Coordinator {
	return &Coordinator{updatable: up}
}

func (c *Coordinator) Notify(key string, msg map[string]any) {
	c.updatable.Update(key, msg)
}
