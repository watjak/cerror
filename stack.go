package cerror

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// Frame represents a program counter inside a stack frame.
// The value of Frame as uintptr represents the program counter + 1.
type Frame uintptr

// pc returns the program counter for this frame.
func (f Frame) pc() uintptr {
	return uintptr(f) - 1
}

// file returns the full path to the file that contains the function for this Frame's pc.
func (f Frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

// line returns the line number of the source code for this Frame's pc.
func (f Frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

// name returns the name of the function for this Frame's pc.
func (f Frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

// Format formats the frame according to the fmt.Formatter interface.
//
//	%s    source file
//	%d    source line
//	%n    function name
//	%v    equivalent to %s:%d
//	%+s   function name and file path relative to the compile-time GOPATH
//	%+v   equivalent to %+s:%d
func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		if s.Flag('+') {
			io.WriteString(s, f.name()+"\n\t"+f.file())
		} else {
			io.WriteString(s, path.Base(f.file()))
		}
	case 'd':
		io.WriteString(s, strconv.Itoa(f.line()))
	case 'n':
		io.WriteString(s, funcname(f.name()))
	case 'v':
		if s.Flag('+') {
			f.Format(s, 's')
			io.WriteString(s, ":")
			f.Format(s, 'd')
		} else {
			f.Format(s, 's')
		}
	}
}

// MarshalText formats a stack trace Frame as text.
func (f Frame) MarshalText() ([]byte, error) {
	name := f.name()
	if name == "unknown" {
		return []byte(name), nil
	}
	return []byte(fmt.Sprintf("%s %s:%d", name, f.file(), f.line())), nil
}

// StackTrace is a stack of Frames from innermost (newest) to outermost (oldest).
type StackTrace []Frame

// Format formats the stack of Frames according to the fmt.Formatter interface.
func (st StackTrace) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			for _, f := range st {
				io.WriteString(s, "\n")
				f.Format(s, verb)
			}
		} else {
			st.formatSlice(s, verb)
		}
	case 's':
		st.formatSlice(s, verb)
	}
}

// formatSlice formats a StackTrace as a slice of Frames.
func (st StackTrace) formatSlice(s fmt.State, verb rune) {
	io.WriteString(s, "[")
	for i, f := range st {
		if i > 0 {
			io.WriteString(s, " ")
		}
		f.Format(s, verb)
	}
	io.WriteString(s, "]")
}

func (st StackTrace) String() string {
	formated := ""
	for _, f := range st {
		formated += fmt.Sprintf("%+v\n", f)
	}
	return formated
}

// stack represents a stack of program counters.
type stack []uintptr

// Format formats a stack according to the fmt.Formatter interface.
func (s *stack) Format(st fmt.State, verb rune) {
	if verb == 'v' && st.Flag('+') {
		for _, pc := range *s {
			f := Frame(pc)
			fmt.Fprintf(st, "\n%+v", f)
		}
	}
}

// StackTrace converts a stack into a StackTrace.
func (s *stack) StackTrace() StackTrace {
	frames := make([]Frame, len(*s))
	for i, pc := range *s {
		frames[i] = Frame(pc)
	}
	return frames
}

// callers captures the current call stack and returns it as a stack.
func callers() *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	st := pcs[:n]
	return (*stack)(&st)
}

// funcname extracts the function name from the fully qualified name.
func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}
