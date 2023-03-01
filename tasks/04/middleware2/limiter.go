package middleware

import (
	"net/http"
	"sync"
)

type LimiterMiddleware struct {
	limiter *Limiter
	handler *http.Handler
}

func (limHandler LimiterMiddleware) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if (*(limHandler.limiter)).TryAcquire() == true {
		(*(limHandler.handler)).ServeHTTP(writer, request)
	} else {
		writer.WriteHeader(http.StatusTooManyRequests)
	}
}

func Limit(l Limiter) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return LimiterMiddleware{handler: &handler, limiter: &l}
	}
}

type Limiter interface {
	TryAcquire() bool
	Release()
}

type MutexLimiter struct {
	mutex   *sync.Mutex
	cap     int
	current int
}

func NewMutexLimiter(count int) *MutexLimiter {
	return &MutexLimiter{mutex: &sync.Mutex{}, cap: count, current: 0}
}

func (l *MutexLimiter) TryAcquire() bool {
	defer l.mutex.Unlock()
	l.mutex.Lock()
	if l.current < l.cap {
		l.current++
		return true
	}
	return false
}

func (l *MutexLimiter) Release() {
	defer l.mutex.Unlock()
	l.mutex.Lock()
	l.current--
}

type ChanLimiter struct {
	chSem chan interface{}
}

func NewChanLimiter(count int) *ChanLimiter {
	return &ChanLimiter{chSem: make(chan interface{}, count)}
}

func (l *ChanLimiter) TryAcquire() bool {
	select {
	case l.chSem <- struct{}{}:
		return true
	default:
		return false
	}
}

func (l *ChanLimiter) Release() {
	<-l.chSem
}
