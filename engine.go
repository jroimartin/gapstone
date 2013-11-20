package gapstone

// #cgo CFLAGS: -I/usr/include/capstone
// #cgo LDFLAGS: -lcapstone
// #include <stdlib.h>
// #include <capstone.h>
import "C"
import "unsafe"
import "reflect"
import "fmt"

type Arch uint
type Mode uint
type Register uint
type Group uint

type Engine struct {
	Handle C.csh
	Arch   Arch
	Mode   Mode
}

type InstructionHeader struct {
	Id               uint
	Address          uint
	Size             uint
	Mnemonic         string
	OpStr            string
	RegistersRead    []Register
	RegistersWritten []Register
	Groups           []Group
}

type Instruction struct {
	InstructionHeader
	Arm ArmInstruction
}

func (e Engine) Close() (bool, error) {
	res, err := C.cs_close(e.Handle)
	return bool(res), err
}

func (e Engine) Disasm(input string, offset, count uint64) ([]Instruction, error) {

	code := C.CString(input)
	defer C.free(unsafe.Pointer(code))

	var insn *C.cs_insn

	disassembled := C.cs_disasm_dyn(
		e.Handle,
		code,
		C.uint64_t(len(input)),
		C.uint64_t(offset),
		C.uint64_t(count),
		&insn,
	)
	defer C.cs_free(unsafe.Pointer(insn))

	if disassembled > 0 {
		// Create a slice, and reflect its header
		var insns []C.cs_insn
		h := (*reflect.SliceHeader)(unsafe.Pointer(&insns))
		// Manually fill in the ptr, len and cap from the raw C data
		h.Data = uintptr(unsafe.Pointer(insn))
		h.Len = int(disassembled)
		h.Cap = int(disassembled)
		switch e.Arch {
		case CS_ARCH_ARM:
			return DecomposeArm(insns), nil
		}
	}
	return nil, fmt.Errorf("Disassembly failed.")
}

func New(arch Arch, mode Mode) (Engine, error) {
	var handle C.csh
	res, err := C.cs_open(C.cs_arch(arch), C.cs_mode(mode), &handle)
	if res {
		return Engine{handle, arch, mode}, nil
	}
	return Engine{}, err
}