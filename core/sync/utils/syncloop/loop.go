package syncloop

import "time"

func (s *Loop) syncLoop() {
	for {
		select {
		case value := <-s.notice:
			switch value {
			case 0:
			case 1:
			}
		}
	}
}

func (s *Loop) doubleTimer() {
	t := time.Duration(s.userData.Interval)
	for {
		s.notice <- 0
		time.Sleep(time.Second * t)
	}
}

func (s *Loop) multiTimer() {
	t := time.Duration(s.userData.Interval)
	for {
		s.notice <- 0
		time.Sleep(time.Second * t)
	}
}
