package rxproto

import (
	"context"
	"errors"

	"google.golang.org/protobuf/proto"
)

var (
	ErrObserverHasUnsubscribed = errors.New("observer has unsubscribed")
)

type RxProtoObserver interface {
	Next(data []byte)
	Error(err error)
	Completed()
}

type RxGoProtoObserver struct {
	RxProtoObserver
	ctx context.Context
}

func (observer *RxGoProtoObserver) Next(message proto.Message) error {
	bytes, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	select {
	case <-observer.ctx.Done():
		return ErrObserverHasUnsubscribed
	default:
		observer.RxProtoObserver.Next(bytes)
	}

	return nil
}
func (observer *RxGoProtoObserver) Error(err error) error {
	select {
	case <-observer.ctx.Done():
		return ErrObserverHasUnsubscribed
	default:
		observer.RxProtoObserver.Error(err)
	}

	return nil
}

func (observer *RxGoProtoObserver) Completed(err error) error {
	select {
	case <-observer.ctx.Done():
		return ErrObserverHasUnsubscribed
	default:
		observer.RxProtoObserver.Completed()
	}

	return nil
}

type RxContext struct {
	Ctx    context.Context
	cancel context.CancelFunc
}

func newRxContext(parentCtx context.Context) *RxContext {
	ctx, cancel := context.WithCancel(parentCtx)

	return &RxContext{
		Ctx:    ctx,
		cancel: cancel,
	}
}

func (c *RxContext) Cancel() {
	c.cancel()
}

type RxProtoObservable struct {
	nextCh         chan proto.Message
	errCh          chan error
	completeCh     chan struct{}
	newMessageFunc func() proto.Message
}

func NewRxProtoObservable() *RxProtoObservable {
	return &RxProtoObservable{
		nextCh:     make(chan proto.Message),
		errCh:      make(chan error),
		completeCh: make(chan struct{}),
	}
}

func (observer *RxProtoObservable) WithNew(function RxProtoMessageFunc) {
	observer.newMessageFunc = function
}

func (observer *RxProtoObservable) Next() <-chan proto.Message {
	return observer.nextCh
}

func (observer *RxProtoObservable) Error() <-chan error {
	return observer.errCh
}

func (observer *RxProtoObservable) Completed() <-chan struct{} {
	return observer.completeCh
}

func (observer *RxProtoObservable) WriteNext(data []byte) {
	message := observer.newMessageFunc()

	_ = proto.Unmarshal(data, message)

	observer.nextCh <- message
}

func (observer *RxProtoObservable) WriteError(err error) {
	observer.errCh <- err
}

func (observer *RxProtoObservable) Complete() {
	close(observer.completeCh)
}

type RxProtoMessageFunc func() proto.Message

type RxProtoUnaryFunc func(ctx context.Context, input proto.Message, observer *RxGoProtoObserver) error

type RxProtoStreamFunc func(ctx context.Context, input *RxProtoObservable, observer *RxGoProtoObserver) error

func RxProtoUnary(
	parentCtx context.Context,
	input []byte,
	observer RxProtoObserver,
	messageFunc RxProtoMessageFunc,
	callbackFunc RxProtoUnaryFunc,
) *RxContext {
	rxCtx := newRxContext(parentCtx)

	var messageInput proto.Message
	if messageFunc != nil {
		messageInput = messageFunc()
	}

	if input != nil {
		messageInput = messageFunc()
		_ = proto.Unmarshal(input, messageInput)
	}

	go func() {
		var goObserver *RxGoProtoObserver

		if observer != nil {
			defer observer.Completed()

			goObserver = &RxGoProtoObserver{
				RxProtoObserver: observer,
				ctx:             rxCtx.Ctx,
			}
		}

		if err := callbackFunc(rxCtx.Ctx, messageInput, goObserver); err != nil && err != ErrObserverHasUnsubscribed {
			if observer != nil {
				observer.Error(err)
			}
		}
	}()

	return rxCtx
}

func RxProtoStream(
	parentCtx context.Context,
	input *RxProtoObservable,
	observer RxProtoObserver,
	messageFunc RxProtoMessageFunc,
	callbackFunc RxProtoStreamFunc,
) *RxContext {
	rxCtx := newRxContext(parentCtx)
	input.WithNew(messageFunc)

	go func() {
		var goObserver *RxGoProtoObserver

		if observer != nil {
			defer observer.Completed()

			goObserver = &RxGoProtoObserver{
				RxProtoObserver: observer,
				ctx:             rxCtx.Ctx,
			}
		}

		if err := callbackFunc(rxCtx.Ctx, input, goObserver); err != nil && err != ErrObserverHasUnsubscribed {
			if observer != nil {
				observer.Error(err)
			}
		}
	}()

	return rxCtx
}
