package actioncable

type Subscription struct {
	identifier string
	toClient   chan<- serverMessage
}

func (s Subscription) Identifier() string {
	return s.identifier
}

func (s Subscription) Receive(message string) (ok bool) {
	select {
	default: // the receiver stopped listening
		return false
	case s.toClient <- newData(s.identifier, message):
		return true
	}
}

func (s Subscription) Close() {
	s.toClient <- newSubscriptionRejection(s.identifier)
	// We leave the toClient channel open because other goroutines and Subscriptions should still be
	// able to send into it
}
