package neotransaction

import (
	"bytes"
	"math/big"
	"github.com/terender/neo-go-sdk/neotransaction/OpCode"
	"github.com/terender/neo-go-sdk/neoutils"
	"utils"
)

// ScriptBuilder NEO 智能合约脚本构建器
type ScriptBuilder struct {
	buff bytes.Buffer
}

// Bytes 获取脚本构建器输出的二进制脚本数据
func (sb *ScriptBuilder) Bytes() []byte {
	return sb.buff.Bytes()
}

// Emit 在脚本构建器中加入一条不带参数的指令
func (sb *ScriptBuilder) Emit(op OpCode.OPCODE) {
	sb.buff.WriteByte(byte(op))
}

// EmitOpArgs 在脚本构建器中加入一条指令以及它的参数
func (sb *ScriptBuilder) EmitOpArgs(op OpCode.OPCODE, args []byte) {
	sb.buff.WriteByte(byte(op))
	sb.buff.Write(args)
}

// EmitAppCall 在脚本构建器中加入一条合约调用指令，参数为被调用的合约的脚本哈希
func (sb *ScriptBuilder) EmitAppCall(scriptHash neoutils.HASH160) {
	sb.Emit(OpCode.APPCALL)
	sb.buff.Write(utils.Reverse(scriptHash))
}

// EmitPushBool 在脚本构建器中加入一条压栈布尔值的指令
func (sb *ScriptBuilder) EmitPushBool(arg bool) {
	if arg {
		sb.Emit(OpCode.PUSHT)
	} else {
		sb.Emit(OpCode.PUSHF)
	}
}

// EmitPushBytes 在脚本构建器中加入一条压栈字节数组的指令
func (sb *ScriptBuilder) EmitPushBytes(arg []byte) {
	if len(arg) <= int(OpCode.PUSHBYTES75) {
		sb.buff.WriteByte(byte(len(arg)))
		sb.buff.Write(arg)
	} else if len(arg) <= 0xff {
		sb.Emit(OpCode.PUSHDATA1)
		sb.buff.WriteByte(byte(len(arg)))
		sb.buff.Write(arg)
	} else if len(arg) <= 0xffff {
		sb.Emit(OpCode.PUSHDATA2)
		utils.WriteUint16ToBuffer(&sb.buff, uint16(len(arg)))
		sb.buff.Write(arg)
	} else {
		sb.Emit(OpCode.PUSHDATA4)
		utils.WriteUint32ToBuffer(&sb.buff, uint32(len(arg)))
		sb.buff.Write(arg)
	}
}

// EmitPushNumber 在脚本构建器中加入一条压栈数字的指令
func (sb *ScriptBuilder) EmitPushNumber(arg int64) {
	if arg == -1 {
		sb.Emit(OpCode.PUSHM1)
		return
	}
	if arg == 0 {
		sb.Emit(OpCode.PUSH0)
		return
	}
	if arg > 0 && arg <= 16 {
		sb.Emit(OpCode.PUSH1 - 1 + OpCode.OPCODE(arg))
		return
	}
	bytes := utils.Reverse(big.NewInt(arg).Bytes())
	sb.EmitPushBytes(bytes)
}

// EmitPushString 在脚本构建器中加入一条压栈字符串的指令，压栈字符串实际上是压栈字节数组
func (sb *ScriptBuilder) EmitPushString(arg string) {
	sb.EmitPushBytes([]byte(arg))
}

// BuildBasicWitnessScript 创建一个基础的鉴证人脚本,包含基础压栈脚本和基础鉴权脚本
// =============基础压栈脚本============
// Push Signature
// ================================
// =============基础鉴权脚本============
// Push PublicKey
// CheckSig
// ================================
func BuildBasicWitnessScript(keyPair *KeyPair, rawTx []byte) (*Script, error) {

	script := &Script{}

	// 对原始交易进行签名
	signature, err := keyPair.Sign(neoutils.Sha256(rawTx))
	if err != nil {
		return script, err
	}

	// 创建压栈脚本
	script.InvocationScript = make([]byte, len(signature)+1)
	script.InvocationScript[0] = byte(len(signature))
	copy(script.InvocationScript[1:], signature)
	script.InvScriptLength.Value = uint64(len(script.InvocationScript))

	// 压缩公钥数据串
	pubKey := keyPair.EncodePubkeyCompressed()

	// 创建鉴权脚本
	script.VerificationScript = make([]byte, len(pubKey)+2)
	script.VerificationScript[0] = byte(len(pubKey))
	copy(script.VerificationScript[1:], pubKey)
	script.VerificationScript[1+len(pubKey)] = 0xac
	script.VrifScriptLength.Value = uint64(len(script.VerificationScript))

	return script, nil
}

// BuildBasicVerifyScript 创建基本账户鉴权脚本
func BuildBasicVerifyScript(keyPair *KeyPair) []byte {

	// 压缩公钥数据串
	pubKey := keyPair.EncodePubkeyCompressed()

	// 创建鉴权脚本
	VerificationScript := make([]byte, len(pubKey)+2)
	VerificationScript[0] = byte(len(pubKey))
	copy(VerificationScript[1:], pubKey)
	VerificationScript[1+len(pubKey)] = 0xac

	return VerificationScript
}
