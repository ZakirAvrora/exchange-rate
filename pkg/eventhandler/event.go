package eventhandler

type Event interface {
	Handle(*Processor) error
}

type Processor struct {
	ch          chan Event
	closeNotify chan struct{}
}

func (p *Processor) Run() {
	for {
		select {
		case event := <-p.ch:
			event.Handle(p)
		case <-p.closeNotify:
			return
		}
	}
}

func (p *Processor) QueueEvent(evt Event) {
	select {
	case p.ch <- evt:
	case <-p.closeNotify:
		return
	}
}
