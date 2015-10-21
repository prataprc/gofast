package gofast

import "sync"

// Stream for a newly started stream on the transport.
type Stream struct {
	transport *Transport
	Rxch      chan Message
	opaque    uint64
	remote    bool
}

func fromrxstrmpool() *Stream {
	stream := rxstrmpool.Get().(*Stream)
	stream.transport, stream.Rxch, stream.opaque = nil, nil, 0
	stream.remote = false
	return stream
}

func (t *Transport) newstream(opaque uint64, remote bool) *Stream {
	stream := fromrxstrmpool()
	fmsg := "%v ##%d(remote:%v) stream created ...\n"
	log.Verbosef(fmsg, t.logprefix, opaque, remote)
	// reset all fields (it is coming from a pool)
	stream.transport, stream.remote, stream.opaque = t, remote, opaque
	stream.Rxch = nil
	return stream
}

func (t *Transport) getstream(ch chan Message) *Stream { // called only be tx.
	stream := <-t.strmpool
	stream.Rxch = ch
	t.putch(t.rxch, stream)
	return stream
}

func (t *Transport) putstream(opaque uint64, stream *Stream, tellrx bool) {
	func() {
		// Rxch could also be closed when transport is closed...
		// Rxch could also be nil in case of post...
		defer func() { recover() }()
	}()
	if stream == nil {
		log.Errorf("%v ##%v unkown stream\n", t.logprefix, opaque)
		return
	}
	if stream.Rxch != nil {
		close(stream.Rxch)
	}
	stream.Rxch = nil
	if tellrx {
		t.putch(t.rxch, stream)
	}
}

// Response to a request.
func (s *Stream) Response(msg Message, flush bool) error {
	obj := s.transport.p_txdata.Get()
	defer s.transport.p_txdata.Put(obj)

	out := obj.([]byte)
	n := s.transport.response(msg, s, out)
	return s.transport.txasync(out[:n], flush)
}

// Stream a message.
func (s *Stream) Stream(msg Message, flush bool) (err error) {
	obj := s.transport.p_txdata.Get()
	defer s.transport.p_txdata.Put(obj)

	out := obj.([]byte)
	n := s.transport.stream(msg, s, out)
	if err = s.transport.txasync(out[:n], flush); err != nil {
		s.transport.putstream(s.opaque, s, true /*tellrx*/)
	}
	return
}

// Close the stream.
func (s *Stream) Close() error {
	obj := s.transport.p_txdata.Get()
	defer s.transport.p_txdata.Put(obj)

	out := obj.([]byte)
	n := s.transport.finish(s, out)
	s.transport.putstream(s.opaque, s, true /*tellrx*/)
	return s.transport.txasync(out[:n], true /*flush*/)
}

// Transport return the underlying transport carrying this stream.
func (s *Stream) Transport() *Transport {
	return s.transport
}

var rxstrmpool *sync.Pool

func init() {
	rxstrmpool = &sync.Pool{New: func() interface{} { return &Stream{} }}
}
