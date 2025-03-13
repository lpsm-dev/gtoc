package logger

import (
	"io"
	"os"
	"time"

	"github.com/charmbracelet/log"
)

var (
	// Instância global do logger
	defaultLogger *log.Logger

	// Configurações padrão
	defaultLevel     = log.WarnLevel
	defaultTimeFormat = time.Kitchen
	defaultCallerOffset = 0
	defaultShowCaller = false
	defaultPrefix     = "gtoc"
)

// Opções de configuração do logger
type Options struct {
	Level        log.Level
	TimeFormat   string
	Output       io.Writer
	ShowCaller   bool
	CallerOffset int
	Prefix       string
}

// Init inicializa o logger global com as opções personalizadas
func Init(opts *Options) {
	if opts == nil {
		opts = &Options{}
	}

	// Aplicar valores padrão para campos não definidos
	if opts.Level == 0 {
		opts.Level = defaultLevel
	}
	if opts.TimeFormat == "" {
		opts.TimeFormat = defaultTimeFormat
	}
	if opts.Output == nil {
		opts.Output = os.Stderr
	}
	if opts.Prefix == "" {
		opts.Prefix = defaultPrefix
	}

	// Criar e configurar o logger
	defaultLogger = log.NewWithOptions(opts.Output, log.Options{
		Level:        opts.Level,
		TimeFormat:   opts.TimeFormat,
		ReportCaller: opts.ShowCaller,
		CallerOffset: opts.CallerOffset,
		Prefix:       opts.Prefix,
	})
}

// GetLogger retorna a instância global do logger
func GetLogger() *log.Logger {
	if defaultLogger == nil {
		// Inicializar com valores padrão se ainda não foi inicializado
		Init(nil)
	}
	return defaultLogger
}

// SetLevel altera o nível de log
func SetLevel(level log.Level) {
	GetLogger().SetLevel(level)
}

// SetOutput altera o destino da saída de log
func SetOutput(output io.Writer) {
	GetLogger().SetOutput(output)
}

// SetTimeFormat altera o formato de tempo para os logs
func SetTimeFormat(format string) {
	GetLogger().SetTimeFormat(format)
}

// SetPrefix altera o prefixo dos logs
func SetPrefix(prefix string) {
	GetLogger().SetPrefix(prefix)
}

// Funções helper para facilitar o uso do logger

// Debug registra uma mensagem de debug
func Debug(msg interface{}, keyvals ...interface{}) {
	GetLogger().Debug(msg, keyvals...)
}

// Info registra uma mensagem de informação
func Info(msg interface{}, keyvals ...interface{}) {
	GetLogger().Info(msg, keyvals...)
}

// Warn registra uma mensagem de aviso
func Warn(msg interface{}, keyvals ...interface{}) {
	GetLogger().Warn(msg, keyvals...)
}

// Error registra uma mensagem de erro
func Error(msg interface{}, keyvals ...interface{}) {
	GetLogger().Error(msg, keyvals...)
}

// Fatal registra uma mensagem fatal e encerra a aplicação
func Fatal(msg interface{}, keyvals ...interface{}) {
	GetLogger().Fatal(msg, keyvals...)
}

// With retorna um novo logger com os valores chave-valor adicionados
func With(keyvals ...interface{}) *log.Logger {
	return GetLogger().With(keyvals...)
} 