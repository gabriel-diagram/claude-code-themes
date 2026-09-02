package pet

// The pet's names, from the design canvas "Tema Terminal Claude CLI".
//
// The English keys are the ids: they are what pet.json has stored since the
// Python, what the hooks and the evolution tree are written in, and renaming
// them would rewrite every life file in the wild. The Spanish is what a person
// reads, and it is the canvas's own wording - artboard 06 for the tree and 07
// for the lineage the panel prints as "metódico › pauta › refactor".
//
// A form with no entry here falls back to its id, so the two never fight: a
// missing name is a name that has not been chosen yet, not a crash.

// Names is the 27 forms plus the three temperaments the lineage starts from.
var Names = map[string]string{
	// the larva
	"spark": "chispa",

	// the three temperaments, as forms and as the counters the lineage shows
	"pattern": "pauta", "probe": "sonda", "ember": "brasa",
	"methodical": "metódico", "inquisitive": "inquisitivo", "impulsive": "impulsivo",

	// the seven trades
	"refactor": "refactor", "tidy": "pulcro",
	"bughunter": "cazabugs", "architect": "arquitecto",
	"sprinter": "velocista", "marathon": "maratón", "feral": "salvaje",

	// the fourteen marks
	"surgeon": "cirujano", "weaver": "tejedor",
	"monk": "monje", "gardener": "jardinero",
	"bloodhound": "sabueso", "exterminator": "exterminador",
	"cartographer": "cartógrafo", "oracle": "oráculo",
	"bolt": "relámpago", "sniper": "francotirador",
	"ox": "buey", "mole": "topo",
	"gremlin": "gremlin", "kraken": "kraken",

	// the two secrets
	"phoenix": "fénix", "chimera": "quimera",

	// the seven states of the vitality layer, artboard 02. The ids stay
	// English because five places compare against them and one of them is
	// written into the session file on disk; only the reading changes.
	"fresh": "fresca", "lively": "vibrante", "easy": "a gusto",
	"sluggish": "espesa", "tired": "cansada", "drowning": "ahogada",
	"k.o.": "k.o.",
}

// Name is what a person reads for a form or a counter.
func Name(id string) string {
	if name, ok := Names[id]; ok {
		return name
	}
	return id
}
