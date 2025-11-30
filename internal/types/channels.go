package types

import (
	"fmt"

	ci "github.com/kubex-ecosystem/gdbase/internal/interfaces"

	logz "github.com/kubex-ecosystem/logz"

	"reflect"

	"github.com/google/uuid"
)

var (
	smBuf, mdBuf, lgBuf = GetDefaultBufferSizes()
)

type ChannelBase[T any] struct {
	*logz.LoggerZ              // Logger for this Channel instance
	*Mutexes                   // Mutexes for this Channel instance
	Name          string       // The name of the channel.
	Channel       any          // The channel for the value. Main channel for this struct.
	Type          reflect.Type // The type of the channel.
	Buffers       int          // The number of buffers for the channel.
	Shared        interface{}  // Shared data for many purposes
}

// NewChannelBase creates a new ChannelBase instance with the provided name and type.
func NewChannelBase[T any](name string, buffers int, logger *logz.LoggerZ) ci.IChannelBase[any] {
	// if logger == nil {
	logger = logz.NewLogger("GoLife")
	// }
	mu := NewMutexesType()
	if buffers <= 0 {
		buffers = lgBuf
	}
	return &ChannelBase[any]{
		LoggerZ: logger,
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
		logz.Log("info", "Closing channel for:", cb.Name)
		close(cb.Channel.(chan T))
	}
	return nil
}
func (cb *ChannelBase[T]) Clear() error {
	cb.MuLock()
	defer cb.MuUnlock()
	if cb.Channel != nil {
		logz.Log("info", "Clearing channel for:", cb.Name)
		close(cb.Channel.(chan T))
		cb.Channel = make(chan T, cb.Buffers)
	}
	return nil
}

type ChannelCtl[T any] struct {
	// IChannelCtl is the interface for this Channel instance.
	//ci.IChannelCtl[T] // Channel interface for this Channel instance

	// Logger is the Logger instance for this Channel instance.
	Logger *logz.LoggerZ // Logger for this Channel instance

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
func NewChannelCtl[T any](name string, logger *logz.LoggerZ) ci.IChannelCtl[T] {
	// if logger == nil {
	logger = logz.NewLogger("GoLife")
	// }
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
func NewChannelCtlWithProperty[T any, P ci.IProperty[T]](name string, buffers *int, property P, withMetrics bool, logger *logz.LoggerZ) ci.IChannelCtl[T] {
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
		logz.Log("info", "Creating channels map for:", cCtl.Name, "ID:", cCtl.ID.String())
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	if rawChannel, ok := cCtl.Channels[name]; ok {
		if channel, ok := rawChannel.(ci.IChannelBase[T]); ok {
			return channel, channel.GetType(), true
		} else {
			logz.Log("error", fmt.Sprintf("Channel %s is not a valid channel type. Expected: %s, receive %s", name, reflect.TypeFor[ci.IChannelBase[T]]().String(), reflect.TypeOf(rawChannel)))
			return nil, nil, false
		}
	}
	logz.Log("error", "Channel not found:", name, "ID:", cCtl.ID.String())
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
	logger := logz.GetLoggerZ("GoLife")

	if v.Channels == nil {
		v.MuLock()
		defer v.MuUnlock()
		logz.Log("ID:", v.ID.String())
		v.Channels = make(map[string]interface{})
		// done is a channel for the done signal.
		v.Channels["done"] = NewChannelBase[bool]("done", smBuf, logger)
		// ctl is a channel for the internal control channel.
		v.Channels["ctl"] = NewChannelBase[string]("ctl", mdBuf, logger)
		// condition is a channel for the condition signal.
		v.Channels["condition"] = NewChannelBase[string]("cond", smBuf, logger)

		if v.withMetrics {
			v.Channels["telemetry"] = NewChannelBase[string]("telemetry", mdBuf, logger)
			v.Channels["monitor"] = NewChannelBase[string]("monitor", mdBuf, logger)
		}
	}
	return v.Channels
}
func getDefaultChannelsMap(withMetrics bool, logger *logz.LoggerZ) map[string]any {
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

func chanRoutineCtl[T any](v *ChannelCtl[T], chCtl chan string, ch chan T) {
	// Evita usar select com um único case — faz uma recepção direta.
	if chCtl == nil {
		logz.Log("debug", "Control channel is nil for:", v.GetName(), "ID:", v.GetID().String())
		return
	}

	msg := <-chCtl
	switch msg {
	case "stop":
		if ch != nil {
			logz.Log("debug", "Stopping channel for:", v.GetName(), "ID:", v.GetID().String())
			ch <- v.GetProperty().GetValue()
			return
		}
	case "get":
		if ch != nil {
			logz.Log("debug", "Getting value from channel for:", v.GetName(), "ID:", v.GetID().String())
			ch <- v.GetProperty().GetValue()
		}
	case "set":
		if ch != nil {
			logz.Log("debug", "Waiting for value from channel for:", v.GetName(), "ID:", v.GetID().String())
			nVal := <-ch
			if reflect.ValueOf(nVal).IsValid() {
				if reflect.ValueOf(nVal).CanConvert(reflect.TypeFor[T]()) {
					logz.Log("debug", "Setting value from channel for:", v.GetName(), "ID:", v.GetID().String())
					v.GetProperty().SetValue(&nVal)
				} else {
					logz.Log("error", "Set: invalid type for channel value (", reflect.TypeFor[T]().String(), ")")
				}
			}
		}
	case "save":
		if ch != nil {
			logz.Log("debug", "Saving value from channel for:", v.GetName(), "ID:", v.GetID().String())
			nVal := <-ch
			if reflect.ValueOf(nVal).IsValid() {
				if reflect.ValueOf(nVal).CanConvert(reflect.TypeFor[T]()) {
					logz.Log("debug", "Saving value from channel for:", v.GetName(), "ID:", v.GetID().String())
					v.GetProperty().SetValue(&nVal)
				} else {
					logz.Log("error", "Save: invalid type for channel value (", reflect.TypeFor[T]().String(), ")")
				}
			}
		}
	case "clear":
		if ch != nil {
			logz.Log("debug", "Clearing channel for:", v.GetName(), "ID:", v.GetID().String())
			v.GetProperty().SetValue(nil)
		}
	}
}
func chanRoutineDefer[T any](v *ChannelCtl[T], chCtl chan string, ch chan T) {
	if r := recover(); r != nil {
		logz.Log("error", "Recovering from panic in monitor routine for:", v.GetName(), "ID:", v.GetID().String(), "Error:", fmt.Sprintf("%v", r))
		// Não recriamos canais aqui porque a atribuição local não afeta o estado
		// externo (são parâmetros pass-by-value). Alocar canais apenas localmente
		// gerava um warning ("value is never used") e não ajudava na recuperação.
		// Se necessário, a lógica de re-criação deve ocorrer no código que mantém
		// referências aos canais (ex: dentro do objeto v).
	} else {
		logz.Log("debug", "Exiting monitor routine for:", v.GetName(), "ID:", v.GetID().String())
		// When the monitor routine is done, we need to close the channels.
		// If the channel is not nil, close it.
		if ch != nil {
			close(ch)
		}
		if chCtl != nil {
			close(chCtl)
		}
		ch = nil
		chCtl = nil
	}
	// Always check the v mutexes to see if someone is locking the mutex or not on exit (defer).
	if v.MuTryLock() {
		// If the mutex was locked, unlock it.
		v.MuUnlock()
	} else if v.MuTryRLock() {
		// If the mutex was locked, unlock it.
		v.MuRUnlock()
	}
}
func chanRoutineWrapper[T any](v *ChannelCtl[T]) {
	logz.Log("debug", "Setting monitor routine for:", v.GetName(), "ID:", v.GetID().String())
	if rawChCtl, chCtlType, chCtlOk := v.GetSubChannelByName("ctl"); !chCtlOk {
		logz.Log("error", "ChannelCtl: no control channel found")
		return
	} else {
		if chCtlType != reflect.TypeOf("string") {
			logz.Log("error", "ChannelCtl: control channel is not a string channel")
			return
		}
		chCtl := reflect.ValueOf(rawChCtl).Interface().(chan string)
		rawCh := v.GetMainChannel()
		ch := reflect.ValueOf(rawCh).Interface().(chan T)

		defer chanRoutineDefer[T](v, chCtl, ch)
		for {
			chanRoutineCtl[T](v, chCtl, ch)
			if ch == nil {
				logz.Log("debug", "Channel is nil for:", v.GetName(), "ID:", v.GetID().String(), "Exiting monitor routine")
				break
			}
			if chCtl == nil {
				logz.Log("debug", "Control channel is nil for:", v.GetName(), "ID:", v.GetID().String(), "Exiting monitor routine")
				break
			}
		}
	}
}

func GetDefaultBufferSizes() (sm, md, lg int) { return 2, 5, 10 }
