package veritas

import "github.com/google/cel-go/cel"

type envLib struct {
	opts []cel.EnvOption
}

func (l *envLib) CompileOptions() []cel.EnvOption {
	return l.opts
}

func (l *envLib) ProgramOptions() []cel.ProgramOption {
	return nil
}

func newEnvLib(opts ...cel.EnvOption) cel.EnvOption {
	return cel.Lib(&envLib{opts: opts})
}
