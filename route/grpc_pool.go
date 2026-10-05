package route

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const methodCacheTTL = 5 * time.Minute

// grpcConns shares one lazily-connecting, self-reconnecting ClientConn per address.
type grpcConns struct {
	mu    sync.Mutex
	conns map[string]*grpc.ClientConn
}

var conns = &grpcConns{conns: make(map[string]*grpc.ClientConn)}

func (p *grpcConns) get(addr string) (*grpc.ClientConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if conn, ok := p.conns[addr]; ok {
		return conn, nil
	}
	target := addr
	if !strings.Contains(addr, "://") {
		target = "passthrough:///" + addr // keep the resolver behaviour of the old grpc.Dial
	}
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	p.conns[addr] = conn
	return conn, nil
}

type methodEntry struct {
	desc *desc.MethodDescriptor
	at   time.Time
}

// methodCache avoids a reflection round-trip on every request.
var methodCache sync.Map // key -> methodEntry

func resolveMethod(ctx context.Context, conn *grpc.ClientConn, addr, service, method string) (*desc.MethodDescriptor, error) {
	key := addr + "|" + service + "|" + method
	if v, ok := methodCache.Load(key); ok {
		if e := v.(methodEntry); time.Since(e.at) < methodCacheTTL {
			return e.desc, nil
		}
	}

	client := grpcreflect.NewClientAuto(ctx, conn)
	defer client.Reset()

	svc, err := client.ResolveService(service)
	if err != nil {
		return nil, err
	}
	m := svc.FindMethodByName(method)
	if m == nil {
		return nil, nil
	}
	methodCache.Store(key, methodEntry{desc: m, at: time.Now()})
	return m, nil
}
