package log

type nullLogger struct{}

// Debugf ...
func (n *nullLogger) Debugf(_ string, _ ...interface{}) {}

// Infof ...
func (n *nullLogger) Infof(_ string, _ ...interface{}) {}

// Warnf ...
func (n *nullLogger) Warnf(_ string, _ ...interface{}) {}

// Errorf ...
func (n *nullLogger) Errorf(_ string, _ ...interface{}) {}

// Panicf ...
func (n *nullLogger) Panicf(_ string, _ ...interface{}) {}

// Fatalf ...
func (n *nullLogger) Fatalf(_ string, _ ...interface{}) {}

// Debugw ...
func (n *nullLogger) Debugw(_ string, _ ...interface{}) {}

// Infow ...
func (n *nullLogger) Infow(_ string, _ ...interface{}) {}

// Warnw ...
func (n *nullLogger) Warnw(_ string, _ ...interface{}) {}

// Errorw ...
func (n *nullLogger) Errorw(_ string, _ ...interface{}) {}

// Panicw ...
func (n *nullLogger) Panicw(_ string, _ ...interface{}) {}

// Fatalw ...
func (n *nullLogger) Fatalw(_ string, _ ...interface{}) {}

// Debug ...
func (n *nullLogger) Debug(_ ...interface{}) {}

// Info ...
func (n *nullLogger) Info(_ ...interface{}) {}

// Warn ...
func (n *nullLogger) Warn(_ ...interface{}) {}

// Error ...
func (n *nullLogger) Error(_ ...interface{}) {}

// Fatal ...
func (n *nullLogger) Fatal(_ ...interface{}) {}

// Sync ...
func (n *nullLogger) Sync() {}

var _ Logger = (*nullLogger)(nil)

func NewNopLogger() Logger {
	return &nullLogger{}
}
