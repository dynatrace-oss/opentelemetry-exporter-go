package trace

import (
	"sync"
	"time"
)

type dtSpanSet map[*dtSpan]struct{}

type dtSpanWatchlist struct {
	spans    dtSpanSet
	lock     sync.Mutex
	maxSpans int
}

func newDtSpanWatchlist(watchlistSize int) dtSpanWatchlist {
	return dtSpanWatchlist{
		spans:    make(dtSpanSet),
		lock:     sync.Mutex{},
		maxSpans: watchlistSize,
	}
}

func (p *dtSpanWatchlist) add(s *dtSpan) bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	if len(p.spans) >= p.maxSpans {
		return false
	}

	p.spans[s] = struct{}{}
	return true
}

func (p *dtSpanWatchlist) remove(s *dtSpan) {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.spans, s)
}

func (p *dtSpanWatchlist) contains(s *dtSpan) bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	_, found := p.spans[s]
	return found
}

func (p *dtSpanWatchlist) len() int {
	p.lock.Lock()
	defer p.lock.Unlock()

	return len(p.spans)
}

// getSpansToExport evaluates available spans and returns those that have to be sent
func (p *dtSpanWatchlist) getSpansToExport() dtSpanSet {
	p.lock.Lock()
	// make a copy of the spans map to avoid keeping the lock for a longer period of time
	spansToExport := make(dtSpanSet)
	for k, v := range p.spans {
		spansToExport[k] = v
	}
	p.lock.Unlock()

	now := time.Now().UnixNano() / int64(time.Millisecond)
	for span := range spansToExport {
		prepareResult := span.prepareSend(now)

		if prepareResult == prepareResultSkip {
			delete(spansToExport, span)
		}

		if prepareResult == prepareResultDrop ||
			(prepareResult == prepareResultSend && span.metadata.sendState == sendStateSpanEnded) {
			p.remove(span)
		}
	}

	return spansToExport
}
