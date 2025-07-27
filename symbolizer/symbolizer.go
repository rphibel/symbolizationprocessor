package symbolizer

// #cgo CFLAGS: -I/home/fedora/src/blazesym/capi/include
// #cgo LDFLAGS: -L/home/fedora/src/blazesym/target/debug -l:libblazesym_c.a
/*
#include "blazesym.h"
struct blaze_sym* get_result(blaze_syms* res, size_t pos) {
	return &res->syms[pos];
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type symbolizer struct {
	sym *C.blaze_symbolizer
}

type CodeInfo struct {
	Dir    string
	File   string
	Line   int32
	Column int32
}

type Symbol struct {
	Name       string
	Module     string
	CodeInfo   CodeInfo
}
// NewSymbolizer creates a new symbolizer instance.
func NewSymbolizer() *symbolizer {
	return &symbolizer{
		sym: C.blaze_symbolizer_new(),
	}
}

func (s *symbolizer) Free() {
	if s.sym != nil {
		C.blaze_symbolizer_free(s.sym)
		s.sym = nil
	}
}	

func (s *symbolizer) Symbolize(ipid int, iaddr uint64) (*Symbol, error) {
	pid := C.uint32_t(ipid)
	addr := C.uint64_t(iaddr)
	stack := C.malloc(C.sizeof_uint64_t * 1)
	defer C.free(unsafe.Pointer(stack))
	*(*C.uint64_t)(stack) = addr
	debug_syms := C.bool(true) // Enable debug symbols
	src := C.struct_blaze_symbolize_src_process {
			type_size: C.sizeof_struct_blaze_symbolize_src_process,
			pid: pid,
			debug_syms: debug_syms,
	}
	syms := C.blaze_symbolize_process_abs_addrs(s.sym, &src, (*C.uint64_t)(stack), 1)
	if (syms == nil) {
		return nil, fmt.Errorf("no symbols found")
	}
	sym := C.get_result(syms, C.size_t(0))
	return &Symbol{
		Name:   C.GoString(sym.name),
		Module: C.GoString(sym.module),
		CodeInfo: CodeInfo{
			Dir:    C.GoString(sym.code_info.dir),
			File:   C.GoString(sym.code_info.file),
			Line:   int32(sym.code_info.line),
			Column: int32(sym.code_info.column),
		},
	}, nil
}
