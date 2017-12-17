package parser

import "gopheros/device/acpi/aml/entity"

const (
	badOpcode   = 0xff
	extOpPrefix = 0x5b
)

// objType represents the object types that are supported by the AML parser.
type objType uint8

// The list of AML object types.
const (
	objTypeAny objType = iota
	objTypeInteger
	objTypeString
	objTypeBuffer
	objTypePackage
	objTypeDevice
	objTypeEvent
	objTypeMethod
	objTypeMutex
	objTypeRegion
	objTypePower
	objTypeProcessor
	objTypeThermal
	objTypeBufferField
	objTypeLocalRegionField
	objTypeLocalBankField
	objTypeLocalReference
	objTypeLocalAlias
	objTypeLocalScope
	objTypeLocalVariable
	objTypeMethodArgument
)

// opFlag specifies a list of OR-able flags that describe the object
// type/attributes generated by a particular opcode.
type opFlag uint16

const (
	opFlagNone opFlag = 1 << iota
	opFlagHasPkgLen
	opFlagNamed
	opFlagConstant
	opFlagReference
	opFlagArithmetic
	opFlagCreate
	opFlagReturn
	opFlagExecutable
	opFlagNoOp
	opFlagScoped
)

// is returns true if f is set in this opFlag.
func (fl opFlag) is(f opFlag) bool {
	return (fl & f) != 0
}

// opArgFlags encodes up to 7 opArgFlag values in a uint64 value.
type opArgFlags uint64

// argCount returns the number of encoded args in the given flag.
func (fl opArgFlags) argCount() (count uint8) {
	// Each argument is specified using 8 bits with 0x0 indicating the end of the
	// argument list
	for ; fl&0xf != 0; fl, count = fl>>8, count+1 {
	}

	return count
}

// arg returns the arg flags for argument "num" where num is the 0-based index
// of the argument to return. The allowed values for num are 0-6.
func (fl opArgFlags) arg(num uint8) opArgFlag {
	return opArgFlag((fl >> (num * 8)) & 0xf)
}

// contains returns true if the arg flags contain any argument with type x.
func (fl opArgFlags) contains(x opArgFlag) bool {
	// Each argument is specified using 8 bits with 0x0 indicating the end of the
	// argument list
	for ; fl&0xf != 0; fl >>= 8 {
		if opArgFlag(fl&0xf) == x {
			return true
		}
	}

	return false
}

// opArgFlag represents the type of an argument expected by a particular opcode.
type opArgFlag uint8

// The list of supported opArgFlag values.
const (
	_ opArgFlag = iota
	opArgTermList
	opArgTermObj
	opArgByteList
	opArgPackage
	opArgString
	opArgByteData
	opArgWord
	opArgDword
	opArgQword
	opArgNameString
	opArgSuperName
	opArgSimpleName
	opArgDataRefObj
	opArgTarget
	opArgFieldList
)

func makeArg0() opArgFlags                     { return 0 }
func makeArg1(arg0 opArgFlag) opArgFlags       { return opArgFlags(arg0) }
func makeArg2(arg0, arg1 opArgFlag) opArgFlags { return opArgFlags(arg1)<<8 | opArgFlags(arg0) }
func makeArg3(arg0, arg1, arg2 opArgFlag) opArgFlags {
	return opArgFlags(arg2)<<16 | opArgFlags(arg1)<<8 | opArgFlags(arg0)
}
func makeArg4(arg0, arg1, arg2, arg3 opArgFlag) opArgFlags {
	return opArgFlags(arg3)<<24 | opArgFlags(arg2)<<16 | opArgFlags(arg1)<<8 | opArgFlags(arg0)
}
func makeArg5(arg0, arg1, arg2, arg3, arg4 opArgFlag) opArgFlags {
	return opArgFlags(arg4)<<32 | opArgFlags(arg3)<<24 | opArgFlags(arg2)<<16 | opArgFlags(arg1)<<8 | opArgFlags(arg0)
}
func makeArg6(arg0, arg1, arg2, arg3, arg4, arg5 opArgFlag) opArgFlags {
	return opArgFlags(arg5)<<40 | opArgFlags(arg4)<<32 | opArgFlags(arg3)<<24 | opArgFlags(arg2)<<16 | opArgFlags(arg1)<<8 | opArgFlags(arg0)
}
func makeArg7(arg0, arg1, arg2, arg3, arg4, arg5, arg6 opArgFlag) opArgFlags {
	return opArgFlags(arg6)<<48 | opArgFlags(arg5)<<40 | opArgFlags(arg4)<<32 | opArgFlags(arg3)<<24 | opArgFlags(arg2)<<16 | opArgFlags(arg1)<<8 | opArgFlags(arg0)
}

// opcodeInfo contains all known information about an opcode,
// its argument count and types as well as the type of object
// represented by it.
type opcodeInfo struct {
	op      entity.AMLOpcode
	objType objType

	flags    opFlag
	argFlags opArgFlags
}

// The opcode table contains all opcode-related information that the parser knows.
// This table is modeled after a similar table used in the acpica implementation.
var opcodeTable = []opcodeInfo{
	/*0x00*/ {entity.OpZero, objTypeInteger, opFlagConstant, makeArg0()},
	/*0x01*/ {entity.OpOne, objTypeInteger, opFlagConstant, makeArg0()},
	/*0x02*/ {entity.OpAlias, objTypeLocalAlias, opFlagNamed, makeArg2(opArgNameString, opArgNameString)},
	/*0x03*/ {entity.OpName, objTypeAny, opFlagNamed, makeArg2(opArgNameString, opArgDataRefObj)},
	/*0x04*/ {entity.OpBytePrefix, objTypeInteger, opFlagConstant, makeArg1(opArgByteData)},
	/*0x05*/ {entity.OpWordPrefix, objTypeInteger, opFlagConstant, makeArg1(opArgWord)},
	/*0x06*/ {entity.OpDwordPrefix, objTypeInteger, opFlagConstant, makeArg1(opArgDword)},
	/*0x07*/ {entity.OpStringPrefix, objTypeString, opFlagConstant, makeArg1(opArgString)},
	/*0x08*/ {entity.OpQwordPrefix, objTypeInteger, opFlagConstant, makeArg1(opArgQword)},
	/*0x09*/ {entity.OpScope, objTypeLocalScope, opFlagNamed, makeArg2(opArgNameString, opArgTermList)},
	/*0x0a*/ {entity.OpBuffer, objTypeBuffer, opFlagHasPkgLen, makeArg2(opArgTermObj, opArgByteList)},
	/*0x0b*/ {entity.OpPackage, objTypePackage, opFlagNone, makeArg2(opArgByteData, opArgTermList)},
	/*0x0c*/ {entity.OpVarPackage, objTypePackage, opFlagNone, makeArg2(opArgByteData, opArgTermList)},
	/*0x0d*/ {entity.OpMethod, objTypeMethod, opFlagNamed | opFlagScoped, makeArg3(opArgNameString, opArgByteData, opArgTermList)},
	/*0x0e*/ {entity.OpExternal, objTypeAny, opFlagNamed, makeArg3(opArgNameString, opArgByteData, opArgByteData)},
	/*0x0f*/ {entity.OpLocal0, objTypeLocalVariable, opFlagExecutable, makeArg0()},
	/*0x10*/ {entity.OpLocal1, objTypeLocalVariable, opFlagExecutable, makeArg0()},
	/*0x11*/ {entity.OpLocal2, objTypeLocalVariable, opFlagExecutable, makeArg0()},
	/*0x12*/ {entity.OpLocal3, objTypeLocalVariable, opFlagExecutable, makeArg0()},
	/*0x13*/ {entity.OpLocal4, objTypeLocalVariable, opFlagExecutable, makeArg0()},
	/*0120*/ {entity.OpLocal5, objTypeLocalVariable, opFlagExecutable, makeArg0()},
	/*0x15*/ {entity.OpLocal6, objTypeLocalVariable, opFlagExecutable, makeArg0()},
	/*0x16*/ {entity.OpLocal7, objTypeLocalVariable, opFlagExecutable, makeArg0()},
	/*0x17*/ {entity.OpArg0, objTypeMethodArgument, opFlagExecutable, makeArg0()},
	/*0x18*/ {entity.OpArg1, objTypeMethodArgument, opFlagExecutable, makeArg0()},
	/*0x19*/ {entity.OpArg2, objTypeMethodArgument, opFlagExecutable, makeArg0()},
	/*0x1a*/ {entity.OpArg3, objTypeMethodArgument, opFlagExecutable, makeArg0()},
	/*0x1b*/ {entity.OpArg4, objTypeMethodArgument, opFlagExecutable, makeArg0()},
	/*0x1c*/ {entity.OpArg5, objTypeMethodArgument, opFlagExecutable, makeArg0()},
	/*0x1d*/ {entity.OpArg6, objTypeMethodArgument, opFlagExecutable, makeArg0()},
	/*0x1e*/ {entity.OpStore, objTypeAny, opFlagExecutable, makeArg2(opArgTermObj, opArgSuperName)},
	/*0x1f*/ {entity.OpRefOf, objTypeAny, opFlagReference | opFlagExecutable, makeArg1(opArgSuperName)},
	/*0x20*/ {entity.OpAdd, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg3(opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x21*/ {entity.OpConcat, objTypeAny, opFlagExecutable, makeArg3(opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x22*/ {entity.OpSubtract, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg3(opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x23*/ {entity.OpIncrement, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg1(opArgSuperName)},
	/*0x24*/ {entity.OpDecrement, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg1(opArgSuperName)},
	/*0x25*/ {entity.OpMultiply, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg3(opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x26*/ {entity.OpDivide, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg4(opArgTermObj, opArgTermObj, opArgTarget, opArgTarget)},
	/*0x27*/ {entity.OpShiftLeft, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg3(opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x28*/ {entity.OpShiftRight, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg3(opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x29*/ {entity.OpAnd, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg3(opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x2a*/ {entity.OpNand, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg3(opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x2b*/ {entity.OpOr, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg3(opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x2c*/ {entity.OpNor, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg3(opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x2d*/ {entity.OpXor, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg3(opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x2e*/ {entity.OpNot, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg2(opArgTermObj, opArgTarget)},
	/*0x2f*/ {entity.OpFindSetLeftBit, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg2(opArgTermObj, opArgTarget)},
	/*0x30*/ {entity.OpFindSetRightBit, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg2(opArgTermObj, opArgTarget)},
	/*0x31*/ {entity.OpDerefOf, objTypeAny, opFlagExecutable, makeArg1(opArgTermObj)},
	/*0x32*/ {entity.OpConcatRes, objTypeAny, opFlagExecutable, makeArg3(opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x33*/ {entity.OpMod, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg3(opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x34*/ {entity.OpNotify, objTypeAny, opFlagExecutable, makeArg2(opArgSuperName, opArgTermObj)},
	/*0x35*/ {entity.OpSizeOf, objTypeAny, opFlagExecutable, makeArg1(opArgSuperName)},
	/*0x36*/ {entity.OpIndex, objTypeAny, opFlagExecutable, makeArg3(opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x37*/ {entity.OpMatch, objTypeAny, opFlagExecutable, makeArg6(opArgTermObj, opArgByteData, opArgTermObj, opArgByteData, opArgTermObj, opArgTermObj)},
	/*0x38*/ {entity.OpCreateDWordField, objTypeBufferField, opFlagNamed | opFlagCreate, makeArg3(opArgTermObj, opArgTermObj, opArgNameString)},
	/*0x39*/ {entity.OpCreateWordField, objTypeBufferField, opFlagNamed | opFlagCreate, makeArg3(opArgTermObj, opArgTermObj, opArgNameString)},
	/*0x3a*/ {entity.OpCreateByteField, objTypeBufferField, opFlagNamed | opFlagCreate, makeArg3(opArgTermObj, opArgTermObj, opArgNameString)},
	/*0x3b*/ {entity.OpCreateBitField, objTypeBufferField, opFlagNamed | opFlagCreate, makeArg3(opArgTermObj, opArgTermObj, opArgNameString)},
	/*0x3c*/ {entity.OpObjectType, objTypeAny, opFlagNone, makeArg1(opArgSuperName)},
	/*0x3d*/ {entity.OpCreateQWordField, objTypeBufferField, opFlagNamed | opFlagCreate, makeArg3(opArgTermObj, opArgTermObj, opArgNameString)},
	/*0x3e*/ {entity.OpLand, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg2(opArgTermObj, opArgTermObj)},
	/*0x3f*/ {entity.OpLor, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg2(opArgTermObj, opArgTermObj)},
	/*0x40*/ {entity.OpLnot, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg1(opArgTermObj)},
	/*0x41*/ {entity.OpLEqual, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg2(opArgTermObj, opArgTermObj)},
	/*0x42*/ {entity.OpLGreater, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg2(opArgTermObj, opArgTermObj)},
	/*0x43*/ {entity.OpLLess, objTypeAny, opFlagArithmetic | opFlagExecutable, makeArg2(opArgTermObj, opArgTermObj)},
	/*0x44*/ {entity.OpToBuffer, objTypeAny, opFlagExecutable, makeArg2(opArgTermObj, opArgTarget)},
	/*0x45*/ {entity.OpToDecimalString, objTypeAny, opFlagExecutable, makeArg2(opArgTermObj, opArgTarget)},
	/*0x46*/ {entity.OpToHexString, objTypeAny, opFlagExecutable, makeArg2(opArgTermObj, opArgTarget)},
	/*0x47*/ {entity.OpToInteger, objTypeAny, opFlagExecutable, makeArg2(opArgTermObj, opArgTarget)},
	/*0x48*/ {entity.OpToString, objTypeAny, opFlagExecutable, makeArg2(opArgTermObj, opArgTarget)},
	/*0x49*/ {entity.OpCopyObject, objTypeAny, opFlagExecutable, makeArg2(opArgTermObj, opArgSimpleName)},
	/*0x4a*/ {entity.OpMid, objTypeAny, opFlagExecutable, makeArg4(opArgTermObj, opArgTermObj, opArgTermObj, opArgTarget)},
	/*0x4b*/ {entity.OpContinue, objTypeAny, opFlagExecutable, makeArg0()},
	/*0x4c*/ {entity.OpIf, objTypeAny, opFlagExecutable, makeArg2(opArgTermObj, opArgTermList)},
	/*0x4d*/ {entity.OpElse, objTypeAny, opFlagExecutable | opFlagScoped, makeArg1(opArgTermList)},
	/*0x4e*/ {entity.OpWhile, objTypeAny, opFlagExecutable, makeArg2(opArgTermObj, opArgTermList)},
	/*0x4f*/ {entity.OpNoop, objTypeAny, opFlagNoOp, makeArg0()},
	/*0x50*/ {entity.OpReturn, objTypeAny, opFlagReturn, makeArg1(opArgTermObj)},
	/*0x51*/ {entity.OpBreak, objTypeAny, opFlagExecutable, makeArg0()},
	/*0x52*/ {entity.OpBreakPoint, objTypeAny, opFlagNoOp, makeArg0()},
	/*0x53*/ {entity.OpOnes, objTypeInteger, opFlagConstant, makeArg0()},
	/*0x54*/ {entity.OpMutex, objTypeMutex, opFlagNamed, makeArg2(opArgNameString, opArgByteData)},
	/*0x55*/ {entity.OpEvent, objTypeEvent, opFlagNamed, makeArg1(opArgNameString)},
	/*0x56*/ {entity.OpCondRefOf, objTypeAny, opFlagExecutable, makeArg2(opArgSuperName, opArgSuperName)},
	/*0x57*/ {entity.OpCreateField, objTypeBufferField, opFlagExecutable, makeArg4(opArgTermObj, opArgTermObj, opArgTermObj, opArgNameString)},
	/*0x58*/ {entity.OpLoadTable, objTypeAny, opFlagExecutable, makeArg7(opArgTermObj, opArgTermObj, opArgTermObj, opArgTermObj, opArgTermObj, opArgTermObj, opArgTermObj)},
	/*0x59*/ {entity.OpLoad, objTypeAny, opFlagExecutable, makeArg2(opArgNameString, opArgSuperName)},
	/*0x5a*/ {entity.OpStall, objTypeAny, opFlagExecutable, makeArg1(opArgTermObj)},
	/*0x5b*/ {entity.OpSleep, objTypeAny, opFlagExecutable, makeArg1(opArgTermObj)},
	/*0x5c*/ {entity.OpAcquire, objTypeAny, opFlagExecutable, makeArg2(opArgNameString, opArgSuperName)},
	/*0x5d*/ {entity.OpSignal, objTypeAny, opFlagExecutable, makeArg1(opArgTermObj)},
	/*0x5e*/ {entity.OpWait, objTypeAny, opFlagExecutable, makeArg2(opArgSuperName, opArgTermObj)},
	/*0x5f*/ {entity.OpReset, objTypeAny, opFlagExecutable, makeArg1(opArgSuperName)},
	/*0x60*/ {entity.OpRelease, objTypeAny, opFlagExecutable, makeArg1(opArgSuperName)},
	/*0x61*/ {entity.OpFromBCD, objTypeAny, opFlagExecutable, makeArg2(opArgTermObj, opArgTarget)},
	/*0x62*/ {entity.OpToBCD, objTypeAny, opFlagExecutable, makeArg2(opArgTermObj, opArgTarget)},
	/*0x63*/ {entity.OpUnload, objTypeAny, opFlagExecutable, makeArg1(opArgSuperName)},
	/*0x64*/ {entity.OpRevision, objTypeInteger, opFlagConstant | opFlagExecutable, makeArg0()},
	/*0x65*/ {entity.OpDebug, objTypeLocalReference, opFlagExecutable, makeArg0()},
	/*0x66*/ {entity.OpFatal, objTypeAny, opFlagExecutable, makeArg3(opArgByteData, opArgDword, opArgTermObj)},
	/*0x67*/ {entity.OpTimer, objTypeAny, opFlagNone, makeArg0()},
	/*0x68*/ {entity.OpOpRegion, objTypeRegion, opFlagNamed, makeArg4(opArgNameString, opArgByteData, opArgTermObj, opArgTermObj)},
	/*0x69*/ {entity.OpField, objTypeAny, opFlagNone, makeArg3(opArgNameString, opArgByteData, opArgFieldList)},
	/*0x6a*/ {entity.OpDevice, objTypeDevice, opFlagNamed | opFlagScoped, makeArg2(opArgNameString, opArgTermList)},
	/*0x6b*/ {entity.OpProcessor, objTypeProcessor, opFlagNamed | opFlagScoped, makeArg5(opArgNameString, opArgByteData, opArgDword, opArgByteData, opArgTermList)},
	/*0x6c*/ {entity.OpPowerRes, objTypePower, opFlagNamed | opFlagScoped, makeArg4(opArgNameString, opArgByteData, opArgWord, opArgTermList)},
	/*0x6d*/ {entity.OpThermalZone, objTypeThermal, opFlagNamed | opFlagScoped, makeArg2(opArgNameString, opArgTermList)},
	/*0x6e*/ {entity.OpIndexField, objTypeAny, opFlagNone, makeArg4(opArgNameString, opArgNameString, opArgByteData, opArgFieldList)},
	/*0x6f*/ {entity.OpBankField, objTypeLocalBankField, opFlagNamed, makeArg5(opArgNameString, opArgNameString, opArgTermObj, opArgByteData, opArgFieldList)},
	/*0x70*/ {entity.OpDataRegion, objTypeLocalRegionField, opFlagNamed, makeArg4(opArgNameString, opArgTermObj, opArgTermObj, opArgTermObj)},
}

// opcodeMap maps an AML opcode to an entry in the opcode table. Entries with
// the value 0xff indicate an invalid/unsupported opcode.
var opcodeMap = [256]uint8{
	/*              0     1     2     3     4     5     6     7*/
	/*0x00 - 0x07*/ 0x00, 0x01, 0xff, 0xff, 0xff, 0xff, 0x02, 0xff,
	/*0x08 - 0x0f*/ 0x03, 0xff, 0x04, 0x05, 0x06, 0x07, 0x08, 0xff,
	/*0x10 - 0x17*/ 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0xff, 0xff,
	/*0x18 - 0x1f*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x20 - 0x27*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x28 - 0x2f*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x30 - 0x37*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x38 - 0x3f*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x40 - 0x47*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x48 - 0x4f*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x50 - 0x57*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x58 - 0x5f*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x60 - 0x67*/ 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16,
	/*0x68 - 0x6f*/ 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0xff,
	/*0x70 - 0x77*/ 0x1e, 0x1f, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25,
	/*0x78 - 0x7f*/ 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d,
	/*0x80 - 0x87*/ 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35,
	/*0x88 - 0x8f*/ 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d,
	/*0x90 - 0x97*/ 0x3e, 0x3f, 0x40, 0x41, 0x42, 0x43, 0x44, 0x45,
	/*0x98 - 0x9f*/ 0x46, 0x47, 0x48, 0x49, 0x4a, 0x49, 0x4a, 0x4b,
	/*0xa0 - 0xa7*/ 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0xff, 0xff,
	/*0xa8 - 0xaf*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xb0 - 0xb7*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xb8 - 0xbf*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xc0 - 0xc7*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xc8 - 0xcf*/ 0xff, 0xff, 0xff, 0xff, 0x52, 0xff, 0xff, 0xff,
	/*0xd0 - 0xd7*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xd8 - 0xdf*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xe0 - 0xe7*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xe8 - 0xef*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xf0 - 0xf7*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xf8 - 0xff*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x53,
}

// extendedOpcodeMap maps an AML extended opcode (extOpPrefix + code) to an
// entry in the opcode table. Entries with the value 0xff indicate an
// invalid/unsupported opcode.
var extendedOpcodeMap = [256]uint8{
	/*              0     1     2     3     4     5     6     7*/
	/*0x00 - 0x07*/ 0xff, 0x54, 0x55, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x08 - 0x0f*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x10 - 0x17*/ 0xff, 0xff, 0x56, 0x57, 0xff, 0xff, 0xff, 0xff,
	/*0x18 - 0x1f*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x58,
	/*0x20 - 0x27*/ 0x59, 0x5a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5f, 0x60,
	/*0x28 - 0x2f*/ 0x61, 0x62, 0x63, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x30 - 0x37*/ 0x64, 0x65, 0x66, 0x67, 0xff, 0xff, 0xff, 0xff,
	/*0x38 - 0x3f*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x40 - 0x47*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x48 - 0x4f*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x50 - 0x57*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x58 - 0x5f*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x60 - 0x67*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x68 - 0x6f*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x70 - 0x77*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x78 - 0x7f*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x80 - 0x87*/ 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f,
	/*0x88 - 0x8f*/ 0x70, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x90 - 0x97*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0x98 - 0x9f*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xa0 - 0xa7*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xa8 - 0xaf*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xb0 - 0xb7*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xb8 - 0xbf*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xc0 - 0xc7*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xc8 - 0xcf*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xd0 - 0xd7*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xd8 - 0xdf*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xe0 - 0xe7*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xe8 - 0xef*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xf0 - 0xf7*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	/*0xf8 - 0xff*/ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x53,
}