package jets

// JetInfo describes a Simplicity jet function
type JetInfo struct {
	GoName         string   // Go function name (e.g., "BIP340Verify")
	SimplicityName string   // Simplicity jet name (e.g., "bip_0340_verify")
	ParamTypes     []string // Parameter types
	ReturnType     string   // Return type
}

// JetRegistry holds all known jet mappings
type JetRegistry struct {
	jets map[string]JetInfo
}

// NewRegistry creates a new jet registry with all known jets
func NewRegistry() *JetRegistry {
	r := &JetRegistry{
		jets: make(map[string]JetInfo),
	}
	r.registerBuiltinJets()
	return r
}

// registerBuiltinJets adds all standard Simplicity jets
func (r *JetRegistry) registerBuiltinJets() {
	// Signature verification
	r.jets["BIP340Verify"] = JetInfo{
		GoName:         "BIP340Verify",
		SimplicityName: "bip_0340_verify",
		ParamTypes:     []string{"u256", "u256", "[u8; 64]"}, // pubkey, msg, sig
		ReturnType:     "()",
	}

	// Transaction introspection
	r.jets["SigAllHash"] = JetInfo{
		GoName:         "SigAllHash",
		SimplicityName: "sig_all_hash",
		ParamTypes:     []string{},
		ReturnType:     "u256",
	}

	// SHA-256 operations
	r.jets["SHA256Init"] = JetInfo{
		GoName:         "SHA256Init",
		SimplicityName: "sha_256_ctx_8_init",
		ParamTypes:     []string{},
		ReturnType:     "Ctx8",
	}

	r.jets["SHA256Add32"] = JetInfo{
		GoName:         "SHA256Add32",
		SimplicityName: "sha_256_ctx_8_add_32",
		ParamTypes:     []string{"Ctx8", "[u8; 32]"},
		ReturnType:     "Ctx8",
	}

	r.jets["SHA256Finalize"] = JetInfo{
		GoName:         "SHA256Finalize",
		SimplicityName: "sha_256_ctx_8_finalize",
		ParamTypes:     []string{"Ctx8"},
		ReturnType:     "u256",
	}

	r.jets["CheckLockHeight"] = JetInfo{
		GoName:         "CheckLockHeight",
		SimplicityName: "check_lock_height",
		ParamTypes:     []string{"u32"},
		ReturnType:     "()",
	}

	// Comparison operations
	r.jets["Eq256"] = JetInfo{
		GoName:         "Eq256",
		SimplicityName: "eq_256",
		ParamTypes:     []string{"u256", "u256"},
		ReturnType:     "bool",
	}

	r.jets["Eq32"] = JetInfo{
		GoName:         "Eq32",
		SimplicityName: "eq_32",
		ParamTypes:     []string{"u32", "u32"},
		ReturnType:     "bool",
	}

	r.jets["Le32"] = JetInfo{
		GoName:         "Le32",
		SimplicityName: "le_32",
		ParamTypes:     []string{"u32", "u32"},
		ReturnType:     "bool",
	}

	// Additional useful jets
	r.jets["Verify"] = JetInfo{
		GoName:         "Verify",
		SimplicityName: "verify",
		ParamTypes:     []string{"bool"},
		ReturnType:     "()",
	}

	r.jets["CurrentIndex"] = JetInfo{
		GoName:         "CurrentIndex",
		SimplicityName: "current_index",
		ParamTypes:     []string{},
		ReturnType:     "u32",
	}

	r.jets["CurrentPrevOutpoint"] = JetInfo{
		GoName:         "CurrentPrevOutpoint",
		SimplicityName: "current_prev_outpoint",
		ParamTypes:     []string{},
		ReturnType:     "(u256, u32)",
	}

	r.jets["CurrentScriptHash"] = JetInfo{
		GoName:         "CurrentScriptHash",
		SimplicityName: "current_script_hash",
		ParamTypes:     []string{},
		ReturnType:     "u256",
	}

	r.jets["LockTime"] = JetInfo{
		GoName:         "LockTime",
		SimplicityName: "lock_time",
		ParamTypes:     []string{},
		ReturnType:     "u32",
	}

	// -------------------------------------------------------------------------
	// Arithmetic jets
	// add_N: (uN, uN) -> (bool, uN)  — bool is the carry flag
	// subtract_N: (uN, uN) -> (bool, uN) — bool is the borrow flag
	// multiply_N: (uN, uN) -> u(2N)  — result is double-width (no overflow)
	// divide_N / modulo_N: (uN, uN) -> uN — quotient / remainder
	// -------------------------------------------------------------------------

	// 8-bit arithmetic
	r.jets["Add8"] = JetInfo{GoName: "Add8", SimplicityName: "add_8", ParamTypes: []string{"u8", "u8"}, ReturnType: "(bool, u8)"}
	r.jets["Subtract8"] = JetInfo{GoName: "Subtract8", SimplicityName: "subtract_8", ParamTypes: []string{"u8", "u8"}, ReturnType: "(bool, u8)"}
	r.jets["Multiply8"] = JetInfo{GoName: "Multiply8", SimplicityName: "multiply_8", ParamTypes: []string{"u8", "u8"}, ReturnType: "u16"}

	// 16-bit arithmetic
	r.jets["Add16"] = JetInfo{GoName: "Add16", SimplicityName: "add_16", ParamTypes: []string{"u16", "u16"}, ReturnType: "(bool, u16)"}
	r.jets["Subtract16"] = JetInfo{GoName: "Subtract16", SimplicityName: "subtract_16", ParamTypes: []string{"u16", "u16"}, ReturnType: "(bool, u16)"}
	r.jets["Multiply16"] = JetInfo{GoName: "Multiply16", SimplicityName: "multiply_16", ParamTypes: []string{"u16", "u16"}, ReturnType: "u32"}

	// 32-bit arithmetic
	r.jets["Add32"] = JetInfo{GoName: "Add32", SimplicityName: "add_32", ParamTypes: []string{"u32", "u32"}, ReturnType: "(bool, u32)"}
	r.jets["Subtract32"] = JetInfo{GoName: "Subtract32", SimplicityName: "subtract_32", ParamTypes: []string{"u32", "u32"}, ReturnType: "(bool, u32)"}
	r.jets["Multiply32"] = JetInfo{GoName: "Multiply32", SimplicityName: "multiply_32", ParamTypes: []string{"u32", "u32"}, ReturnType: "u64"}
	r.jets["Divide32"] = JetInfo{GoName: "Divide32", SimplicityName: "divide_32", ParamTypes: []string{"u32", "u32"}, ReturnType: "u32"}
	r.jets["Modulo32"] = JetInfo{GoName: "Modulo32", SimplicityName: "modulo_32", ParamTypes: []string{"u32", "u32"}, ReturnType: "u32"}

	// 64-bit arithmetic
	r.jets["Add64"] = JetInfo{GoName: "Add64", SimplicityName: "add_64", ParamTypes: []string{"u64", "u64"}, ReturnType: "(bool, u64)"}
	r.jets["Subtract64"] = JetInfo{GoName: "Subtract64", SimplicityName: "subtract_64", ParamTypes: []string{"u64", "u64"}, ReturnType: "(bool, u64)"}
	r.jets["Multiply64"] = JetInfo{GoName: "Multiply64", SimplicityName: "multiply_64", ParamTypes: []string{"u64", "u64"}, ReturnType: "u128"}
	r.jets["Divide64"] = JetInfo{GoName: "Divide64", SimplicityName: "divide_64", ParamTypes: []string{"u64", "u64"}, ReturnType: "u64"}
	r.jets["Modulo64"] = JetInfo{GoName: "Modulo64", SimplicityName: "modulo_64", ParamTypes: []string{"u64", "u64"}, ReturnType: "u64"}

	// -------------------------------------------------------------------------
	// Comparison jets — strict and non-strict, various widths.
	// All return bool. le_32 is already registered above; others added here.
	// -------------------------------------------------------------------------

	// Strict less-than (a < b)
	r.jets["Lt8"] = JetInfo{GoName: "Lt8", SimplicityName: "lt_8", ParamTypes: []string{"u8", "u8"}, ReturnType: "bool"}
	r.jets["Lt16"] = JetInfo{GoName: "Lt16", SimplicityName: "lt_16", ParamTypes: []string{"u16", "u16"}, ReturnType: "bool"}
	r.jets["Lt32"] = JetInfo{GoName: "Lt32", SimplicityName: "lt_32", ParamTypes: []string{"u32", "u32"}, ReturnType: "bool"}
	r.jets["Lt64"] = JetInfo{GoName: "Lt64", SimplicityName: "lt_64", ParamTypes: []string{"u64", "u64"}, ReturnType: "bool"}

	// Less-than-or-equal (a <= b) — le_32 already registered; add others
	r.jets["Le8"] = JetInfo{GoName: "Le8", SimplicityName: "le_8", ParamTypes: []string{"u8", "u8"}, ReturnType: "bool"}
	r.jets["Le16"] = JetInfo{GoName: "Le16", SimplicityName: "le_16", ParamTypes: []string{"u16", "u16"}, ReturnType: "bool"}
	r.jets["Le64"] = JetInfo{GoName: "Le64", SimplicityName: "le_64", ParamTypes: []string{"u64", "u64"}, ReturnType: "bool"}

	// Equality — eq_32 and eq_256 already registered; add other widths
	r.jets["Eq8"] = JetInfo{GoName: "Eq8", SimplicityName: "eq_8", ParamTypes: []string{"u8", "u8"}, ReturnType: "bool"}
	r.jets["Eq16"] = JetInfo{GoName: "Eq16", SimplicityName: "eq_16", ParamTypes: []string{"u16", "u16"}, ReturnType: "bool"}
	r.jets["Eq64"] = JetInfo{GoName: "Eq64", SimplicityName: "eq_64", ParamTypes: []string{"u64", "u64"}, ReturnType: "bool"}

	// -------------------------------------------------------------------------
	// Bitwise logic jets — 32-bit subset (most common in contracts)
	// All return uN of the same width as input.
	// -------------------------------------------------------------------------
	r.jets["And8"] = JetInfo{GoName: "And8", SimplicityName: "and_8", ParamTypes: []string{"u8", "u8"}, ReturnType: "u8"}
	r.jets["And16"] = JetInfo{GoName: "And16", SimplicityName: "and_16", ParamTypes: []string{"u16", "u16"}, ReturnType: "u16"}
	r.jets["And32"] = JetInfo{GoName: "And32", SimplicityName: "and_32", ParamTypes: []string{"u32", "u32"}, ReturnType: "u32"}
	r.jets["And64"] = JetInfo{GoName: "And64", SimplicityName: "and_64", ParamTypes: []string{"u64", "u64"}, ReturnType: "u64"}

	r.jets["Or8"] = JetInfo{GoName: "Or8", SimplicityName: "or_8", ParamTypes: []string{"u8", "u8"}, ReturnType: "u8"}
	r.jets["Or16"] = JetInfo{GoName: "Or16", SimplicityName: "or_16", ParamTypes: []string{"u16", "u16"}, ReturnType: "u16"}
	r.jets["Or32"] = JetInfo{GoName: "Or32", SimplicityName: "or_32", ParamTypes: []string{"u32", "u32"}, ReturnType: "u32"}
	r.jets["Or64"] = JetInfo{GoName: "Or64", SimplicityName: "or_64", ParamTypes: []string{"u64", "u64"}, ReturnType: "u64"}

	r.jets["Xor8"] = JetInfo{GoName: "Xor8", SimplicityName: "xor_8", ParamTypes: []string{"u8", "u8"}, ReturnType: "u8"}
	r.jets["Xor16"] = JetInfo{GoName: "Xor16", SimplicityName: "xor_16", ParamTypes: []string{"u16", "u16"}, ReturnType: "u16"}
	r.jets["Xor32"] = JetInfo{GoName: "Xor32", SimplicityName: "xor_32", ParamTypes: []string{"u32", "u32"}, ReturnType: "u32"}
	r.jets["Xor64"] = JetInfo{GoName: "Xor64", SimplicityName: "xor_64", ParamTypes: []string{"u64", "u64"}, ReturnType: "u64"}

	r.jets["Complement8"] = JetInfo{GoName: "Complement8", SimplicityName: "complement_8", ParamTypes: []string{"u8"}, ReturnType: "u8"}
	r.jets["Complement16"] = JetInfo{GoName: "Complement16", SimplicityName: "complement_16", ParamTypes: []string{"u16"}, ReturnType: "u16"}
	r.jets["Complement32"] = JetInfo{GoName: "Complement32", SimplicityName: "complement_32", ParamTypes: []string{"u32"}, ReturnType: "u32"}
	r.jets["Complement64"] = JetInfo{GoName: "Complement64", SimplicityName: "complement_64", ParamTypes: []string{"u64"}, ReturnType: "u64"}

	// -------------------------------------------------------------------------
	// Time lock jets
	// -------------------------------------------------------------------------
	r.jets["CheckLockTime"] = JetInfo{GoName: "CheckLockTime", SimplicityName: "check_lock_time", ParamTypes: []string{"u32"}, ReturnType: "()"}
	r.jets["TxIsFinal"] = JetInfo{GoName: "TxIsFinal", SimplicityName: "tx_is_final", ParamTypes: []string{}, ReturnType: "bool"}
	r.jets["TxLockHeight"] = JetInfo{GoName: "TxLockHeight", SimplicityName: "tx_lock_height", ParamTypes: []string{}, ReturnType: "u32"}
	r.jets["TxLockTime"] = JetInfo{GoName: "TxLockTime", SimplicityName: "tx_lock_time", ParamTypes: []string{}, ReturnType: "u32"}
	r.jets["CheckLockDistance"] = JetInfo{GoName: "CheckLockDistance", SimplicityName: "check_lock_distance", ParamTypes: []string{"u16"}, ReturnType: "()"}
	r.jets["CheckLockDuration"] = JetInfo{GoName: "CheckLockDuration", SimplicityName: "check_lock_duration", ParamTypes: []string{"u16"}, ReturnType: "()"}
	r.jets["TxLockDistance"] = JetInfo{GoName: "TxLockDistance", SimplicityName: "tx_lock_distance", ParamTypes: []string{}, ReturnType: "u16"}
	r.jets["TxLockDuration"] = JetInfo{GoName: "TxLockDuration", SimplicityName: "tx_lock_duration", ParamTypes: []string{}, ReturnType: "u16"}

	// -------------------------------------------------------------------------
	// Transaction introspection jets (Bitcoin subset)
	// -------------------------------------------------------------------------
	r.jets["NumInputs"] = JetInfo{GoName: "NumInputs", SimplicityName: "num_inputs", ParamTypes: []string{}, ReturnType: "u32"}
	r.jets["NumOutputs"] = JetInfo{GoName: "NumOutputs", SimplicityName: "num_outputs", ParamTypes: []string{}, ReturnType: "u32"}
	r.jets["InputPrevOutpoint"] = JetInfo{GoName: "InputPrevOutpoint", SimplicityName: "input_prev_outpoint", ParamTypes: []string{"u32"}, ReturnType: "(u256, u32)"}
	r.jets["OutputScriptHash"] = JetInfo{GoName: "OutputScriptHash", SimplicityName: "output_script_hash", ParamTypes: []string{"u32"}, ReturnType: "u256"}
	r.jets["InputScriptHash"] = JetInfo{GoName: "InputScriptHash", SimplicityName: "input_script_hash", ParamTypes: []string{"u32"}, ReturnType: "u256"}
	r.jets["CurrentSequence"] = JetInfo{GoName: "CurrentSequence", SimplicityName: "current_sequence", ParamTypes: []string{}, ReturnType: "u32"}
	r.jets["Version"] = JetInfo{GoName: "Version", SimplicityName: "version", ParamTypes: []string{}, ReturnType: "u32"}
	r.jets["TransactionId"] = JetInfo{GoName: "TransactionId", SimplicityName: "transaction_id", ParamTypes: []string{}, ReturnType: "u256"}
	r.jets["GenesisBlockHash"] = JetInfo{GoName: "GenesisBlockHash", SimplicityName: "genesis_block_hash", ParamTypes: []string{}, ReturnType: "u256"}
	r.jets["InternalKey"] = JetInfo{GoName: "InternalKey", SimplicityName: "internal_key", ParamTypes: []string{}, ReturnType: "u256"}
	r.jets["TapleafVersion"] = JetInfo{GoName: "TapleafVersion", SimplicityName: "tapleaf_version", ParamTypes: []string{}, ReturnType: "u8"}
	r.jets["Tappath"] = JetInfo{GoName: "Tappath", SimplicityName: "tappath", ParamTypes: []string{}, ReturnType: "u256"}
	r.jets["ScriptCmr"] = JetInfo{GoName: "ScriptCmr", SimplicityName: "script_cmr", ParamTypes: []string{}, ReturnType: "u256"}

	// -------------------------------------------------------------------------
	// SHA-256 variant jets (additional byte-width add operations)
	// -------------------------------------------------------------------------
	r.jets["SHA256Add1"] = JetInfo{GoName: "SHA256Add1", SimplicityName: "sha_256_ctx_8_add_1", ParamTypes: []string{"Ctx8", "u8"}, ReturnType: "Ctx8"}
	r.jets["SHA256Add2"] = JetInfo{GoName: "SHA256Add2", SimplicityName: "sha_256_ctx_8_add_2", ParamTypes: []string{"Ctx8", "[u8; 2]"}, ReturnType: "Ctx8"}
	r.jets["SHA256Add4"] = JetInfo{GoName: "SHA256Add4", SimplicityName: "sha_256_ctx_8_add_4", ParamTypes: []string{"Ctx8", "[u8; 4]"}, ReturnType: "Ctx8"}
	r.jets["SHA256Add8"] = JetInfo{GoName: "SHA256Add8", SimplicityName: "sha_256_ctx_8_add_8", ParamTypes: []string{"Ctx8", "[u8; 8]"}, ReturnType: "Ctx8"}
	r.jets["SHA256Add16"] = JetInfo{GoName: "SHA256Add16", SimplicityName: "sha_256_ctx_8_add_16", ParamTypes: []string{"Ctx8", "[u8; 16]"}, ReturnType: "Ctx8"}
	r.jets["SHA256Add64"] = JetInfo{GoName: "SHA256Add64", SimplicityName: "sha_256_ctx_8_add_64", ParamTypes: []string{"Ctx8", "[u8; 64]"}, ReturnType: "Ctx8"}
	r.jets["SHA256Add128"] = JetInfo{GoName: "SHA256Add128", SimplicityName: "sha_256_ctx_8_add_128", ParamTypes: []string{"Ctx8", "[u8; 128]"}, ReturnType: "Ctx8"}
	r.jets["SHA256Add256"] = JetInfo{GoName: "SHA256Add256", SimplicityName: "sha_256_ctx_8_add_256", ParamTypes: []string{"Ctx8", "[u8; 256]"}, ReturnType: "Ctx8"}
	r.jets["SHA256Add512"] = JetInfo{GoName: "SHA256Add512", SimplicityName: "sha_256_ctx_8_add_512", ParamTypes: []string{"Ctx8", "[u8; 512]"}, ReturnType: "Ctx8"}

	// Low-level SHA-256 primitives
	r.jets["SHA256Block"] = JetInfo{GoName: "SHA256Block", SimplicityName: "sha_256_block", ParamTypes: []string{"u256", "[u8; 64]"}, ReturnType: "u256"}
	r.jets["SHA256IV"] = JetInfo{GoName: "SHA256IV", SimplicityName: "sha_256_iv", ParamTypes: []string{}, ReturnType: "u256"}
}

// Lookup returns the jet info for a given Go function name
func (r *JetRegistry) Lookup(goName string) (JetInfo, bool) {
	info, ok := r.jets[goName]
	return info, ok
}

// IsJet checks if a function name is a known jet
func (r *JetRegistry) IsJet(goName string) bool {
	_, ok := r.jets[goName]
	return ok
}

// GetSimplicityName returns the Simplicity jet name for a Go function name
func (r *JetRegistry) GetSimplicityName(goName string) string {
	if info, ok := r.jets[goName]; ok {
		return info.SimplicityName
	}
	return ""
}

// AllJets returns all registered jets
func (r *JetRegistry) AllJets() map[string]JetInfo {
	return r.jets
}
