package ecpair

import (
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/projection"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
)

func (ec *ECPair) csBinaryConstraints(comp *wizard.CompiledIOP) {
	mustBeBinary(comp, ec.IsActive)
	mustBeBinary(comp, ec.UnalignedPairingData.IsPulling)
	mustBeBinary(comp, ec.UnalignedPairingData.IsComputed)
	mustBeBinary(comp, ec.UnalignedPairingData.IsFirstLineOfInstance)
	mustBeBinary(comp, ec.UnalignedPairingData.IsAccumulatorInit)
	mustBeBinary(comp, ec.UnalignedPairingData.IsAccumulatorCurr)
	mustBeBinary(comp, ec.UnalignedPairingData.IsAccumulatorPrev)
	mustBeBinary(comp, ec.UnalignedPairingData.ToMillerLoopCircuitMask)
	mustBeBinary(comp, ec.UnalignedPairingData.ToFinalExpCircuitMask)
	mustBeBinary(comp, ec.UnalignedG2MembershipData.IsPulling)
	mustBeBinary(comp, ec.UnalignedG2MembershipData.IsComputed)
	mustBeBinary(comp, ec.UnalignedG2MembershipData.ToG2MembershipCircuitMask)
}

func (ec *ECPair) csFlagConsistency(comp *wizard.CompiledIOP) {
	// flag consistency. That the assigned data is pulled from input or computed.
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_FLAG_CONSISTENCY", nameECPair),
		sym.Sub(ec.IsActive,
			ec.UnalignedPairingData.IsPulling,
			ec.UnalignedPairingData.IsComputed,
			ec.UnalignedG2MembershipData.IsPulling,
			ec.UnalignedG2MembershipData.IsComputed),
	)
}

func (ec *ECPair) csOffWhenInactive(comp *wizard.CompiledIOP) {
	// nothing is set when inactive
	isZeroWhenInactive(comp, ec.UnalignedPairingData.Limb, ec.IsActive)
	isZeroWhenInactive(comp, ec.UnalignedPairingData.ToMillerLoopCircuitMask, ec.IsActive)
	isZeroWhenInactive(comp, ec.UnalignedPairingData.ToFinalExpCircuitMask, ec.IsActive)
	isZeroWhenInactive(comp, ec.UnalignedG2MembershipData.Limb, ec.IsActive)
	isZeroWhenInactive(comp, ec.UnalignedG2MembershipData.ToG2MembershipCircuitMask, ec.IsActive)
}

func (ec *ECPair) csProjections(comp *wizard.CompiledIOP) {
	// we project data from the arithmetization correctly to the unaligned part of the circuit
	projection.InsertProjection(
		comp, ifaces.QueryIDf("%v_PROJECTION_PAIRING", nameECPair),
		[]ifaces.Column{ec.ECPairSource.Limb, ec.ECPairSource.AccPairings, ec.ECPairSource.TotalPairings, ec.ECPairSource.ID},
		[]ifaces.Column{ec.UnalignedPairingData.Limb, ec.UnalignedPairingData.PairID, ec.UnalignedPairingData.TotalPairs, ec.UnalignedPairingData.InstanceID},
		ec.ECPairSource.CsEcpairing,
		ec.UnalignedPairingData.IsPulling,
	)
	projection.InsertProjection(
		comp, ifaces.QueryIDf("%v_PROJECTION_MEMBERSHIP", nameECPair),
		[]ifaces.Column{ec.ECPairSource.Limb, ec.ECPairSource.SuccessBit},
		[]ifaces.Column{ec.UnalignedG2MembershipData.Limb, ec.UnalignedG2MembershipData.SuccessBit},
		ec.ECPairSource.CsG2Membership, ec.UnalignedG2MembershipData.IsPulling,
	)
}

func (ec *ECPair) csMembershipComputedResult(comp *wizard.CompiledIOP) {
	// when row is IS_COMPUTED, then this corresponds to the result of the G2
	// membership check. In the source it is in separate column and we have to
	// show that it corresponds to the column (but in previous row).
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_MEMBERSHIP_CHECK_RESULT", nameECPair),
		sym.Mul(ec.UnalignedG2MembershipData.IsComputed, sym.Sub(ec.UnalignedG2MembershipData.Limb, column.Shift(ec.UnalignedG2MembershipData.SuccessBit, -1))),
	)
}

func (ec *ECPair) csConstantWhenIsComputing(comp *wizard.CompiledIOP) {
	// we want to ensure that when we compute lines then PAIRING_ID and
	// TOTAL_PAIRINGS is consistent with the projected values

	// IF IS_COMPUTING AND IS_ACTIVE AND NOT FIRST_LINE => PAIRING_ID_{i} = PAIRING_ID_{i-1} AND TOTAL_PAIRINGS_{i} = TOTAL_PAIRINGS_{i-1}
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_COUNTERS_CONSISTENCY", nameECPair),
		sym.Mul(
			ec.IsActive,
			ec.UnalignedPairingData.IsComputed,
			sym.Sub(1, ec.UnalignedPairingData.IsFirstLineOfInstance),
			sym.Sub(ec.UnalignedPairingData.PairID, column.Shift(ec.UnalignedPairingData.PairID, -1)),
			sym.Sub(ec.UnalignedPairingData.TotalPairs, column.Shift(ec.UnalignedPairingData.TotalPairs, -1)),
		),
	)
}

func (ec *ECPair) csInstanceIDChangeWhenNewInstance(comp *wizard.CompiledIOP) {
	// when we are at the first line of the new instance then the instance ID
	// should change
	prevEqualCurrID, cptPrevEqualCurrID := dedicated.IsZero(
		comp,
		sym.Sub(
			ec.UnalignedPairingData.InstanceID,
			column.Shift(ec.UnalignedPairingData.InstanceID, -1),
		),
	)

	ec.CptPrevEqualCurrID = cptPrevEqualCurrID

	// IF IS_ACTIVE AND FIRST_LINE AND INSTANCE_ID != 0 => INSTANCE_ID_{i} = INSTANCE_ID_{i-1} + 1
	// And the constraint does not apply on the first row.
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_INSTANCE_ID_CHANGE", nameECPair),
		sym.Mul(
			column.Shift(verifiercol.NewConstantCol(field.One(), ec.IsActive.Size()), -1), // this "useless" line helps cancelling the constraint on the first row
			ec.IsActive,
			ec.UnalignedPairingData.IsFirstLineOfInstance,
			ec.UnalignedPairingData.InstanceID,
			prevEqualCurrID,
		),
	)
}

func (ec *ECPair) csAccumulatorInit(comp *wizard.CompiledIOP) {
	// accumulator is set to (1, 0, 0, ..., 0) with 11 zeros. But we work with two limbs per fp element.

	// we omit range checking as the limbs will be projected to the gnark
	// circuit which already performs range checking for 128 bit limbs.

	// first HI=0, LO=1
	accLimbSum := sym.Add(ec.UnalignedPairingData.Limb, sym.Sub(column.Shift(ec.UnalignedPairingData.Limb, 1), 1))
	for i := 2; i < nbGtLimbs; i++ {
		// rest HI=0, LO=0
		accLimbSum = sym.Add(accLimbSum, column.Shift(ec.UnalignedPairingData.Limb, i))
	}

	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_ACCUMULATOR_INIT", nameECPair),
		sym.Mul(
			ec.IsActive,
			ec.UnalignedPairingData.IsFirstLineOfInstance,
			accLimbSum,
		),
	)
}

func (ec *ECPair) csAccumulatorConsistency(comp *wizard.CompiledIOP) {
	// that the accumulator between pairs is consistent
	projection.InsertProjection(
		comp,
		ifaces.QueryIDf("%v_ACCUMULATOR_CONSISTENCY", nameECPair),
		[]ifaces.Column{ec.UnalignedPairingData.Limb}, []ifaces.Column{ec.UnalignedPairingData.Limb},
		ec.UnalignedPairingData.IsAccumulatorCurr, ec.UnalignedPairingData.IsAccumulatorPrev,
	)
}

func (ec *ECPair) csLastPairToFinalExp(comp *wizard.CompiledIOP) {
	// if the last pair then the final exp circuit should be active

	// IF IS_ACTIVE AND (PAIR_ID == TOTAL_PAIRS OR PAIR_ID == 0) => TO_FINAL_EXP_CIRCUIT_MASK = 1
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_LAST_PAIR_TO_FINAL_EXP", nameECPair),
		sym.Mul(
			ec.IsActive,
			sym.Sub(ec.UnalignedPairingData.PairID, ec.UnalignedPairingData.TotalPairs),
			ec.UnalignedPairingData.PairID,
			ec.UnalignedPairingData.ToFinalExpCircuitMask,
		),
	)
}

func (ec *ECPair) csIndexConsistency(comp *wizard.CompiledIOP) {
	// index switches to zero when the first line of new instance. Otherwise increases
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_INDEX_START", nameECPair),
		sym.Mul(
			ec.IsActive,
			ec.UnalignedPairingData.IsFirstLineOfInstance,
			ec.UnalignedPairingData.Index,
		),
	)
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_INDEX_INCREMENT", nameECPair),
		sym.Mul(
			ec.IsActive,
			sym.Sub(1, ec.UnalignedG2MembershipData.IsPulling, ec.UnalignedG2MembershipData.IsComputed), // we dont use index in the G2 membership check
			sym.Sub(1, ec.UnalignedPairingData.IsFirstLineOfInstance),
			sym.Sub(ec.UnalignedPairingData.Index, column.Shift(ec.UnalignedPairingData.Index, -1), 1),
		),
	)
}

func (ec *ECPair) csAccumulatorMask(comp *wizard.CompiledIOP) {
	// accumulator sum is IS_COMPUTED
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_ACCUMULATOR_MASK", nameECPair),
		sym.Sub(
			ec.UnalignedPairingData.IsComputed,
			ec.UnalignedPairingData.IsAccumulatorCurr,
			ec.UnalignedPairingData.IsAccumulatorPrev,
			ec.UnalignedPairingData.IsAccumulatorInit,
		),
	)

	// first prev accumulator is 1 when pairID*60 == index
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_FIRST_ACC_PREV", nameECPair),
		sym.Mul(
			ec.IsActive,
			ec.UnalignedPairingData.IsFirstLineOfPrevAccumulator,
			sym.Sub(
				sym.Mul(nbG1Limbs+nbG2Limbs+2*nbGtLimbs, sym.Sub(ec.UnalignedPairingData.PairID, 1)),
				ec.UnalignedPairingData.Index,
			),
		),
	)

	// first curr accumulator is 1 when pairID*60-24 = index
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_FIRST_ACC_CURR", nameECPair),
		sym.Mul(
			ec.IsActive,
			ec.UnalignedPairingData.IsFirstLineOfCurrAccumulator,
			sym.Sub(
				sym.Mul(nbG1Limbs+nbG2Limbs+2*nbGtLimbs, ec.UnalignedPairingData.PairID),
				ec.UnalignedPairingData.Index,
				24,
			),
		),
	)

	sumMask := func(col ifaces.Column) *sym.Expression {
		r := sym.NewConstant(0)
		for i := 0; i < nbGtLimbs; i++ {
			r = sym.Add(r, column.Shift(col, i))
		}
		return r
	}
	// init accumulator mask is 1 when at the start of the instance
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_INIT_ACC_MASK", nameECPair),
		sym.Mul(
			ec.IsActive,
			ec.UnalignedPairingData.IsFirstLineOfInstance,
			sym.Sub(nbGtLimbs, sumMask(ec.UnalignedPairingData.IsAccumulatorInit)),
		),
	)

	// curr accumulator mask is 1 when at the end of the pair
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_CURR_ACC_MASK", nameECPair),
		sym.Mul(
			ec.IsActive,
			ec.UnalignedPairingData.IsFirstLineOfCurrAccumulator,
			sym.Sub(nbGtLimbs, sumMask(ec.UnalignedPairingData.IsAccumulatorCurr)),
		),
	)

	// prev accumulator mask is 1 when at the start of the pair
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_PREV_ACC_MASK", nameECPair),
		sym.Mul(
			ec.IsActive,
			ec.UnalignedPairingData.IsFirstLineOfPrevAccumulator,
			sym.Sub(nbGtLimbs, sumMask(ec.UnalignedPairingData.IsAccumulatorPrev)),
		),
	)
}

// -- utils. Copied from prover/zkevm/prover/statemanager/statesummary/state_summary.go

// isZeroWhenInactive constraints the column to cancel when inactive.
func isZeroWhenInactive(comp *wizard.CompiledIOP, c, isActive ifaces.Column) {
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_IS_ZERO_WHEN_INACTIVE", c.GetColID()),
		sym.Sub(c, sym.Mul(c, isActive)),
	)
}

// mustBeBinary constrains the current column to be binary.
func mustBeBinary(comp *wizard.CompiledIOP, c ifaces.Column) {
	comp.InsertGlobal(
		roundNr,
		ifaces.QueryIDf("%v_MUST_BE_BINARY", c.GetColID()),
		sym.Mul(c, sym.Sub(c, 1)),
	)
}
