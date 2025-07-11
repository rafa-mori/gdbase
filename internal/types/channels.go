package types

import (
	"fmt"

	l "github.com/rafa-mori/logz"
	ci "github.com/rafa-mori/gdbase/internal/interfaces"
	gl "github.com/rafa-mori/gdbase/logger"
	tu "github.com/rafa-mori/gdbase/utils"

	"reflect"

	"github.com/google/uuid"
)

var (
	smBuf, mdBuf, lgBuf = tu.GetDefaultBufferSizes()
)

type ChannelBase[T any] struct {
	*Mutexes              // Mutexes for this Channel instance
	Name     string       // The name of the channel.
	Channel  any          // The channel for the value. Main channel for this struct.
	Type     reflect.Type // The type of the channel.
	Buffers  int          // The number of buffers for the channel.
	Shared   interface{}  // Shared data for many purposes
}

// NewChannelBase creates a new ChannelBase instance with the provided name and type.
func NewChannelBase[T any](name string, buffers int, logger l.Logger) ci.IChannelBase[any] {
	if logger == nil {
		logger = l.GetLogger("GoLife")
	}
	mu := NewMutexesType()
	if buffers <= 0 {
		buffers = lgBuf
	}
	return &ChannelBase[any]{
		Mutexes: mu,
		Name:    name,
		Channel: make(chan T, buffers),
		Type:    reflect.TypeFor[T](),
		Buffers: buffers,
	}
}

func (cb *ChannelBase[T]) GetName() string {
	cb.MuRLock()
	defer cb.MuRUnlock()
	return cb.Name
}
func (cb *ChannelBase[T]) GetChannel() (any, reflect.Type) {
	cb.MuRLock()
	defer cb.MuRUnlock()
	return cb.Channel, reflect.TypeOf(cb.Channel)
}
func (cb *ChannelBase[T]) GetType() reflect.Type {
	cb.MuRLock()
	defer cb.MuRUnlock()
	return cb.Type
}
func (cb *ChannelBase[T]) GetBuffers() int {
	cb.MuRLock()
	defer cb.MuRUnlock()
	return cb.Buffers
}
func (cb *ChannelBase[T]) SetName(name string) string {
	cb.MuLock()
	defer cb.MuUnlock()
	cb.Name = name
	return cb.Name
}
func (cb *ChannelBase[T]) SetChannel(typE reflect.Type, bufferSize int) any {
	cb.MuLock()
	defer cb.MuUnlock()
	cb.Channel = reflect.MakeChan(typE, bufferSize)
	return cb.Channel
}
func (cb *ChannelBase[T]) SetBuffers(buffers int) int {
	cb.MuLock()
	defer cb.MuUnlock()
	cb.Buffers = buffers
	cb.Channel = make(chan T, buffers)
	return cb.Buffers
}
func (cb *ChannelBase[T]) Close() error {
	cb.MuLock()
	defer cb.MuUnlock()
	if cb.Channel != nil {
		gl.LogObjLogger(cb, "info", "Closing channel for:", cb.Name)
		close(cb.Channel.(chan T))
	}
	return nil
}
func (cb *ChannelBase[T]) Clear() error {
	cb.MuLock()
	defer cb.MuUnlock()
	if cb.Channel != nil {
		gl.LogObjLogger(cb, "info", "Clearing channel for:", cb.Name)
		close(cb.Channel.(chan T))
		cb.Channel = make(chan T, cb.Buffers)
	}
	return nil
}

type ChannelCtl[T any] struct {
	// IChannelCtl is the interface for this Channel instance.
	//ci.IChannelCtl[T] // Channel interface for this Channel instance

	// Logger is the Logger instance for this Channel instance.
	Logger l.Logger // Logger for this Channel instance

	// IMutexes is the interface for the mutexes in this Channel instance.
	*Mutexes // Mutexes for this Channel instance

	// property is the property for the channel.
	property ci.IProperty[T] // Lazy load, only used when needed or created by NewChannelCtlWithProperty constructor

	// Shared is a shared data used for many purposes like sync.Cond, Telemetry, Monitor, etc.
	Shared interface{} // Shared data for many purposes

	withMetrics bool // If true, will create the telemetry and monitor channels

	// ch is a channel for the value.
	ch chan T // The channel for the value. Main channel for this struct.

	// Reference is the reference ID and name.
	*Reference `json:"reference" yaml:"reference" xml:"reference" gorm:"reference"`

	// buffers is the number of buffers for the channel.
	Buffers int `json:"buffers" yaml:"buffers" xml:"buffers" gorm:"buffers"`

	Channels map[string]any `json:"channels,omitempty" yaml:"channels,omitempty" xml:"channels,omitempty" gorm:"channels,omitempty"`
}

// NewChannelCtl creates a new ChannelCtl instance with the provided name.
func NewChannelCtl[T any](name string, logger l.Logger) ci.IChannelCtl[T] {
	if logger == nil {
		logger = l.GetLogger("GoLife")
	}
	ref := NewReference(name)
	mu := NewMutexesType()

	// Create a new ChannelCtl instance
	channelCtl := &ChannelCtl[T]{
		Logger:    logger,
		Reference: ref.GetReference(),
		Mutexes:   mu,
		ch:        make(chan T, lgBuf),
		Channels:  make(map[string]any),
	}
	channelCtl.Channels = getDefaultChannelsMap(false, logger)
	return channelCtl
}

// NewChannelCtlWithProperty creates a new ChannelCtl instance with the provided name and type.
func NewChannelCtlWithProperty[T any, P ci.IProperty[T]](name string, buffers *int, property P, withMetrics bool, logger l.Logger) ci.IChannelCtl[T] {
	if logger == nil {
		logger = l.GetLogger("GoLife")
	}
	ref := NewReference(name)
	mu := NewMutexesType()
	buf := 3
	if buffers != nil {
		buf = *buffers
	}
	channelCtl := &ChannelCtl[T]{
		Logger:    logger,
		Reference: ref.GetReference(),
		Mutexes:   mu,
		ch:        make(chan T, buf),
		Channels:  make(map[string]any),
		property:  property,
	}
	channelCtl.Channels = getDefaultChannelsMap(withMetrics, logger)

	return channelCtl
}

func (cCtl *ChannelCtl[T]) GetID() uuid.UUID {
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return cCtl.ID
}
func (cCtl *ChannelCtl[T]) GetName() string {
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return cCtl.Name
}
func (cCtl *ChannelCtl[T]) SetName(name string) string {
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.Name = name
	return cCtl.Name
}
func (cCtl *ChannelCtl[T]) GetProperty() ci.IProperty[T] {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return cCtl.property
}
func (cCtl *ChannelCtl[T]) GetSubChannels() map[string]interface{} {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return cCtl.Channels
}
func (cCtl *ChannelCtl[T]) SetSubChannels(channels map[string]interface{}) map[string]interface{} {
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	for k, v := range channels {
		if _, ok := cCtl.Channels[k]; ok {
			cCtl.Channels[k] = v
		} else {
			cCtl.Channels[k] = v
		}
	}
	return cCtl.Channels
}
func (cCtl *ChannelCtl[T]) GetSubChannelByName(name string) (any, reflect.Type, bool) {
	if cCtl.Channels == nil {
		gl.LogObjLogger(cCtl, "info", "Creating channels map for:", cCtl.Name, "ID:", cCtl.ID.String())
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	if rawChannel, ok := cCtl.Channels[name]; ok {
		if channel, ok := rawChannel.(ci.IChannelBase[T]); ok {
			return channel, channel.GetType(), true
		} else {
			gl.LogObjLogger(cCtl, "error", fmt.Sprintf("Channel %s is not a valid channel type. Expected: %s, receive %s", name, reflect.TypeFor[ci.IChannelBase[T]]().String(), reflect.TypeOf(rawChannel)))
			return nil, nil, false
		}
	}
	gl.LogObjLogger(cCtl, "error", "Channel not found:", name, "ID:", cCtl.ID.String())
	return nil, nil, false
}
func (cCtl *ChannelCtl[T]) SetSubChannelByName(name string, channel any) (any, error) {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	if _, ok := cCtl.Channels[name]; ok {
		cCtl.Channels[name] = channel
	} else {
		cCtl.Channels[name] = channel
	}
	return channel, nil
}
func (cCtl *ChannelCtl[T]) GetSubChannelTypeByName(name string) (reflect.Type, bool) {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	if channel, ok := cCtl.Channels[name]; ok {
		return channel.(ci.IChannelBase[any]).GetType(), true
	}
	return nil, false
}
func (cCtl *ChannelCtl[T]) GetSubChannelBuffersByName(name string) (int, bool) {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	if channel, ok := cCtl.Channels[name]; ok {
		return channel.(ci.IChannelBase[any]).GetBuffers(), true
	}
	return 0, false
}
func (cCtl *ChannelCtl[T]) SetSubChannelBuffersByName(name string, buffers int) (int, error) {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	if channel, ok := cCtl.Channels[name]; ok {
		channel.(ci.IChannelBase[any]).SetBuffers(buffers)
		return buffers, nil
	}
	return 0, nil
}
func (cCtl *ChannelCtl[T]) GetMainChannel() any {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return cCtl.ch
}
func (cCtl *ChannelCtl[T]) SetMainChannel(channel chan T) chan T {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.ch = channel
	return cCtl.ch
}
func (cCtl *ChannelCtl[T]) GetMainChannelType() reflect.Type {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return reflect.TypeOf(cCtl.ch)
}
func (cCtl *ChannelCtl[T]) GetHasMetrics() bool {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return cCtl.withMetrics
}
func (cCtl *ChannelCtl[T]) SetHasMetrics(hasMetrics bool) bool {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.withMetrics = hasMetrics
	return cCtl.withMetrics
}
func (cCtl *ChannelCtl[T]) GetBufferSize() int {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return cCtl.Buffers
}
func (cCtl *ChannelCtl[T]) SetBufferSize(size int) int {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.Buffers = size
	return cCtl.Buffers
}
func (cCtl *ChannelCtl[T]) Close() error {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	if cCtl.Channels != nil {
		for _, channel := range cCtl.Channels {
			if ch, ok := channel.(ci.IChannelBase[any]); ok {
				_ = ch.Close()
			}
		}
	}
	return nil
}
func (cCtl *ChannelCtl[T]) WithProperty(property ci.IProperty[T]) ci.IChannelCtl[T] {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.property = property
	return cCtl
}
func (cCtl *ChannelCtl[T]) WithChannel(channel chan T) ci.IChannelCtl[T] {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.ch = channel
	return cCtl
}
func (cCtl *ChannelCtl[T]) WithBufferSize(size int) ci.IChannelCtl[T] {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.Buffers = size
	return cCtl
}
func (cCtl *ChannelCtl[T]) WithMetrics(metrics bool) ci.IChannelCtl[T] {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.withMetrics = metrics
	return cCtl
}

func initChannelsMap[T any](v *ChannelCtl[T]) map[string]interface{} {
	if v.Channels == nil {
		v.MuLock()
		defer v.MuUnlock()
		gl.LogObjLogger(v, "info", "Creating channels map for:", v.Name, "ID:", v.ID.String())
		v.Channels = make(map[string]interface{})
		// done is a channel for the done signal.
		v.Channels["done"] = NewChannelBase[bool]("done", smBuf, v.Logger)
		// ctl is a channel for the internal control channel.
		v.Channels["ctl"] = NewChannelBase[string]("ctl", mdBuf, v.Logger)
		// condition is a channel for the condition signal.
		v.Channels["condition"] = NewChannelBase[string]("cond", smBuf, v.Logger)

		if v.withMetrics {
			v.Channels["telemetry"] = NewChannelBase[string]("telemetry", mdBuf, v.Logger)
			v.Channels["monitor"] = NewChannelBase[string]("monitor", mdBuf, v.Logger)
		}
	}
	return v.Channels
}
func getDefaultChannelsMap(withMetrics bool, logger l.Logger) map[string]any {
	mp := map[string]any{
		// done is a channel for the done signal.
		"done": NewChannelBase[bool]("done", smBuf, logger),
		// ctl is a channel for the internal control channel.
		"ctl": NewChannelBase[string]("ctl", mdBuf, logger),
		// condition is a channel for the condition signal.
		"condition": NewChannelBase[string]("cond", smBuf, logger),
	}

	if withMetrics {
		// metrics is a channel for the telemetry signal.
		mp["metrics"] = NewChannelBase[string]("metrics", mdBuf, logger)
		// monitor is a channel for monitoring the channel.
		mp["monitor"] = NewChannelBase[string]("monitor", mdBuf, logger)
	}

	return mp
}
