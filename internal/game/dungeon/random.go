package dungeon

import "math/rand"

type trackedRandomSource struct {
	source rand.Source64
	draws  uint64
}

func newTrackedRandomSource(seed int64) *trackedRandomSource {
	return &trackedRandomSource{source: rand.NewSource(seed).(rand.Source64)}
}

func newTrackedRandomSourceAt(seed int64, draws uint64) *trackedRandomSource {
	source := newTrackedRandomSource(seed)
	// ponytail: restore by replaying the cursor; serialize the source when save-time latency matters.
	for i := uint64(0); i < draws; i++ {
		source.source.Uint64()
	}
	source.draws = draws
	return source
}

func (s *trackedRandomSource) rand() *rand.Rand {
	return rand.New(s)
}

func (s *trackedRandomSource) Int63() int64 {
	s.draws++
	return s.source.Int63()
}

func (s *trackedRandomSource) Uint64() uint64 {
	s.draws++
	return s.source.Uint64()
}

func (s *trackedRandomSource) Seed(seed int64) {
	s.source.Seed(seed)
	s.draws = 0
}
