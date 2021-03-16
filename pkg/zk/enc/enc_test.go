package zkenc

import (
	"testing"

	"github.com/taurusgroup/cmp-ecdsa/pkg/arith"
	"github.com/taurusgroup/cmp-ecdsa/pkg/zk/zkcommon"
)

func TestEnc(t *testing.T) {
	verifier := zkcommon.Pedersen
	prover := zkcommon.ProverPaillierPublic

	k := arith.Sample(arith.L, false)
	K, rho := prover.Enc(k, nil)

	proof := NewProof(prover, verifier, K, k, rho)
	if !proof.Verify(prover, verifier, K) {
		t.Error("failed to verify")
	}
}
