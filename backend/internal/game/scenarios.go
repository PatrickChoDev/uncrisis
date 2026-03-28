// Package game contains the built-in crisis scenario bank.
package game

// DefaultScenarios is the scenario bank seeded into DynamoDB on startup.
var DefaultScenarios = []Scenario{
	{
		ScenarioID: "SC001",
		Title:      "Arctic Resource Dispute",
		Context: `Rival nations have deployed military vessels to the Arctic Circle following the
discovery of vast untapped oil reserves. Melting ice has opened new shipping
routes, escalating territorial claims. The UN Security Council has been called
to an emergency session.`,
		Options: []Option{
			{ID: "A", Label: "Propose joint development treaty with revenue sharing"},
			{ID: "B", Label: "Call for immediate demilitarisation and neutral observer deployment"},
			{ID: "C", Label: "Refer the dispute to the International Court of Justice"},
			{ID: "D", Label: "Impose targeted sanctions on the aggressor nation"},
		},
	},
	{
		ScenarioID: "SC002",
		Title:      "Refugee Crisis at the Southern Border",
		Context: `Tens of thousands of displaced civilians are massing at a heavily fortified
border after civil conflict in a neighbouring state. The host nation has closed
its crossings, citing security concerns. Aid organisations warn of imminent
humanitarian catastrophe.`,
		Options: []Option{
			{ID: "A", Label: "Establish UN-managed safe zones inside the conflict zone"},
			{ID: "B", Label: "Negotiate temporary humanitarian corridors with both governments"},
			{ID: "C", Label: "Mobilise an emergency international resettlement programme"},
			{ID: "D", Label: "Dispatch peacekeeping forces to escort aid convoys"},
		},
	},
	{
		ScenarioID: "SC003",
		Title:      "Cyber-Attack on Critical Infrastructure",
		Context: `A sophisticated cyber-attack has crippled the power grid of a NATO member,
causing hospital failures and economic panic. Attribution points to a state actor
but evidence is circumstantial. Tensions are at their highest in a generation.`,
		Options: []Option{
			{ID: "A", Label: "Activate collective-defence protocols and issue a formal ultimatum"},
			{ID: "B", Label: "Open backchannel talks to de-escalate before public attribution"},
			{ID: "C", Label: "Present evidence to the UN General Assembly and seek consensus"},
			{ID: "D", Label: "Deploy international cyber-forensics team for impartial investigation"},
		},
	},
	{
		ScenarioID: "SC004",
		Title:      "Food Security Crisis in the Sahel",
		Context: `Climate-driven drought has decimated harvests across five Sahel nations.
Grain prices have tripled; food riots are breaking out in capitals. Warlords are
blocking aid convoys in exchange for political concessions. 40 million people face
acute hunger.`,
		Options: []Option{
			{ID: "A", Label: "Establish an emergency UN food airlift bypassing road routes"},
			{ID: "B", Label: "Negotiate protected aid corridors with warlord factions"},
			{ID: "C", Label: "Release global strategic grain reserves through the WFP"},
			{ID: "D", Label: "Fast-track long-term agricultural investment and climate adaptation funds"},
		},
	},
	{
		ScenarioID: "SC005",
		Title:      "Nuclear Brinkmanship in Southeast Asia",
		Context: `Two neighbouring nuclear-armed states have exchanged artillery fire along a
disputed river boundary. Both governments have raised their alert levels to DEFCON 2.
Regional allies are pressing for a security guarantee; others urge restraint.`,
		Options: []Option{
			{ID: "A", Label: "Convene emergency P5+1 talks and propose a mutual stand-down"},
			{ID: "B", Label: "Deploy UN military observers to the disputed boundary"},
			{ID: "C", Label: "Offer binding security guarantees in exchange for nuclear stand-down"},
			{ID: "D", Label: "Invoke Chapter VII and call for immediate ceasefire under UN authority"},
		},
	},
}
