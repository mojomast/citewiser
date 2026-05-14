package packer

// QueryType describes the downstream answer/action mode for slot policy.
type QueryType string

const (
	QueryFactual     QueryType = "Factual"
	QueryComparative QueryType = "Comparative"
	QueryProcedural  QueryType = "Procedural"
	QueryExploratory QueryType = "Exploratory"
	QueryTemporal    QueryType = "Temporal"
	QueryAdversarial QueryType = "Adversarial"
	QueryAgentic     QueryType = "Agentic"
)

// SlotType identifies the reason a node was packed.
type SlotType string

const (
	SlotOverview     SlotType = "overview"
	SlotFoundation   SlotType = "foundation"
	SlotBridge       SlotType = "bridge"
	SlotCounterpoint SlotType = "counterpoint"
	SlotProcedure    SlotType = "procedure"
	SlotPermission   SlotType = "permission"
	SlotDecision     SlotType = "decision"
	SlotSupport      SlotType = "support"
)

// SlotPosition indicates the context band for lost-in-the-middle mitigation.
type SlotPosition string

const (
	PositionFront  SlotPosition = "front"
	PositionMiddle SlotPosition = "middle"
	PositionBack   SlotPosition = "back"
)

// HygieneSignal summarizes whether the packed plan is safe to use.
type HygieneSignal string

const (
	HygieneGreen  HygieneSignal = "green"
	HygieneYellow HygieneSignal = "yellow"
	HygieneRed    HygieneSignal = "red"
)
