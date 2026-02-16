package reqresp

import (
	"log"
	pb "newsgetter/contracts"
	"sync"

	"google.golang.org/protobuf/types/known/emptypb"
)

type OrphieStreamServer struct {
	pb.UnimplementedOrphieServiceServer
	DataChan <-chan *pb.OrphieDataResponse

	dialers   map[*DialerStream]bool
	dialersMu sync.RWMutex
}

type DialerStream struct {
	stream pb.OrphieService_GetOrphieDataServer
	done   chan bool
}

func (s *OrphieStreamServer) BroadcastDataToDialers() {
	log.Println("in broadcasting")
	for data := range s.DataChan {
		s.dialersMu.RLocker()

		for dialer := range s.dialers {
			go func(d *DialerStream, da *pb.OrphieDataResponse) {
				if err := d.stream.Send(da); err != nil {
					log.Panicln("couldn't send data to client, ", err)
					select {
					case d.done <- true:
					default: //god-untouched contraption
					}
				}
			}(dialer, data)
		}

		s.dialersMu.RUnlock()
	}
	log.Println("finished broadcasting datachan closed")
}

/*Returns a server with given input chan, binds with Broadcastingbdsm*/
func NewOrphieStreamServer(dataChan <-chan *pb.OrphieDataResponse) *OrphieStreamServer {
	server := &OrphieStreamServer{
		DataChan: dataChan,
		dialers:  make(map[*DialerStream]bool),
	}

	go server.BroadcastDataToDialers()

	return server
}

// Can't belive that they made a dir only for empty val
func (s *OrphieStreamServer) StreamData(req *emptypb.Empty, stream pb.OrphieService_GetOrphieDataServer) error {
	dialer := &DialerStream{
		stream: stream,
		done:   make(chan bool),
	}

	s.dialersMu.Lock()
	s.dialers[dialer] = true

	dialersCount := len(s.dialers)
	s.dialersMu.Unlock()

	log.Printf("By Uber's code tips dialers count now is stable and: %v", dialersCount)

	<-dialer.done

	s.dialersMu.Lock()
	delete(s.dialers, dialer)

	dialersCount = len(s.dialers)
	s.dialersMu.Unlock()

	log.Printf("That is not by uber specification, but dialers count: %v", dialersCount)

	return nil
}
