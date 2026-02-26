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
		SimplicityName: "sha_256_iv",
		ParamTypes:     []string{},
		ReturnType:     "Ctx8",
	}

	r.jets["SHA256Add32"] = JetInfo{
		GoName:         "SHA256Add32",
		SimplicityName: "sha_256_block",
		ParamTypes:     []string{"Ctx8", "[u8; 32]"},
		ReturnType:     "Ctx8",
	}

	r.jets["SHA256Finalize"] = JetInfo{
		GoName:         "SHA256Finalize",
		SimplicityName: "sha_256_finalize",
		ParamTypes:     []string{"Ctx8"},
		ReturnType:     "u256",
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
